package project

import (
	"context"
	"errors"
	"testing"
)

func freePlan(_ context.Context, _ string) (string, error) { return "free", nil }

func newSvc() *Service {
	return NewService(NewMemoryStore(), Limits{"free": 2}, freePlan)
}

func TestCreateGetUpdateDelete(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	p, err := svc.Create(ctx, "tenant-a", "user-1", "Alpha", "first")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.Status != StatusActive {
		t.Fatalf("status = %q, want active", p.Status)
	}

	got, err := svc.Get(ctx, "tenant-a", p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Alpha" {
		t.Fatalf("name = %q", got.Name)
	}

	upd, err := svc.Update(ctx, "tenant-a", p.ID, "Alpha2", "", StatusPaused)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if upd.Name != "Alpha2" || upd.Status != StatusPaused || upd.Description != "first" {
		t.Fatalf("update result: %+v", upd)
	}

	if err := svc.Delete(ctx, "tenant-a", p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.Get(ctx, "tenant-a", p.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestCreateRequiresName(t *testing.T) {
	svc := newSvc()
	if _, err := svc.Create(context.Background(), "t", "u", "  ", ""); !errors.Is(err, ErrNameRequired) {
		t.Fatalf("expected ErrNameRequired, got %v", err)
	}
}

func TestUpdateInvalidStatus(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()
	p, err := svc.Create(ctx, "t", "u", "P", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.Update(ctx, "t", p.ID, "", "", Status("bogus")); !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestPlanLimitEnforced(t *testing.T) {
	svc := newSvc() // free plan limit = 2
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if _, err := svc.Create(ctx, "tenant-a", "u", "P", ""); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}
	if _, err := svc.Create(ctx, "tenant-a", "u", "P3", ""); !errors.Is(err, ErrLimitReached) {
		t.Fatalf("expected ErrLimitReached, got %v", err)
	}
	// Other tenants are unaffected.
	if _, err := svc.Create(ctx, "tenant-b", "u", "Q", ""); err != nil {
		t.Fatalf("tenant-b create: %v", err)
	}
}

func TestTenantIsolation(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	pa, err := svc.Create(ctx, "tenant-a", "u", "A", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Tenant B cannot see, update, or delete tenant A's project by ID.
	if _, err := svc.Get(ctx, "tenant-b", pa.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-tenant get: expected ErrNotFound, got %v", err)
	}
	if _, err := svc.Update(ctx, "tenant-b", pa.ID, "Stolen", "", ""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-tenant update: expected ErrNotFound, got %v", err)
	}
	if err := svc.Delete(ctx, "tenant-b", pa.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-tenant delete: expected ErrNotFound, got %v", err)
	}

	listB, err := svc.List(ctx, "tenant-b")
	if err != nil {
		t.Fatalf("list b: %v", err)
	}
	if len(listB) != 0 {
		t.Fatalf("tenant-b sees %d projects, want 0", len(listB))
	}
}

func TestUsage(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()
	if _, err := svc.Create(ctx, "t", "u", "P", ""); err != nil {
		t.Fatalf("create: %v", err)
	}
	count, limit, err := svc.Usage(ctx, "t")
	if err != nil {
		t.Fatalf("usage: %v", err)
	}
	if count != 1 || limit != 2 {
		t.Fatalf("usage = (%d, %d), want (1, 2)", count, limit)
	}
}
