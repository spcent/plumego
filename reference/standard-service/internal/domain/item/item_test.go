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
}
