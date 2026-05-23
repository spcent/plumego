package item

import (
	"context"
	"testing"
)

var ctx = context.Background()

func TestMemoryStoreCreateAndGet(t *testing.T) {
	s := NewMemoryStore()

	got := s.Create(ctx, "widget")
	if got.ID == "" || got.Name != "widget" || got.CreatedAt.IsZero() {
		t.Fatalf("Create: unexpected item: %+v", got)
	}

	found, ok := s.Get(ctx, got.ID)
	if !ok || found.ID != got.ID || found.Name != "widget" {
		t.Fatalf("Get existing: unexpected result ok=%v item=%+v", ok, found)
	}

	_, ok = s.Get(ctx, "no-such-id")
	if ok {
		t.Fatal("Get missing: want false, got true")
	}
}

func TestMemoryStoreIDsAreUnique(t *testing.T) {
	s := NewMemoryStore()
	a := s.Create(ctx, "alpha")
	b := s.Create(ctx, "beta")
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
		s.Create(ctx, "alpha")
		s.Create(ctx, "beta")
		items := s.List(ctx)
		if len(items) != 2 {
			t.Fatalf("want 2 items, got %d", len(items))
		}
	})

	t.Run("list order matches creation order", func(t *testing.T) {
		s := NewMemoryStore()
		names := []string{"first", "second", "third"}
		for _, n := range names {
			s.Create(ctx, n)
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
		s.Create(ctx, "alpha")
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
		s.Create(ctx, "alpha")
		b := s.Create(ctx, "beta")
		s.Create(ctx, "gamma")
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
		created := s.Create(ctx, "original")

		updated, ok := s.Update(ctx, created.ID, "renamed")
		if !ok {
			t.Fatal("Update: want true for existing id")
		}
		if updated.ID != created.ID || updated.Name != "renamed" || updated.CreatedAt != created.CreatedAt {
			t.Fatalf("Update: unexpected result: %+v", updated)
		}

		// Verify the store reflects the change.
		got, _ := s.Get(ctx, created.ID)
		if got.Name != "renamed" {
			t.Fatalf("Get after Update: name = %q, want %q", got.Name, "renamed")
		}
	})

	t.Run("missing id returns false", func(t *testing.T) {
		s := NewMemoryStore()
		_, ok := s.Update(ctx, "no-such-id", "anything")
		if ok {
			t.Fatal("Update missing: want false, got true")
		}
	})

	t.Run("update does not change list order", func(t *testing.T) {
		s := NewMemoryStore()
		s.Create(ctx, "alpha")
		b := s.Create(ctx, "beta")
		s.Create(ctx, "gamma")
		s.Update(ctx, b.ID, "BETA")
		items := s.List(ctx)
		if items[1].Name != "BETA" {
			t.Fatalf("item[1].Name = %q, want BETA", items[1].Name)
		}
	})
}

func TestMemoryStoreDelete(t *testing.T) {
	s := NewMemoryStore()
	it := s.Create(ctx, "gadget")

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

func TestMemoryStoreDeleteDoesNotAffectOthers(t *testing.T) {
	s := NewMemoryStore()
	keep := s.Create(ctx, "keeper")
	remove := s.Create(ctx, "removable")

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
