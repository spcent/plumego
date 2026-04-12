package idempotency

import (
	"context"
	"testing"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

func TestKVStoreIdempotency(t *testing.T) {
	store, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("open kv: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	idem := NewKVStore(store, DefaultKVConfig())

	record := Record{
		Key:         "req-1",
		RequestHash: "hash-1",
		Status:      StatusInProgress,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	created, err := idem.PutIfAbsent(context.Background(), record)
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if !created {
		t.Fatalf("expected created")
	}

	_, found, err := idem.Get(context.Background(), "req-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !found {
		t.Fatalf("expected found")
	}

	if err := idem.Complete(context.Background(), "req-1", []byte("ok")); err != nil {
		t.Fatalf("complete: %v", err)
	}

	got, found, err := idem.Get(context.Background(), "req-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !found {
		t.Fatalf("expected found")
	}
	if got.Status != StatusCompleted {
		t.Fatalf("expected completed, got %s", got.Status)
	}
}

func TestKVStoreIdempotencyExpired(t *testing.T) {
	store, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("open kv: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	now := time.Now()
	idem := NewKVStore(store, KVConfig{Prefix: "idem:", Now: func() time.Time { return now }})

	record := Record{
		Key:       "req-expired",
		ExpiresAt: now.Add(-1 * time.Minute),
	}

	_, err = idem.PutIfAbsent(context.Background(), record)
	if err != ErrExpired {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}
