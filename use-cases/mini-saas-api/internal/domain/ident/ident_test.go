package ident

import "testing"

func TestNewUniqueAndWellFormed(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := New()
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if len(id) != 32 {
			t.Fatalf("id length = %d, want 32 hex chars", len(id))
		}
		if seen[id] {
			t.Fatalf("duplicate id generated: %s", id)
		}
		seen[id] = true
	}
}
