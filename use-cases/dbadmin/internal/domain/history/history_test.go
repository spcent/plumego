package history

import (
	"testing"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("create kv store: %v", err)
	}
	return NewStore(kv)
}

func seedEntries(t *testing.T, store *Store, connID string, entries []Entry) {
	t.Helper()
	// Add prepends, so insert in reverse order to preserve the given order
	// as "most recent first" after all Add calls complete.
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if err := store.Add(&e); err != nil {
			t.Fatalf("seed entry %q: %v", e.ID, err)
		}
	}
}

func TestStore_Search_QueryFilter(t *testing.T) {
	store := newTestStore(t)
	seedEntries(t, store, "conn1", []Entry{
		{ID: "1", ConnID: "conn1", SQL: "SELECT * FROM users", CreatedAt: time.Now().UTC()},
		{ID: "2", ConnID: "conn1", SQL: "INSERT INTO orders VALUES (1)", CreatedAt: time.Now().UTC()},
		{ID: "3", ConnID: "conn1", SQL: "select id from USERS_ARCHIVE", CreatedAt: time.Now().UTC()},
	})

	results, err := store.Search("conn1", SearchOptions{Query: "users"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2 (case-insensitive substring match): %#v", len(results), results)
	}
	for _, r := range results {
		if r.ID == "2" {
			t.Errorf("unexpected entry %q matched query %q", r.SQL, "users")
		}
	}
}

func TestStore_Search_HasErrorFilter(t *testing.T) {
	store := newTestStore(t)
	seedEntries(t, store, "conn1", []Entry{
		{ID: "1", ConnID: "conn1", SQL: "SELECT 1", CreatedAt: time.Now().UTC()},
		{ID: "2", ConnID: "conn1", SQL: "SELECT 2", Error: "syntax error", CreatedAt: time.Now().UTC()},
	})

	trueVal := true
	results, err := store.Search("conn1", SearchOptions{HasError: &trueVal})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].ID != "2" {
		t.Fatalf("HasError=true: got %#v, want only entry 2", results)
	}

	falseVal := false
	results, err = store.Search("conn1", SearchOptions{HasError: &falseVal})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].ID != "1" {
		t.Fatalf("HasError=false: got %#v, want only entry 1", results)
	}
}

func TestStore_Search_DatabaseFilter(t *testing.T) {
	store := newTestStore(t)
	seedEntries(t, store, "conn1", []Entry{
		{ID: "1", ConnID: "conn1", Database: "db_a", SQL: "SELECT 1", CreatedAt: time.Now().UTC()},
		{ID: "2", ConnID: "conn1", Database: "db_b", SQL: "SELECT 2", CreatedAt: time.Now().UTC()},
	})

	results, err := store.Search("conn1", SearchOptions{Database: "db_b"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].ID != "2" {
		t.Fatalf("Database filter: got %#v, want only entry 2", results)
	}
}

func TestStore_Search_Limit(t *testing.T) {
	store := newTestStore(t)
	seedEntries(t, store, "conn1", []Entry{
		{ID: "1", ConnID: "conn1", SQL: "SELECT 1", CreatedAt: time.Now().UTC()},
		{ID: "2", ConnID: "conn1", SQL: "SELECT 2", CreatedAt: time.Now().UTC()},
		{ID: "3", ConnID: "conn1", SQL: "SELECT 3", CreatedAt: time.Now().UTC()},
	})

	results, err := store.Search("conn1", SearchOptions{Limit: 2})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	// Most recent first: entry 1 was added last among the seeded entries.
	if results[0].ID != "1" || results[1].ID != "2" {
		t.Fatalf("unexpected order: %#v", results)
	}
}

func TestStore_Search_NoFilters_MatchesList(t *testing.T) {
	store := newTestStore(t)
	seedEntries(t, store, "conn1", []Entry{
		{ID: "1", ConnID: "conn1", SQL: "SELECT 1", CreatedAt: time.Now().UTC()},
		{ID: "2", ConnID: "conn1", SQL: "SELECT 2", CreatedAt: time.Now().UTC()},
	})

	listed, err := store.List("conn1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	searched, err := store.Search("conn1", SearchOptions{})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(listed) != len(searched) {
		t.Fatalf("len(listed)=%d, len(searched)=%d", len(listed), len(searched))
	}
	for i := range listed {
		if listed[i].ID != searched[i].ID {
			t.Errorf("order mismatch at %d: listed=%q searched=%q", i, listed[i].ID, searched[i].ID)
		}
	}
}

func TestStore_Search_EmptyConnection(t *testing.T) {
	store := newTestStore(t)
	results, err := store.Search("missing-conn", SearchOptions{Query: "anything"})
	if err != nil {
		t.Fatalf("search on missing connection: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(results))
	}
}
