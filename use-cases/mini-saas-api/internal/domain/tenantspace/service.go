package tenantspace

import (
	"context"
	"regexp"
	"strings"
	"time"

	"mini-saas-api/internal/domain/access"
	"mini-saas-api/internal/domain/ident"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$`)

// Service implements workspace and membership lifecycle including the
// last-owner invariant: a tenant always has at least one owner.
type Service struct {
	repo Repository
}

// NewService constructs a Service over the given repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateWorkspace creates a tenant and its first owner membership.
func (s *Service) CreateWorkspace(ctx context.Context, name, slug, ownerUserID string) (Tenant, Membership, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !slugPattern.MatchString(slug) {
		return Tenant{}, Membership{}, ErrInvalidSlug
	}
	tenantID, err := ident.New()
	if err != nil {
		return Tenant{}, Membership{}, err
	}
	now := time.Now().UTC()
	t := Tenant{
		ID:        tenantID,
		Slug:      slug,
		Name:      strings.TrimSpace(name),
		Plan:      PlanFree,
		CreatedAt: now,
	}
	if err := s.repo.CreateTenant(ctx, t); err != nil {
		return Tenant{}, Membership{}, err
	}
	m, err := s.addMembership(ctx, tenantID, ownerUserID, access.RoleOwner)
	if err != nil {
		return Tenant{}, Membership{}, err
	}
	return t, m, nil
}

// Get returns the tenant by ID.
func (s *Service) Get(ctx context.Context, tenantID string) (Tenant, error) {
	t, ok, err := s.repo.TenantByID(ctx, tenantID)
	if err != nil {
		return Tenant{}, err
	}
	if !ok {
		return Tenant{}, ErrNotFound
	}
	return t, nil
}

// Update renames the workspace and/or changes its plan label.
// Empty fields keep their current value.
func (s *Service) Update(ctx context.Context, tenantID, name, plan string) (Tenant, error) {
	t, err := s.Get(ctx, tenantID)
	if err != nil {
		return Tenant{}, err
	}
	if name = strings.TrimSpace(name); name != "" {
		t.Name = name
	}
	if plan = strings.TrimSpace(plan); plan != "" {
		t.Plan = plan
	}
	if err := s.repo.UpdateTenant(ctx, t); err != nil {
		return Tenant{}, err
	}
	return t, nil
}

// Members lists all memberships of the tenant, oldest first.
func (s *Service) Members(ctx context.Context, tenantID string) ([]Membership, error) {
	return s.repo.Memberships(ctx, tenantID)
}

// AddMember adds an existing user to the tenant with the given role.
func (s *Service) AddMember(ctx context.Context, tenantID, userID string, role access.Role) (Membership, error) {
	if !role.Valid() {
		return Membership{}, ErrInvalidRole
	}
	return s.addMembership(ctx, tenantID, userID, role)
}

func (s *Service) addMembership(ctx context.Context, tenantID, userID string, role access.Role) (Membership, error) {
	id, err := ident.New()
	if err != nil {
		return Membership{}, err
	}
	m := Membership{
		ID:        id,
		TenantID:  tenantID,
		UserID:    userID,
		Role:      role,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.AddMembership(ctx, m); err != nil {
		return Membership{}, err
	}
	return m, nil
}

// ChangeRole updates a membership's role, refusing to demote the last owner.
func (s *Service) ChangeRole(ctx context.Context, tenantID, membershipID string, role access.Role) (Membership, error) {
	if !role.Valid() {
		return Membership{}, ErrInvalidRole
	}
	m, ok, err := s.repo.MembershipByID(ctx, tenantID, membershipID)
	if err != nil {
		return Membership{}, err
	}
	if !ok {
		return Membership{}, ErrNotFound
	}
	if m.Role == access.RoleOwner && role != access.RoleOwner {
		last, err := s.isLastOwner(ctx, tenantID, m.ID)
		if err != nil {
			return Membership{}, err
		}
		if last {
			return Membership{}, ErrLastOwner
		}
	}
	m.Role = role
	if err := s.repo.UpdateMembership(ctx, m); err != nil {
		return Membership{}, err
	}
	return m, nil
}

// RemoveMember deletes a membership, refusing to remove the last owner.
func (s *Service) RemoveMember(ctx context.Context, tenantID, membershipID string) error {
	m, ok, err := s.repo.MembershipByID(ctx, tenantID, membershipID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	if m.Role == access.RoleOwner {
		last, err := s.isLastOwner(ctx, tenantID, m.ID)
		if err != nil {
			return err
		}
		if last {
			return ErrLastOwner
		}
	}
	return s.repo.RemoveMembership(ctx, tenantID, membershipID)
}

// MembershipForUser returns the caller's membership in the tenant.
func (s *Service) MembershipForUser(ctx context.Context, tenantID, userID string) (Membership, error) {
	m, ok, err := s.repo.MembershipByUser(ctx, tenantID, userID)
	if err != nil {
		return Membership{}, err
	}
	if !ok {
		return Membership{}, ErrNotFound
	}
	return m, nil
}

// MembershipBySlug returns the user's membership in the workspace identified
// by slug. Unknown slug and non-membership both yield ErrNotFound.
func (s *Service) MembershipBySlug(ctx context.Context, slug, userID string) (Membership, error) {
	t, ok, err := s.repo.TenantBySlug(ctx, slug)
	if err != nil {
		return Membership{}, err
	}
	if !ok {
		return Membership{}, ErrNotFound
	}
	return s.MembershipForUser(ctx, t.ID, userID)
}

// PrimaryMembership returns the user's oldest membership — the default tenant
// for login when no tenant is specified.
func (s *Service) PrimaryMembership(ctx context.Context, userID string) (Membership, error) {
	all, err := s.repo.MembershipsForUser(ctx, userID)
	if err != nil {
		return Membership{}, err
	}
	if len(all) == 0 {
		return Membership{}, ErrNotFound
	}
	return all[0], nil
}

// isLastOwner reports whether membershipID is the only owner of the tenant.
func (s *Service) isLastOwner(ctx context.Context, tenantID, membershipID string) (bool, error) {
	members, err := s.repo.Memberships(ctx, tenantID)
	if err != nil {
		return false, err
	}
	for _, other := range members {
		if other.ID != membershipID && other.Role == access.RoleOwner {
			return false, nil
		}
	}
	return true, nil
}
