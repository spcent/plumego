package redishistory

import (
	"fmt"
	"testing"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

func TestStore_AddListDeleteClear(t *testing.T) {
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("create kv store: %v", err)
	}
	store := NewStore(kv)

	entry := &Entry{
		ID:        "r1",
		ConnID:    "conn1",
		DBIndex:   0,
		Command:   "PING",
		Duration:  3,
		Success:   true,
		CreatedAt: time.Now().UTC(),
	}
	if err := store.Add(entry); err != nil {
		t.Fatalf("add entry: %v", err)
	}

	entries, err := store.List("conn1")
	if err != nil {
		t.Fatalf("list entries: %v", err)
	}
	if len(entries) != 1 || entries[0].Command != "PING" {
		t.Fatalf("unexpected entries: %#v", entries)
	}

	if err := store.Delete("conn1", "r1"); err != nil {
		t.Fatalf("delete entry: %v", err)
	}
	entries, err = store.List("conn1")
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty history after delete, got %#v", entries)
	}

	if err := store.Add(entry); err != nil {
		t.Fatalf("re-add entry: %v", err)
	}
	if err := store.Clear("conn1"); err != nil {
		t.Fatalf("clear entries: %v", err)
	}
	entries, err = store.List("conn1")
	if err != nil {
		t.Fatalf("list after clear: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty history after clear, got %#v", entries)
	}
}

func TestStore_AddCapsEntries(t *testing.T) {
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("create kv store: %v", err)
	}
	store := NewStore(kv)

	for i := 0; i < maxPerConn+5; i++ {
		if err := store.Add(&Entry{
			ID:        fmt.Sprintf("r%d", i),
			ConnID:    "conn1",
			Command:   "PING",
			Success:   true,
			CreatedAt: time.Now().UTC(),
		}); err != nil {
			t.Fatalf("add entry %d: %v", i, err)
		}
	}

	entries, err := store.List("conn1")
	if err != nil {
		t.Fatalf("list entries: %v", err)
	}
	if len(entries) != maxPerConn {
		t.Fatalf("entries len=%d, want %d", len(entries), maxPerConn)
	}
}
