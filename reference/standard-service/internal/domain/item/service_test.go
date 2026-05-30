package item

import (
	"context"
	"testing"
)

// TestItemServiceCreateAndGet verifies the create → get roundtrip through the
// service layer. These tests cover the service contract, not MemoryStore
// implementation details (which are tested in item_test.go).
func TestItemServiceCreateAndGet(t *testing.T) {
	svc := NewItemService(NewMemoryStore())
	ctx := context.Background()

	created, err := svc.Create(ctx, "widget", "a widget")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" || created.Name != "widget" || created.Description != "a widget" || created.CreatedAt.IsZero() {
		t.Fatalf("Create: unexpected item: %+v", created)
	}

	got, ok := svc.Get(ctx, created.ID)
	if !ok || got.ID != created.ID || got.Name != "widget" {
		t.Fatalf("Get existing: ok=%v item=%+v", ok, got)
	}

	_, ok = svc.Get(ctx, "no-such-id")
	if ok {
		t.Fatal("Get missing: want false, got true")
	}
}

func TestItemServiceList(t *testing.T) {
	svc := NewItemService(NewMemoryStore())
	ctx := context.Background()

	items, total, err := svc.List(ctx, 0, 10)
	if err != nil || total != 0 || len(items) != 0 {
		t.Fatalf("List empty store: err=%v total=%d len=%d", err, total, len(items))
	}

	if _, err := svc.Create(ctx, "alpha", "alpha item"); err != nil {
		t.Fatalf("Create alpha: %v", err)
	}
	if _, err := svc.Create(ctx, "beta", "beta item"); err != nil {
		t.Fatalf("Create beta: %v", err)
	}

	items, total, err = svc.List(ctx, 0, 10)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("List: err=%v total=%d len=%d", err, total, len(items))
	}
	if items[0].Name != "alpha" || items[1].Name != "beta" {
		t.Fatalf("List: unexpected order: [%s, %s], want [alpha, beta]", items[0].Name, items[1].Name)
	}
}

func TestItemServiceUpdate(t *testing.T) {
	svc := NewItemService(NewMemoryStore())
	ctx := context.Background()

	created, err := svc.Create(ctx, "original", "original desc")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, ok, err := svc.Update(ctx, created.ID, "renamed", "new desc")
	if err != nil || !ok {
		t.Fatalf("Update: want true nil, got ok=%v err=%v", ok, err)
	}
	if updated.Name != "renamed" || updated.Description != "new desc" {
		t.Fatalf("Update: name=%q desc=%q, want renamed/new desc", updated.Name, updated.Description)
	}
	if updated.ID != created.ID || !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Fatal("Update: ID and CreatedAt must be immutable")
	}

	_, ok, err = svc.Update(ctx, "no-such-id", "x", "y")
	if err != nil || ok {
		t.Fatalf("Update missing: want false nil, got ok=%v err=%v", ok, err)
	}
}

func TestItemServicePatch(t *testing.T) {
	svc := NewItemService(NewMemoryStore())
	ctx := context.Background()

	created, err := svc.Create(ctx, "original", "original desc")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	patched, ok, err := svc.Patch(ctx, created.ID, "renamed", "")
	if err != nil || !ok {
		t.Fatalf("Patch name only: want true nil, got ok=%v err=%v", ok, err)
	}
	if patched.Name != "renamed" || patched.Description != "original desc" {
		t.Fatalf("Patch name only: name=%q desc=%q, want renamed/original desc", patched.Name, patched.Description)
	}

	_, ok, err = svc.Patch(ctx, "no-such-id", "x", "")
	if err != nil || ok {
		t.Fatalf("Patch missing: want false nil, got ok=%v err=%v", ok, err)
	}
}

func TestItemServiceDelete(t *testing.T) {
	svc := NewItemService(NewMemoryStore())
	ctx := context.Background()

	created, err := svc.Create(ctx, "widget", "a widget")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ok, err := svc.Delete(ctx, created.ID)
	if err != nil || !ok {
		t.Fatalf("Delete existing: want true nil, got ok=%v err=%v", ok, err)
	}

	_, found := svc.Get(ctx, created.ID)
	if found {
		t.Fatal("Get after Delete: item should be gone")
	}

	ok, err = svc.Delete(ctx, "no-such-id")
	if err != nil || ok {
		t.Fatalf("Delete missing: want false nil, got ok=%v err=%v", ok, err)
	}
}
