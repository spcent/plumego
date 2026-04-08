package kvstore

import (
	"bytes"
	"testing"
	"time"
)

func TestKVStoreBasicOperations(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if err := store.Set("alpha", []byte("one"), 0); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := store.Get("alpha")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(got, []byte("one")) {
		t.Fatalf("unexpected value %q", got)
	}

	if !store.Exists("alpha") {
		t.Fatalf("expected alpha to exist")
	}

	if err := store.Delete("alpha"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get("alpha"); err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestKVStoreTTL(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	if err := store.Set("ttl", []byte("value"), 20*time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if store.Exists("ttl") {
		t.Fatalf("expected expired key to be gone")
	}
	if _, err := store.Get("ttl"); err != ErrKeyNotFound && err != ErrKeyExpired {
		t.Fatalf("expected expiration error, got %v", err)
	}
}

func TestKVStorePersistenceAcrossReopen(t *testing.T) {
	dir := t.TempDir()

	store, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	if err := store.Set("persist", []byte("value"), 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	reopened, err := NewKVStore(Options{DataDir: dir})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()

	got, err := reopened.Get("persist")
	if err != nil {
		t.Fatalf("Get after reopen: %v", err)
	}
	if !bytes.Equal(got, []byte("value")) {
		t.Fatalf("unexpected reopened value %q", got)
	}
}

func TestKVStoreKeysAndStats(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	defer store.Close()

	_ = store.Set("b", []byte("2"), 0)
	_ = store.Set("a", []byte("1"), 0)
	_, _ = store.Get("a")
	_, _ = store.Get("missing")

	keys := store.Keys()
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("unexpected keys: %#v", keys)
	}

	stats := store.GetStats()
	if stats.Entries != 2 {
		t.Fatalf("expected 2 entries, got %d", stats.Entries)
	}
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestKVStoreClose(t *testing.T) {
	store, err := NewKVStore(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := store.Set("closed", []byte("x"), 0); err != ErrStoreClosed {
		t.Fatalf("expected ErrStoreClosed, got %v", err)
	}
}
