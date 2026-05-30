package handler

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestQueryRegistry_RegisterAndCancel(t *testing.T) {
	reg := NewQueryRegistry()
	cancelled := false
	cancel := func() { cancelled = true }

	reg.Register("q1", "conn1", "mydb", "SELECT 1", cancel)

	if reg.Count() != 1 {
		t.Fatalf("expected count 1, got %d", reg.Count())
	}

	ok := reg.Cancel("q1")
	if !ok {
		t.Fatal("expected Cancel to return true for registered query")
	}
	if !cancelled {
		t.Fatal("expected cancel function to be called")
	}
}

func TestQueryRegistry_CancelNotFound(t *testing.T) {
	reg := NewQueryRegistry()
	if reg.Cancel("nonexistent") {
		t.Fatal("expected Cancel to return false for unknown query")
	}
}

func TestQueryRegistry_Unregister(t *testing.T) {
	reg := NewQueryRegistry()
	reg.Register("q1", "conn1", "db", "SELECT 1", func() {})
	reg.Unregister("q1")
	if reg.Count() != 0 {
		t.Fatalf("expected count 0 after unregister, got %d", reg.Count())
	}
	if reg.Cancel("q1") {
		t.Fatal("expected Cancel to return false after unregister")
	}
}

func TestQueryRegistry_ListActive(t *testing.T) {
	reg := NewQueryRegistry()
	reg.Register("q1", "conn1", "db1", "SELECT 1", func() {})
	reg.Register("q2", "conn2", "db2", "SELECT 2", func() {})

	active := reg.ListActive()
	if len(active) != 2 {
		t.Fatalf("expected 2 active queries, got %d", len(active))
	}
	ids := map[string]bool{}
	for _, q := range active {
		ids[q.QueryID] = true
	}
	if !ids["q1"] || !ids["q2"] {
		t.Fatalf("expected q1 and q2, got %v", ids)
	}
}

func TestQueryRegistry_GetActive(t *testing.T) {
	reg := NewQueryRegistry()
	reg.Register("q1", "conn1", "db1", "SELECT 1", func() {})

	q := reg.GetActive("q1")
	if q == nil {
		t.Fatal("expected query q1")
	}
	if q.ConnID != "conn1" {
		t.Fatalf("expected conn1, got %s", q.ConnID)
	}

	if reg.GetActive("nonexistent") != nil {
		t.Fatal("expected nil for nonexistent query")
	}
}

func TestQueryRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewQueryRegistry()
	var wg sync.WaitGroup

	// Concurrent register
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			qid := fmt.Sprintf("q%d", id)
			reg.Register(qid, "conn", "db", "SELECT 1", func() {})
		}(i)
	}
	wg.Wait()

	// Concurrent cancel + list
	for i := 0; i < 100; i++ {
		wg.Add(3)
		qid := fmt.Sprintf("q%d", i)
		go func() { defer wg.Done(); reg.Cancel(qid) }()
		go func() { defer wg.Done(); reg.ListActive() }()
		go func() { defer wg.Done(); reg.Count() }()
	}
	wg.Wait()
}

func TestQueryRegistry_CancelCallsContextCancel(t *testing.T) {
	reg := NewQueryRegistry()
	ctx, cancel := context.WithCancel(context.Background())

	reg.Register("q1", "conn1", "db", "SELECT 1", cancel)

	if ctx.Err() != nil {
		t.Fatal("context should not be cancelled yet")
	}
	reg.Cancel("q1")
	if ctx.Err() == nil {
		t.Fatal("context should be cancelled after Cancel call")
	}
}
