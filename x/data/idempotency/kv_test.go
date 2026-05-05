package idempotency

import (
	"sync/atomic"
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

	created, err := idem.PutIfAbsent(t.Context(), record)
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if !created {
		t.Fatalf("expected created")
	}

	_, found, err := idem.Get(t.Context(), "req-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !found {
		t.Fatalf("expected found")
	}

	if err := idem.Complete(t.Context(), "req-1", []byte("ok")); err != nil {
		t.Fatalf("complete: %v", err)
	}

	got, found, err := idem.Get(t.Context(), "req-1")
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

	_, err = idem.PutIfAbsent(t.Context(), record)
	if err != ErrExpired {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}

func TestKVStoreCompleteReturnsNotFoundWhenRecordExpiresAfterGet(t *testing.T) {
	store, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("open kv: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	base := time.Now()
	idem := NewKVStore(store, KVConfig{Prefix: "idem:", Now: func() time.Time { return base }})
	created, err := idem.PutIfAbsent(t.Context(), Record{
		Key:       "req-expiring",
		ExpiresAt: base.Add(time.Minute),
	})
	if err != nil || !created {
		t.Fatalf("PutIfAbsent: created=%v err=%v", created, err)
	}

	var calls atomic.Int32
	expiring := NewKVStore(store, KVConfig{Prefix: "idem:", Now: func() time.Time {
		if calls.Add(1) == 1 {
			return base
		}
		return base.Add(2 * time.Minute)
	}})

	err = expiring.Complete(t.Context(), "req-expiring", []byte("response"))
	if err != ErrNotFound {
		t.Fatalf("Complete error = %v, want ErrNotFound", err)
	}
	_, found, err := idem.Get(t.Context(), "req-expiring")
	if err != nil {
		t.Fatalf("Get after expired Complete: %v", err)
	}
	if found {
		t.Fatal("expired record should be deleted after Complete")
	}
}
