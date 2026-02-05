package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/net/mq"
	"github.com/spcent/plumego/store/idempotency"
)

func TestSMSGatewaySQLDeduper(t *testing.T) {
	if os.Getenv("SMS_GATEWAY_SQL_IT") != "1" {
		t.Skip("set SMS_GATEWAY_SQL_IT=1 to run the SQL integration test")
	}

	driver := strings.TrimSpace(os.Getenv("SMS_GATEWAY_MQ_DEDUPE_DRIVER"))
	dsn := strings.TrimSpace(os.Getenv("SMS_GATEWAY_MQ_DEDUPE_DSN"))
	if driver == "" || dsn == "" {
		t.Skip("set SMS_GATEWAY_MQ_DEDUPE_DRIVER and SMS_GATEWAY_MQ_DEDUPE_DSN to run the SQL integration test")
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

	table := strings.TrimSpace(os.Getenv("SMS_GATEWAY_MQ_DEDUPE_TABLE"))
	cfg := idempotency.DefaultSQLConfig()
	cfg.Dialect = dedupeDialectFromDriver(driver)
	if table != "" {
		cfg.Table = table
	}

	prefix := "it-dedupe"
	key := "tenant-1:task-1"
	storedKey := prefix + ":" + key
	defer cleanupIdempotencyKey(ctx, db, cfg, storedKey)

	deduper := mq.NewSQLDeduper(db, mq.SQLDeduperConfig{
		Dialect:     cfg.Dialect,
		Table:       cfg.Table,
		Prefix:      prefix,
		DefaultTTL:  200 * time.Millisecond,
		RequestHash: "mq-it",
	})

	completed, err := deduper.IsCompleted(ctx, key)
	if err != nil {
		t.Fatalf("IsCompleted: %v", err)
	}
	if completed {
		t.Fatalf("expected not completed")
	}

	if err := deduper.MarkCompleted(ctx, key, 150*time.Millisecond); err != nil {
		t.Fatalf("MarkCompleted: %v (apply docs/migrations/idempotency_*.sql)", err)
	}

	completed, err = deduper.IsCompleted(ctx, key)
	if err != nil {
		t.Fatalf("IsCompleted after mark: %v", err)
	}
	if !completed {
		t.Fatalf("expected completed")
	}

	time.Sleep(250 * time.Millisecond)

	completed, err = deduper.IsCompleted(ctx, key)
	if err != nil {
		t.Fatalf("IsCompleted after ttl: %v", err)
	}
	if completed {
		t.Fatalf("expected expired entry")
	}
}

func cleanupIdempotencyKey(ctx context.Context, db *sql.DB, cfg idempotency.SQLConfig, key string) {
	if db == nil || strings.TrimSpace(cfg.Table) == "" || strings.TrimSpace(key) == "" {
		return
	}
	query := "DELETE FROM " + cfg.Table + " WHERE key = " + idempotencyPlaceholder(1, cfg.Dialect)
	_, _ = db.ExecContext(ctx, query, key)
}

func idempotencyPlaceholder(idx int, dialect idempotency.Dialect) string {
	if dialect == idempotency.DialectPostgres {
		return fmt.Sprintf("$%d", idx)
	}
	return "?"
}
