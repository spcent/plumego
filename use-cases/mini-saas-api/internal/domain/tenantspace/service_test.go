package tenantspace

import (
	"context"
	"errors"
	"testing"

	"mini-saas-api/internal/domain/access"
)

func newWorkspace(t *testing.T, svc *Service) (Tenant, Membership) {
	t.Helper()
	tenant, owner, err := svc.CreateWorkspace(context.Background(), "Acme", "acme", "user-owner")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	return tenant, owner
}

func TestCreateWorkspaceSetsOwner(t *testing.T) {
	svc := NewService(NewMemoryStore())
	tenant, owner := newWorkspace(t, svc)
	if tenant.Plan != PlanFree {
		t.Fatalf("plan = %q, want %q", tenant.Plan, PlanFree)
	}
	if owner.Role != access.RoleOwner {
		t.Fatalf("first membership role = %q, want owner", owner.Role)
	}
	if owner.TenantID != tenant.ID {
		t.Fatal("membership not bound to tenant")
	}
}

func TestCreateWorkspaceDuplicateSlug(t *testing.T) {
	svc := NewService(NewMemoryStore())
	newWorkspace(t, svc)
	if _, _, err := svc.CreateWorkspace(context.Background(), "Other", "ACME", "user-2"); !errors.Is(err, ErrSlugTaken) {
		t.Fatalf("expected ErrSlugTaken for case-insensitive duplicate, got %v", err)
	}
}

func TestCreateWorkspaceInvalidSlug(t *testing.T) {
	svc := NewService(NewMemoryStore())
	for _, slug := range []string{"", "ab", "-bad", "bad-", "has space", "UPPER!"} {
		if _, _, err := svc.CreateWorkspace(context.Background(), "X", slug, "u"); !errors.Is(err, ErrInvalidSlug) {
			t.Errorf("slug %q: expected ErrInvalidSlug, got %v", slug, err)
		}
	}
}

func TestAddMemberAndDuplicate(t *testing.T) {
	svc := NewService(NewMemoryStore())
	tenant, _ := newWorkspace(t, svc)
	ctx := context.Background()

	m, err := svc.AddMember(ctx, tenant.ID, "user-2", access.RoleMember)
	if err != nil {
		t.Fatalf("add member: %v", err)
	}
	if m.Role != access.RoleMember {
		t.Fatalf("role = %q", m.Role)
	}
	if _, err := svc.AddMember(ctx, tenant.ID, "user-2", access.RoleAdmin); !errors.Is(err, ErrAlreadyMember) {
		t.Fatalf("expected ErrAlreadyMember, got %v", err)
	}
}

func TestAddMemberInvalidRole(t *testing.T) {
	svc := NewService(NewMemoryStore())
	tenant, _ := newWorkspace(t, svc)
	if _, err := svc.AddMember(context.Background(), tenant.ID, "u2", access.Role("root")); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}

func TestLastOwnerCannotBeDemoted(t *testing.T) {
	svc := NewService(NewMemoryStore())
	tenant, owner := newWorkspace(t, svc)
	ctx := context.Background()

	if _, err := svc.ChangeRole(ctx, tenant.ID, owner.ID, access.RoleMember); !errors.Is(err, ErrLastOwner) {
		t.Fatalf("expected ErrLastOwner on demote, got %v", err)
	}
	if err := svc.RemoveMember(ctx, tenant.ID, owner.ID); !errors.Is(err, ErrLastOwner) {
		t.Fatalf("expected ErrLastOwner on remove, got %v", err)
	}
}

func TestSecondOwnerUnblocksDemotion(t *testing.T) {
	svc := NewService(NewMemoryStore())
	tenant, owner := newWorkspace(t, svc)
	ctx := context.Background()

	if _, err := svc.AddMember(ctx, tenant.ID, "user-2", access.RoleOwner); err != nil {
		t.Fatalf("add second owner: %v", err)
	}
	m, err := svc.ChangeRole(ctx, tenant.ID, owner.ID, access.RoleAdmin)
	if err != nil {
		t.Fatalf("demote with second owner present: %v", err)
	}
	if m.Role != access.RoleAdmin {
		t.Fatalf("role = %q, want admin", m.Role)
	}
}

func TestRemoveMember(t *testing.T) {
	svc := NewService(NewMemoryStore())
	tenant, _ := newWorkspace(t, svc)
	ctx := context.Background()

	m, err := svc.AddMember(ctx, tenant.ID, "user-2", access.RoleMember)
	if err != nil {
		t.Fatalf("add member: %v", err)
	}
	if err := svc.RemoveMember(ctx, tenant.ID, m.ID); err != nil {
		t.Fatalf("remove member: %v", err)
	}
	members, err := svc.Members(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("members: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("members = %d, want 1 (owner only)", len(members))
	}
}

func TestMembershipIsolationAcrossTenants(t *testing.T) {
	svc := NewService(NewMemoryStore())
	tenantA, _ := newWorkspace(t, svc)
	tenantB, _, err := svc.CreateWorkspace(context.Background(), "Beta", "beta", "user-b")
	if err != nil {
		t.Fatalf("create tenant B: %v", err)
	}
	ctx := context.Background()

	m, err := svc.AddMember(ctx, tenantA.ID, "user-2", access.RoleMember)
	if err != nil {
		t.Fatalf("add member: %v", err)
	}
	// Tenant B must not see or mutate tenant A's membership by ID.
	if _, err := svc.ChangeRole(ctx, tenantB.ID, m.ID, access.RoleAdmin); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-tenant ChangeRole: expected ErrNotFound, got %v", err)
	}
	if err := svc.RemoveMember(ctx, tenantB.ID, m.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-tenant RemoveMember: expected ErrNotFound, got %v", err)
	}
	if _, err := svc.MembershipForUser(ctx, tenantB.ID, "user-2"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-tenant MembershipForUser: expected ErrNotFound, got %v", err)
	}
}

func TestPrimaryMembershipOldestFirst(t *testing.T) {
	svc := NewService(NewMemoryStore())
	tenantA, _ := newWorkspace(t, svc)
	ctx := context.Background()

	// user-2 joins tenant A first, then gets their own tenant.
	if _, err := svc.AddMember(ctx, tenantA.ID, "user-2", access.RoleMember); err != nil {
		t.Fatalf("add member: %v", err)
	}
	if _, _, err := svc.CreateWorkspace(ctx, "Own", "own-space", "user-2"); err != nil {
		t.Fatalf("create second workspace: %v", err)
	}
	primary, err := svc.PrimaryMembership(ctx, "user-2")
	if err != nil {
		t.Fatalf("primary membership: %v", err)
	}
	if primary.TenantID != tenantA.ID {
		t.Fatalf("primary tenant = %s, want oldest (%s)", primary.TenantID, tenantA.ID)
	}
}

func TestUpdateTenant(t *testing.T) {
	svc := NewService(NewMemoryStore())
	tenant, _ := newWorkspace(t, svc)
	got, err := svc.Update(context.Background(), tenant.ID, "Acme Renamed", PlanTeam)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.Name != "Acme Renamed" || got.Plan != PlanTeam {
		t.Fatalf("update result: %+v", got)
	}
}
