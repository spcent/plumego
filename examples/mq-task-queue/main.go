package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/spcent/plumego/net/mq"
	mqstore "github.com/spcent/plumego/net/mq/store"
)

func main() {
	store := buildStore()
	queue := mq.NewTaskQueue(store)

	worker := mq.NewWorker(queue, mq.WorkerConfig{
		ConsumerID:  "demo-worker",
		Concurrency: 2,
	})

	worker.Register("sms.send", func(ctx context.Context, task mq.Task) error {
		log.Printf("processing task %s payload=%s", task.ID, string(task.Payload))
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Start(ctx)

	_ = queue.Enqueue(context.Background(), mq.Task{
		ID:       "task-1",
		Topic:    "sms.send",
		TenantID: "tenant-a",
		Payload:  []byte("hello"),
	}, mq.EnqueueOptions{})
	_ = queue.Enqueue(context.Background(), mq.Task{
		ID:       "task-2",
		Topic:    "sms.send",
		TenantID: "tenant-a",
		Payload:  []byte("from"),
	}, mq.EnqueueOptions{})
	_ = queue.Enqueue(context.Background(), mq.Task{
		ID:       "task-3",
		Topic:    "sms.send",
		TenantID: "tenant-a",
		Payload:  []byte("queue"),
	}, mq.EnqueueOptions{})

	time.Sleep(500 * time.Millisecond)

	if err := worker.Stop(context.Background()); err != nil {
		log.Printf("worker stop: %v", err)
	}
}

func buildStore() mq.TaskStore {
	dsn := os.Getenv("MQ_DEMO_DSN")
	driver := os.Getenv("MQ_DEMO_DRIVER")
	dialect := os.Getenv("MQ_DEMO_DIALECT")
	if dsn == "" || driver == "" || dialect == "" {
		log.Printf("MQ_DEMO_DSN/DRIVER/DIALECT not set; using in-memory store")
		return mqstore.NewMemory(mqstore.DefaultMemConfig())
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Printf("sql open failed: %v; using in-memory store", err)
		return mqstore.NewMemory(mqstore.DefaultMemConfig())
	}
	if err := db.Ping(); err != nil {
		log.Printf("sql ping failed: %v; using in-memory store", err)
		return mqstore.NewMemory(mqstore.DefaultMemConfig())
	}

	cfg := mqstore.DefaultSQLConfig()
	cfg.EnableAttemptLog = true
	cfg.AttemptLogHook = func(ctx context.Context, info mqstore.AttemptLogError) {
		log.Printf("attempt log failed op=%s task=%s attempt=%d err=%v", info.Op, info.TaskID, info.Attempt, info.Err)
	}

	switch dialect {
	case "postgres":
		cfg.Dialect = mqstore.DialectPostgres
	case "mysql":
		cfg.Dialect = mqstore.DialectMySQL
	default:
		log.Printf("unknown MQ_DEMO_DIALECT=%s; using in-memory store", dialect)
		return mqstore.NewMemory(mqstore.DefaultMemConfig())
	}

	if os.Getenv("MQ_DEMO_BREAK_ATTEMPT_LOG") == "1" {
		cfg.AttemptsTable = "mq_task_attempts_missing"
	}

	log.Printf("using SQL store (%s). Ensure migrations are applied.", dialect)
	return mqstore.NewSQL(db, cfg)
}
