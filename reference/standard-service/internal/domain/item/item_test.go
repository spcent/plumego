package item

import (
	"context"
	"math"
	"testing"
)

var ctx = context.Background()

// mustCreate is a test helper that calls Create and fails the test on error.
func mustCreate(t *testing.T, s *MemoryStore, name, description string) Item {
	t.Helper()
	item, err := s.Create(ctx, name, description)
	if err != nil {
		t.Fatalf("Create(%q, %q): %v", name, description, err)
	}
	return item
}

func TestMemoryStoreCreateAndGet(t *testing.T) {
	s := NewMemoryStore()

	got, err := s.Create(ctx, "widget", "a widget")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.ID == "" || got.Name != "widget" || got.Description != "a widget" || got.CreatedAt.IsZero() {
		t.Fatalf("Create: unexpected item: %+v", got)
	}

	found, ok := s.Get(ctx, got.ID)
	if !ok || found.ID != got.ID || found.Name != "widget" || found.Description != "a widget" {
		t.Fatalf("Get existing: unexpected result ok=%v item=%+v", ok, found)
	}

	_, ok = s.Get(ctx, "no-such-id")
	if ok {
		t.Fatal("Get missing: want false, got true")
	}
}

func TestMemoryStoreIDsAreUnique(t *testing.T) {
	s := NewMemoryStore()
	a := mustCreate(t, s, "alpha", "alpha item")
	b := mustCreate(t, s, "beta", "beta item")
	if a.ID == b.ID {
		t.Fatalf("Create: expected unique IDs, both got %q", a.ID)
	}
}

func TestMemoryStoreList(t *testing.T) {
	t.Run("empty store returns empty slice", func(t *testing.T) {
		s := NewMemoryStore()
		items, total, err := s.List(ctx, 0, math.MaxInt)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(items) != 0 || total != 0 {
			t.Fatalf("want 0 items and total=0, got len=%d total=%d", len(items), total)
		}
	})

	t.Run("populated store returns all items", func(t *testing.T) {
		s := NewMemoryStore()
		mustCreate(t, s, "alpha", "alpha item")
		mustCreate(t, s, "beta", "beta item")
		items, total, err := s.List(ctx, 0, math.MaxInt)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(items) != 2 || total != 2 {
			t.Fatalf("want len=2 total=2, got len=%d total=%d", len(items), total)
		}
	})

	t.Run("list order matches creation order", func(t *testing.T) {
		s := NewMemoryStore()
		names := []string{"first", "second", "third"}
		for _, n := range names {
			mustCreate(t, s, n, n+" item")
		}
		items, total, err := s.List(ctx, 0, math.MaxInt)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(items) != len(names) || total != len(names) {
			t.Fatalf("want len=%d total=%d, got len=%d total=%d", len(names), len(names), len(items), total)
		}
		for i, want := range names {
			if items[i].Name != want {
				t.Fatalf("item[%d].Name = %q, want %q", i, items[i].Name, want)
			}
		}
	})

	t.Run("returned slice is a copy not internal storage", func(t *testing.T) {
		s := NewMemoryStore()
		mustCreate(t, s, "alpha", "alpha item")
		items, _, _ := s.List(ctx, 0, math.MaxInt)
		// Mutate the returned slice; the store must be unaffected.
		items[0].Name = "tampered"
		fresh, _, _ := s.List(ctx, 0, math.MaxInt)
		if fresh[0].Name == "tampered" {
			t.Fatal("List returned a reference to internal storage; caller mutation should not propagate")
		}
	})

	t.Run("list excludes deleted items in stable order", func(t *testing.T) {
		s := NewMemoryStore()
		mustCreate(t, s, "alpha", "alpha item")
		b := mustCreate(t, s, "beta", "beta item")
		mustCreate(t, s, "gamma", "gamma item")
		s.Delete(ctx, b.ID)
		items, _, _ := s.List(ctx, 0, math.MaxInt)
		if len(items) != 2 {
			t.Fatalf("want 2 items after delete, got %d", len(items))
		}
		if items[0].Name != "alpha" || items[1].Name != "gamma" {
			t.Fatalf("unexpected order after delete: %v %v", items[0].Name, items[1].Name)
		}
	})
}

func TestMemoryStoreListPagination(t *testing.T) {
	t.Run("limit restricts the number of returned items", func(t *testing.T) {
		s := NewMemoryStore()
		for _, n := range []string{"a", "b", "c", "d", "e"} {
			mustCreate(t, s, n, n+" item")
		}
		items, total, err := s.List(ctx, 0, 3)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		// total reflects the full count regardless of limit.
		if total != 5 {
			t.Fatalf("total = %d, want 5", total)
		}
		if len(items) != 3 {
			t.Fatalf("len(items) = %d, want 3", len(items))
		}
		if items[0].Name != "a" || items[2].Name != "c" {
			t.Fatalf("unexpected page items: %v", items)
		}
	})

	t.Run("offset skips items in creation order", func(t *testing.T) {
		s := NewMemoryStore()
		for _, n := range []string{"a", "b", "c", "d", "e"} {
			mustCreate(t, s, n, n+" item")
		}
		items, total, err := s.List(ctx, 3, math.MaxInt)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 5 {
			t.Fatalf("total = %d, want 5", total)
		}
		if len(items) != 2 {
			t.Fatalf("len(items) = %d, want 2", len(items))
		}
		if items[0].Name != "d" || items[1].Name != "e" {
			t.Fatalf("unexpected items: %v", items)
		}
	})

	t.Run("offset beyond total returns empty page and correct total", func(t *testing.T) {
		s := NewMemoryStore()
		mustCreate(t, s, "only", "the only item")
		items, total, err := s.List(ctx, 100, 20)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 1 {
			t.Fatalf("total = %d, want 1", total)
		}
		if len(items) != 0 {
			t.Fatalf("len(items) = %d, want 0", len(items))
		}
	})

	t.Run("offset and limit together select a page", func(t *testing.T) {
		s := NewMemoryStore()
		for _, n := range []string{"a", "b", "c", "d", "e"} {
			mustCreate(t, s, n, n+" item")
		}
		items, total, err := s.List(ctx, 2, 2)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 5 {
			t.Fatalf("total = %d, want 5", total)
		}
		if len(items) != 2 {
			t.Fatalf("len(items) = %d, want 2", len(items))
		}
		if items[0].Name != "c" || items[1].Name != "d" {
			t.Fatalf("unexpected page: %v", items)
		}
	})
}

func TestMemoryStoreUpdate(t *testing.T) {
	t.Run("existing item is updated and returned", func(t *testing.T) {
		s := NewMemoryStore()
		created := mustCreate(t, s, "original", "an original item")

		updated, ok, err := s.Update(ctx, created.ID, "renamed", "updated description")
		if err != nil || !ok {
			t.Fatalf("Update: want true, nil; got ok=%v err=%v", ok, err)
		}
		// CreatedAt and ID are immutable; name and description are replaced.
		if updated.ID != created.ID || updated.Name != "renamed" ||
			updated.Description != "updated description" || updated.CreatedAt != created.CreatedAt {
			t.Fatalf("Update: unexpected result: %+v", updated)
		}

		// Verify the store reflects the change.
		got, _ := s.Get(ctx, created.ID)
		if got.Name != "renamed" || got.Description != "updated description" {
			t.Fatalf("Get after Update: name=%q desc=%q, want renamed/updated description", got.Name, got.Description)
		}
	})

	t.Run("missing id returns false and no error", func(t *testing.T) {
		s := NewMemoryStore()
		_, ok, err := s.Update(ctx, "no-such-id", "anything", "any desc")
		if err != nil {
			t.Fatalf("Update missing: want nil error, got %v", err)
		}
		if ok {
			t.Fatal("Update missing: want false, got true")
		}
	})

	t.Run("update does not change list order", func(t *testing.T) {
		s := NewMemoryStore()
		mustCreate(t, s, "alpha", "alpha item")
		b := mustCreate(t, s, "beta", "beta item")
		mustCreate(t, s, "gamma", "gamma item")
		s.Update(ctx, b.ID, "BETA", "beta updated")
		items, _, _ := s.List(ctx, 0, math.MaxInt)
		if items[1].Name != "BETA" {
			t.Fatalf("item[1].Name = %q, want BETA", items[1].Name)
		}
	})

	t.Run("description is replaced by update", func(t *testing.T) {
		s := NewMemoryStore()
		created := mustCreate(t, s, "widget", "original description")

		updated, ok, err := s.Update(ctx, created.ID, "widget", "new description")
		if err != nil || !ok {
			t.Fatalf("Update: want true, nil; got ok=%v err=%v", ok, err)
		}
		if updated.Description != "new description" {
			t.Fatalf("Update: Description = %q, want %q", updated.Description, "new description")
		}
	})
}

func TestMemoryStoreDelete(t *testing.T) {
	s := NewMemoryStore()
	it := mustCreate(t, s, "gadget", "a gadget")

	ok, err := s.Delete(ctx, it.ID)
	if err != nil || !ok {
		t.Fatalf("Delete existing: want true, nil; got ok=%v err=%v", ok, err)
	}
	if _, found := s.Get(ctx, it.ID); found {
		t.Fatal("after Delete: Get should miss, but item still found")
	}
	ok2, err2 := s.Delete(ctx, it.ID)
	if err2 != nil || ok2 {
		t.Fatalf("Delete already-deleted: want false, nil; got ok=%v err=%v", ok2, err2)
	}
}

func TestMemoryStorePatch(t *testing.T) {
	t.Run("patch name only leaves description unchanged", func(t *testing.T) {
		s := NewMemoryStore()
		created := mustCreate(t, s, "original", "original desc")

		patched, ok, err := s.Patch(ctx, created.ID, "renamed", "")
		if err != nil || !ok {
			t.Fatalf("Patch: want true, nil; got ok=%v err=%v", ok, err)
		}
		if patched.Name != "renamed" || patched.Description != "original desc" {
			t.Fatalf("Patch name only: name=%q desc=%q, want renamed/original desc", patched.Name, patched.Description)
		}
		if patched.ID != created.ID || !patched.CreatedAt.Equal(created.CreatedAt) {
			t.Fatalf("Patch: ID and CreatedAt must be immutable")
		}
	})

	t.Run("patch description only leaves name unchanged", func(t *testing.T) {
		s := NewMemoryStore()
		created := mustCreate(t, s, "original", "original desc")

		patched, ok, err := s.Patch(ctx, created.ID, "", "new desc")
		if err != nil || !ok {
			t.Fatalf("Patch: want true, nil; got ok=%v err=%v", ok, err)
		}
		if patched.Name != "original" || patched.Description != "new desc" {
			t.Fatalf("Patch desc only: name=%q desc=%q, want original/new desc", patched.Name, patched.Description)
		}
	})

	t.Run("patch both fields replaces both", func(t *testing.T) {
		s := NewMemoryStore()
		created := mustCreate(t, s, "original", "original desc")

		patched, ok, err := s.Patch(ctx, created.ID, "renamed", "new desc")
		if err != nil || !ok {
			t.Fatalf("Patch: want true, nil; got ok=%v err=%v", ok, err)
		}
		if patched.Name != "renamed" || patched.Description != "new desc" {
			t.Fatalf("Patch both: name=%q desc=%q, want renamed/new desc", patched.Name, patched.Description)
		}
	})

	t.Run("patch missing id returns false and no error", func(t *testing.T) {
		s := NewMemoryStore()
		_, ok, err := s.Patch(ctx, "no-such-id", "name", "")
		if err != nil {
			t.Fatalf("Patch missing: want nil error, got %v", err)
		}
		if ok {
			t.Fatal("Patch missing: want false, got true")
		}
	})

	t.Run("patch is reflected by subsequent get", func(t *testing.T) {
		s := NewMemoryStore()
		created := mustCreate(t, s, "original", "original desc")
		s.Patch(ctx, created.ID, "patched", "")

		got, ok := s.Get(ctx, created.ID)
		if !ok {
			t.Fatal("Get after Patch: want found")
		}
		if got.Name != "patched" || got.Description != "original desc" {
			t.Fatalf("Get after Patch: name=%q desc=%q, want patched/original desc", got.Name, got.Description)
		}
	})
}

func TestMemoryStoreContextCancellation(t *testing.T) {
	t.Run("cancelled ctx returns error from Create with no side effects", func(t *testing.T) {
		s := NewMemoryStore()
		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		got, err := s.Create(cancelled, "item", "an item")
		if err == nil {
			t.Fatal("Create with cancelled ctx: want error, got nil")
		}
		if got.ID != "" {
			t.Fatalf("Create with cancelled ctx: want zero Item, got %+v", got)
		}
		// Store must remain empty — the cancelled create must not have side effects.
		items, _, _ := s.List(ctx, 0, math.MaxInt)
		if len(items) != 0 {
			t.Fatalf("Create with cancelled ctx: store should be empty, got %d items", len(items))
		}
	})

	t.Run("cancelled ctx returns false from Get", func(t *testing.T) {
		s := NewMemoryStore()
		created := mustCreate(t, s, "item", "an item")

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		_, ok := s.Get(cancelled, created.ID)
		if ok {
			t.Fatal("Get with cancelled ctx: want false, got true")
		}
	})

	t.Run("cancelled ctx returns error from List with no side effects", func(t *testing.T) {
		s := NewMemoryStore()
		mustCreate(t, s, "item", "an item")

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		items, _, err := s.List(cancelled, 0, math.MaxInt)
		if err == nil {
			t.Fatal("List with cancelled ctx: want error, got nil")
		}
		if items != nil {
			t.Fatalf("List with cancelled ctx: want nil slice, got %v", items)
		}
	})

	t.Run("cancelled ctx returns error from Update with no side effects", func(t *testing.T) {
		s := NewMemoryStore()
		created := mustCreate(t, s, "item", "an item")

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		_, ok, err := s.Update(cancelled, created.ID, "renamed", "new desc")
		if err == nil {
			t.Fatal("Update with cancelled ctx: want error, got nil")
		}
		if ok {
			t.Fatal("Update with cancelled ctx: want false, got true")
		}
		// Original item must be unchanged.
		got, exists := s.Get(ctx, created.ID)
		if !exists || got.Name != "item" {
			t.Fatalf("Update with cancelled ctx: original item changed; got %+v", got)
		}
	})

	t.Run("cancelled ctx returns error from Patch with no side effects", func(t *testing.T) {
		s := NewMemoryStore()
		created := mustCreate(t, s, "item", "an item")

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		_, ok, err := s.Patch(cancelled, created.ID, "renamed", "")
		if err == nil {
			t.Fatal("Patch with cancelled ctx: want error, got nil")
		}
		if ok {
			t.Fatal("Patch with cancelled ctx: want false, got true")
		}
	})

	t.Run("cancelled ctx returns error from Delete with no side effects", func(t *testing.T) {
		s := NewMemoryStore()
		created := mustCreate(t, s, "item", "an item")

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		ok, err := s.Delete(cancelled, created.ID)
		if err == nil {
			t.Fatal("Delete with cancelled ctx: want error, got nil")
		}
		if ok {
			t.Fatal("Delete with cancelled ctx: want false, got true")
		}
		// Item must still exist — the cancelled delete must not have side effects.
		if _, found := s.Get(ctx, created.ID); !found {
			t.Fatal("Delete with cancelled ctx: item was deleted but should still exist")
		}
	})
}

func TestMemoryStoreDeleteDoesNotAffectOthers(t *testing.T) {
	s := NewMemoryStore()
	keep := mustCreate(t, s, "keeper", "a keeper item")
	remove := mustCreate(t, s, "removable", "a removable item")

	ok, err := s.Delete(ctx, remove.ID)
	if err != nil || !ok {
		t.Fatalf("Delete: want true, nil; got ok=%v err=%v", ok, err)
	}
	if _, found := s.Get(ctx, keep.ID); !found {
		t.Fatal("unrelated item was deleted")
	}
	items, _, _ := s.List(ctx, 0, math.MaxInt)
	if len(items) != 1 || items[0].ID != keep.ID {
		t.Fatalf("List after delete: want [keeper], got %v", items)
	}
}
