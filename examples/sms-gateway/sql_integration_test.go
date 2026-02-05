package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/examples/sms-gateway/internal/message"
	"github.com/spcent/plumego/examples/sms-gateway/internal/pipeline"
	"github.com/spcent/plumego/examples/sms-gateway/internal/tasks"
	"github.com/spcent/plumego/net/mq"
	mqstore "github.com/spcent/plumego/net/mq/store"
)

func TestSMSGatewaySQLPersistence(t *testing.T) {
	if os.Getenv("SMS_GATEWAY_SQL_IT") != "1" {
		t.Skip("set SMS_GATEWAY_SQL_IT=1 to run the SQL integration test")
	}

	driver := strings.TrimSpace(os.Getenv("SMS_GATEWAY_MESSAGE_DRIVER"))
	dsn := strings.TrimSpace(os.Getenv("SMS_GATEWAY_MESSAGE_DSN"))
	if driver == "" || dsn == "" {
		t.Skip("set SMS_GATEWAY_MESSAGE_DRIVER and SMS_GATEWAY_MESSAGE_DSN to run the SQL integration test")
	}

	ctx := context.Background()
	db, err := sql.Open(driver, dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	cfg := message.DefaultSQLConfig()
	cfg.Dialect = sqlDialectFromDriver(driver)
	cfg.Table = getenv("SMS_GATEWAY_MESSAGE_TABLE", cfg.Table)
	repo := message.NewSQLRepository(db, cfg)

	msgID := fmt.Sprintf("msg-it-%d", time.Now().UnixNano())
	defer cleanupSMSMessage(ctx, db, cfg.Table, cfg.Dialect, msgID)

	now := time.Now().UTC()
	msg := message.Message{
		ID:          msgID,
		TenantID:    "tenant-it",
		Status:      message.StatusAccepted,
		MaxAttempts: 2,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.Insert(ctx, msg); err != nil {
		t.Fatalf("insert message: %v (apply docs/migrations/sms_gateway_messages_*.sql)", err)
	}

	queue := mq.NewTaskQueue(mqstore.NewMemory(mqstore.MemConfig{}))
	worker := mq.NewWorker(queue, mq.WorkerConfig{
		ConsumerID:      "sms-it-worker",
		Concurrency:     1,
		PollInterval:    50 * time.Millisecond,
		LeaseDuration:   200 * time.Millisecond,
		RetryPolicy:     mq.ExponentialBackoff{Base: 20 * time.Millisecond, Max: 100 * time.Millisecond, Factor: 2},
		ShutdownTimeout: 2 * time.Second,
	})
	processor := &pipeline.Processor{
		Repo: repo,
		Providers: map[string]pipeline.ProviderSender{
			"provider-fail": &pipeline.MockProvider{Name: "provider-fail", FailFirstN: 10},
		},
	}

	worker.Register(tasks.SendTopic, processor.Handle)
	worker.Start(ctx)
	defer func() {
		_ = worker.Stop(context.Background())
	}()

	payload, err := tasks.EncodeSendTask(tasks.SendTaskPayload{
		MessageID: msgID,
		TenantID:  msg.TenantID,
		To:        "+10000000000",
		Body:      "fail",
		Provider:  "provider-fail",
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	task := mq.Task{
		ID:        fmt.Sprintf("task-it-%d", time.Now().UnixNano()),
		Topic:     tasks.SendTopic,
		TenantID:  msg.TenantID,
		Payload:   payload,
		DedupeKey: msgID,
	}
	if err := queue.Enqueue(ctx, task, mq.EnqueueOptions{MaxAttempts: 2}); err != nil {
		t.Fatalf("enqueue task: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		stored, found, err := repo.Get(ctx, msgID)
		if err != nil {
			t.Fatalf("get message: %v", err)
		}
		if found && stored.ReasonCode == message.ReasonDLQ && stored.Attempts >= 2 && stored.Status == message.StatusFailed {
			if strings.TrimSpace(stored.ReasonDetail) == "" {
				t.Fatalf("expected DLQ reason detail to be persisted")
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	stored, _, _ := repo.Get(ctx, msgID)
	t.Fatalf("timeout waiting for DLQ persistence: status=%s reason=%s attempts=%d", stored.Status, stored.ReasonCode, stored.Attempts)
}

func cleanupSMSMessage(ctx context.Context, db *sql.DB, table string, dialect message.SQLDialect, id string) {
	if db == nil || strings.TrimSpace(table) == "" || strings.TrimSpace(id) == "" {
		return
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE id = %s", table, sqlPlaceholder(1, dialect))
	_, _ = db.ExecContext(ctx, query, id)
}

func sqlPlaceholder(idx int, dialect message.SQLDialect) string {
	if dialect == message.DialectPostgres {
		return fmt.Sprintf("$%d", idx)
	}
	return "?"
}
