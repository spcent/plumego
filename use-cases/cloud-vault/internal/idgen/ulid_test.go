package idgen

import (
	"sync"
	"testing"
)

func TestNew_notEmpty(t *testing.T) {
	id := New()
	if id == "" {
		t.Error("New() returned empty string")
	}
}

func TestNew_uniqueness(t *testing.T) {
	const n = 100
	seen := make(map[string]bool, n)
	for i := 0; i < n; i++ {
		id := New()
		if seen[id] {
			t.Fatalf("New() produced duplicate ID %q at iteration %d", id, i)
		}
		seen[id] = true
	}
}

func TestNew_length(t *testing.T) {
	// ULID strings are always 26 characters.
	id := New()
	if len(id) != 26 {
		t.Errorf("New() length = %d, want 26", len(id))
	}
}

func TestNew_concurrentNoPanic(t *testing.T) {
	const goroutines = 50
	results := make([]string, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			results[i] = New()
		}()
	}
	wg.Wait()

	seen := make(map[string]bool, goroutines)
	for _, id := range results {
		if id == "" {
			t.Error("concurrent New() returned empty string")
		}
		if seen[id] {
			t.Errorf("concurrent New() produced duplicate ID %q", id)
		}
		seen[id] = true
	}
}
