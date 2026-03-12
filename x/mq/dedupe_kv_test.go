package mq

import (
	"context"
	"os"
	"testing"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

func TestKVDeduperLifecycle(t *testing.T) {
	dir, err := os.MkdirTemp("", "plumego-mq-dedupe")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: dir})
	if err != nil {
		t.Fatalf("kv store: %v", err)
	}
	defer kv.Close()

	deduper := NewKVDeduper(kv, KVDeduperConfig{
		Prefix:     "mq",
		DefaultTTL: 50 * time.Millisecond,
	})

	key := "tenant-1:task-1"
	completed, err := deduper.IsCompleted(context.Background(), key)
	if err != nil {
		t.Fatalf("IsCompleted err: %v", err)
	}
	if completed {
		t.Fatalf("expected not completed")
	}

	if err := deduper.MarkCompleted(context.Background(), key, 0); err != nil {
		t.Fatalf("MarkCompleted: %v", err)
	}

	completed, err = deduper.IsCompleted(context.Background(), key)
	if err != nil {
		t.Fatalf("IsCompleted after mark: %v", err)
	}
	if !completed {
		t.Fatalf("expected completed")
	}

	time.Sleep(80 * time.Millisecond)

	completed, err = deduper.IsCompleted(context.Background(), key)
	if err != nil {
		t.Fatalf("IsCompleted after ttl: %v", err)
	}
	if completed {
		t.Fatalf("expected expired entry")
	}
}
