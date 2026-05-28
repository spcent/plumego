package item

import (
	"context"
	"testing"
)

var ctx = context.Background()

func TestMemoryStoreCreateAndGet(t *testing.T) {
	s := NewMemoryStore()

	got := s.Create(ctx, "widget", "a widget")
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
	a := s.Create(ctx, "alpha", "alpha item")
	b := s.Create(ctx, "beta", "beta item")
	if a.ID == b.ID {
		t.Fatalf("Create: expected unique IDs, both got %q", a.ID)
	}
}

func TestMemoryStoreList(t *testing.T) {
	t.Run("empty store returns empty slice", func(t *testing.T) {
		s := NewMemoryStore()
		items := s.List(ctx)
		if len(items) != 0 {
			t.Fatalf("want 0 items, got %d", len(items))
		}
	})

	t.Run("populated store returns all items", func(t *testing.T) {
		s := NewMemoryStore()
		s.Create(ctx, "alpha", "alpha item")
		s.Create(ctx, "beta", "beta item")
		items := s.List(ctx)
		if len(items) != 2 {
			t.Fatalf("want 2 items, got %d", len(items))
		}
	})

	t.Run("list order matches creation order", func(t *testing.T) {
		s := NewMemoryStore()
		names := []string{"first", "second", "third"}
		for _, n := range names {
			s.Create(ctx, n, n+" item")
		}
		items := s.List(ctx)
		if len(items) != len(names) {
			t.Fatalf("want %d items, got %d", len(names), len(items))
		}
		for i, want := range names {
			if items[i].Name != want {
				t.Fatalf("item[%d].Name = %q, want %q", i, items[i].Name, want)
			}
		}
	})

	t.Run("returned slice is a copy not internal storage", func(t *testing.T) {
		s := NewMemoryStore()
		s.Create(ctx, "alpha", "alpha item")
		items := s.List(ctx)
		// Mutate the returned slice; the store must be unaffected.
		items[0].Name = "tampered"
		fresh := s.List(ctx)
		if fresh[0].Name == "tampered" {
			t.Fatal("List returned a reference to internal storage; caller mutation should not propagate")
		}
	})

	t.Run("list excludes deleted items in stable order", func(t *testing.T) {
		s := NewMemoryStore()
		s.Create(ctx, "alpha", "alpha item")
		b := s.Create(ctx, "beta", "beta item")
		s.Create(ctx, "gamma", "gamma item")
		s.Delete(ctx, b.ID)
		items := s.List(ctx)
		if len(items) != 2 {
			t.Fatalf("want 2 items after delete, got %d", len(items))
		}
		if items[0].Name != "alpha" || items[1].Name != "gamma" {
			t.Fatalf("unexpected order after delete: %v %v", items[0].Name, items[1].Name)
		}
	})
}

func TestMemoryStoreUpdate(t *testing.T) {
	t.Run("existing item is updated and returned", func(t *testing.T) {
		s := NewMemoryStore()
		created := s.Create(ctx, "original", "an original item")

		updated, ok := s.Update(ctx, created.ID, "renamed", "updated description")
		if !ok {
			t.Fatal("Update: want true for existing id")
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

	t.Run("missing id returns false", func(t *testing.T) {
		s := NewMemoryStore()
		_, ok := s.Update(ctx, "no-such-id", "anything", "any desc")
		if ok {
			t.Fatal("Update missing: want false, got true")
		}
	})

	t.Run("update does not change list order", func(t *testing.T) {
		s := NewMemoryStore()
		s.Create(ctx, "alpha", "alpha item")
		b := s.Create(ctx, "beta", "beta item")
		s.Create(ctx, "gamma", "gamma item")
		s.Update(ctx, b.ID, "BETA", "beta updated")
		items := s.List(ctx)
		if items[1].Name != "BETA" {
			t.Fatalf("item[1].Name = %q, want BETA", items[1].Name)
		}
	})

	t.Run("description is replaced by update", func(t *testing.T) {
		s := NewMemoryStore()
		created := s.Create(ctx, "widget", "original description")

		updated, ok := s.Update(ctx, created.ID, "widget", "new description")
		if !ok {
			t.Fatal("Update: want true for existing id")
		}
		if updated.Description != "new description" {
			t.Fatalf("Update: Description = %q, want %q", updated.Description, "new description")
		}
	})
}

func TestMemoryStoreDelete(t *testing.T) {
	s := NewMemoryStore()
	it := s.Create(ctx, "gadget", "a gadget")

	if deleted := s.Delete(ctx, it.ID); !deleted {
		t.Fatalf("Delete existing: want true, got false")
	}
	if _, ok := s.Get(ctx, it.ID); ok {
		t.Fatal("after Delete: Get should miss, but item still found")
	}
	if deleted := s.Delete(ctx, it.ID); deleted {
		t.Fatal("Delete already-deleted: want false, got true")
	}
}

func TestMemoryStorePatch(t *testing.T) {
	t.Run("patch name only leaves description unchanged", func(t *testing.T) {
		s := NewMemoryStore()
		created := s.Create(ctx, "original", "original desc")

		patched, ok := s.Patch(ctx, created.ID, "renamed", "")
		if !ok {
			t.Fatal("Patch: want true for existing id")
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
		created := s.Create(ctx, "original", "original desc")

		patched, ok := s.Patch(ctx, created.ID, "", "new desc")
		if !ok {
			t.Fatal("Patch: want true for existing id")
		}
		if patched.Name != "original" || patched.Description != "new desc" {
			t.Fatalf("Patch desc only: name=%q desc=%q, want original/new desc", patched.Name, patched.Description)
		}
	})

	t.Run("patch both fields replaces both", func(t *testing.T) {
		s := NewMemoryStore()
		created := s.Create(ctx, "original", "original desc")

		patched, ok := s.Patch(ctx, created.ID, "renamed", "new desc")
		if !ok {
			t.Fatal("Patch: want true for existing id")
		}
		if patched.Name != "renamed" || patched.Description != "new desc" {
			t.Fatalf("Patch both: name=%q desc=%q, want renamed/new desc", patched.Name, patched.Description)
		}
	})

	t.Run("patch missing id returns false", func(t *testing.T) {
		s := NewMemoryStore()
		_, ok := s.Patch(ctx, "no-such-id", "name", "")
		if ok {
			t.Fatal("Patch missing: want false, got true")
		}
	})

	t.Run("patch is reflected by subsequent get", func(t *testing.T) {
		s := NewMemoryStore()
		created := s.Create(ctx, "original", "original desc")
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
	t.Run("cancelled ctx returns false from Get", func(t *testing.T) {
		s := NewMemoryStore()
		created := s.Create(ctx, "item", "an item")

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		_, ok := s.Get(cancelled, created.ID)
		if ok {
			t.Fatal("Get with cancelled ctx: want false, got true")
		}
	})

	t.Run("cancelled ctx returns nil from List", func(t *testing.T) {
		s := NewMemoryStore()
		s.Create(ctx, "item", "an item")

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		items := s.List(cancelled)
		if items != nil {
			t.Fatalf("List with cancelled ctx: want nil, got %v", items)
		}
	})

	t.Run("cancelled ctx returns false from Patch", func(t *testing.T) {
		s := NewMemoryStore()
		created := s.Create(ctx, "item", "an item")

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		_, ok := s.Patch(cancelled, created.ID, "renamed", "")
		if ok {
			t.Fatal("Patch with cancelled ctx: want false, got true")
		}
	})
}

func TestMemoryStoreDeleteDoesNotAffectOthers(t *testing.T) {
	s := NewMemoryStore()
	keep := s.Create(ctx, "keeper", "a keeper item")
	remove := s.Create(ctx, "removable", "a removable item")

	if !s.Delete(ctx, remove.ID) {
		t.Fatal("Delete: want true")
	}
	if _, ok := s.Get(ctx, keep.ID); !ok {
		t.Fatal("unrelated item was deleted")
	}
	items := s.List(ctx)
	if len(items) != 1 || items[0].ID != keep.ID {
		t.Fatalf("List after delete: want [keeper], got %v", items)
	}
}
