package audit

import (
	"fmt"
	"testing"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

func TestStoreTrimsByMaxEventsAndRetention(t *testing.T) {
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("create kv store: %v", err)
	}
	store := NewStoreWithOptions(kv, Options{MaxEvents: 2, Retention: time.Hour})

	old := time.Now().UTC().Add(-2 * time.Hour)
	if err := store.Add(Event{ID: "old", Action: "POST /old", CreatedAt: old}); err != nil {
		t.Fatalf("add old event: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := store.Add(Event{
			ID:        fmt.Sprintf("new-%d", i),
			Action:    "POST /new",
			CreatedAt: time.Now().UTC().Add(time.Duration(i) * time.Minute),
		}); err != nil {
			t.Fatalf("add event %d: %v", i, err)
		}
	}

	events, err := store.List()
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2: %#v", len(events), events)
	}
	if events[0].ID != "new-2" || events[1].ID != "new-1" {
		t.Fatalf("unexpected retained events: %#v", events)
	}
}
