package item

import (
	"testing"
)

func TestMemoryStoreCreateAndGet(t *testing.T) {
	s := NewMemoryStore()

	got := s.Create("widget")
	if got.ID == "" || got.Name != "widget" || got.CreatedAt == "" {
		t.Fatalf("Create: unexpected item: %+v", got)
	}

	found, ok := s.Get(got.ID)
	if !ok || found.ID != got.ID || found.Name != "widget" {
		t.Fatalf("Get existing: unexpected result ok=%v item=%+v", ok, found)
	}

	_, ok = s.Get("no-such-id")
	if ok {
		t.Fatal("Get missing: want false, got true")
	}
}

func TestMemoryStoreIDsAreUnique(t *testing.T) {
	s := NewMemoryStore()
	a := s.Create("alpha")
	b := s.Create("beta")
	if a.ID == b.ID {
		t.Fatalf("Create: expected unique IDs, both got %q", a.ID)
	}
}

func TestMemoryStoreList(t *testing.T) {
	t.Run("empty store returns empty slice", func(t *testing.T) {
		s := NewMemoryStore()
		items := s.List()
		if len(items) != 0 {
			t.Fatalf("want 0 items, got %d", len(items))
		}
	})

	t.Run("populated store returns all items", func(t *testing.T) {
		s := NewMemoryStore()
		s.Create("alpha")
		s.Create("beta")
		items := s.List()
		if len(items) != 2 {
			t.Fatalf("want 2 items, got %d", len(items))
		}
	})

	t.Run("list order matches creation order", func(t *testing.T) {
		s := NewMemoryStore()
		names := []string{"first", "second", "third"}
		for _, n := range names {
			s.Create(n)
		}
		items := s.List()
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
		s.Create("alpha")
		items := s.List()
		// Mutate the returned slice; the store must be unaffected.
		items[0].Name = "tampered"
		fresh := s.List()
		if fresh[0].Name == "tampered" {
			t.Fatal("List returned a reference to internal storage; caller mutation should not propagate")
		}
	})

	t.Run("list excludes deleted items in stable order", func(t *testing.T) {
		s := NewMemoryStore()
		s.Create("alpha")
		b := s.Create("beta")
		s.Create("gamma")
		s.Delete(b.ID)
		items := s.List()
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
		created := s.Create("original")

		updated, ok := s.Update(created.ID, "renamed")
		if !ok {
			t.Fatal("Update: want true for existing id")
		}
		if updated.ID != created.ID || updated.Name != "renamed" || updated.CreatedAt != created.CreatedAt {
			t.Fatalf("Update: unexpected result: %+v", updated)
		}

		// Verify the store reflects the change.
		got, _ := s.Get(created.ID)
		if got.Name != "renamed" {
			t.Fatalf("Get after Update: name = %q, want %q", got.Name, "renamed")
		}
	})

	t.Run("missing id returns false", func(t *testing.T) {
		s := NewMemoryStore()
		_, ok := s.Update("no-such-id", "anything")
		if ok {
			t.Fatal("Update missing: want false, got true")
		}
	})

	t.Run("update does not change list order", func(t *testing.T) {
		s := NewMemoryStore()
		s.Create("alpha")
		b := s.Create("beta")
		s.Create("gamma")
		s.Update(b.ID, "BETA")
		items := s.List()
		if items[1].Name != "BETA" {
			t.Fatalf("item[1].Name = %q, want BETA", items[1].Name)
		}
	})
}

func TestMemoryStoreDelete(t *testing.T) {
	s := NewMemoryStore()
	item := s.Create("gadget")

	if deleted := s.Delete(item.ID); !deleted {
		t.Fatalf("Delete existing: want true, got false")
	}
	if _, ok := s.Get(item.ID); ok {
		t.Fatal("after Delete: Get should miss, but item still found")
	}
	if deleted := s.Delete(item.ID); deleted {
		t.Fatal("Delete already-deleted: want false, got true")
	}
}

func TestMemoryStoreDeleteDoesNotAffectOthers(t *testing.T) {
	s := NewMemoryStore()
	keep := s.Create("keeper")
	remove := s.Create("removable")

	if !s.Delete(remove.ID) {
		t.Fatal("Delete: want true")
	}
	if _, ok := s.Get(keep.ID); !ok {
		t.Fatal("unrelated item was deleted")
	}
	items := s.List()
	if len(items) != 1 || items[0].ID != keep.ID {
		t.Fatalf("List after delete: want [keeper], got %v", items)
	}
}
