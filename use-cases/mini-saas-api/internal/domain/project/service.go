package project

import (
	"context"
	"strings"
	"time"

	"mini-saas-api/internal/domain/ident"
)

// Limits maps plan labels to the maximum number of projects per tenant.
// Zero or missing means unlimited.
type Limits map[string]int

// DefaultLimits returns the plan → project-count limits used by the app.
func DefaultLimits() Limits {
	return Limits{
		"free": 5,
		"team": 100,
	}
}

// PlanLookup returns the plan label for a tenant. The tenantspace service
// satisfies this via a small adapter in app wiring.
type PlanLookup func(ctx context.Context, tenantID string) (string, error)

// Service implements tenant-scoped project CRUD with a per-plan count limit.
type Service struct {
	repo   Repository
	limits Limits
	plan   PlanLookup
}

// NewService constructs a Service. plan may be nil when no per-plan limit is
// wanted (all tenants unlimited).
func NewService(repo Repository, limits Limits, plan PlanLookup) *Service {
	return &Service{repo: repo, limits: limits, plan: plan}
}

// List returns the tenant's projects, oldest first.
func (s *Service) List(ctx context.Context, tenantID string) ([]Project, error) {
	return s.repo.List(ctx, tenantID)
}

// Get returns one project; missing and cross-tenant IDs both yield ErrNotFound.
func (s *Service) Get(ctx context.Context, tenantID, id string) (Project, error) {
	p, ok, err := s.repo.ByID(ctx, tenantID, id)
	if err != nil {
		return Project{}, err
	}
	if !ok {
		return Project{}, ErrNotFound
	}
	return p, nil
}

// Create adds a project after checking the tenant's plan project limit.
func (s *Service) Create(ctx context.Context, tenantID, createdBy, name, description string) (Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Project{}, ErrNameRequired
	}
	if err := s.checkLimit(ctx, tenantID); err != nil {
		return Project{}, err
	}
	id, err := ident.New()
	if err != nil {
		return Project{}, err
	}
	now := time.Now().UTC()
	p := Project{
		ID:          id,
		TenantID:    tenantID,
		Name:        name,
		Description: strings.TrimSpace(description),
		Status:      StatusActive,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return Project{}, err
	}
	return p, nil
}

// Update replaces name/description/status; empty fields keep current values.
func (s *Service) Update(ctx context.Context, tenantID, id, name, description string, status Status) (Project, error) {
	p, err := s.Get(ctx, tenantID, id)
	if err != nil {
		return Project{}, err
	}
	if name = strings.TrimSpace(name); name != "" {
		p.Name = name
	}
	if description = strings.TrimSpace(description); description != "" {
		p.Description = description
	}
	if status != "" {
		if !status.Valid() {
			return Project{}, ErrInvalidStatus
		}
		p.Status = status
	}
	p.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, p); err != nil {
		return Project{}, err
	}
	return p, nil
}

// Delete removes a project; missing and cross-tenant IDs both yield ErrNotFound.
func (s *Service) Delete(ctx context.Context, tenantID, id string) error {
	if _, err := s.Get(ctx, tenantID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, tenantID, id)
}

// Usage returns current project count and the plan limit (0 = unlimited).
func (s *Service) Usage(ctx context.Context, tenantID string) (count, limit int, err error) {
	count, err = s.repo.Count(ctx, tenantID)
	if err != nil {
		return 0, 0, err
	}
	limit, err = s.limitFor(ctx, tenantID)
	if err != nil {
		return 0, 0, err
	}
	return count, limit, nil
}

func (s *Service) checkLimit(ctx context.Context, tenantID string) error {
	limit, err := s.limitFor(ctx, tenantID)
	if err != nil {
		return err
	}
	if limit <= 0 {
		return nil
	}
	count, err := s.repo.Count(ctx, tenantID)
	if err != nil {
		return err
	}
	if count >= limit {
		return ErrLimitReached
	}
	return nil
}

func (s *Service) limitFor(ctx context.Context, tenantID string) (int, error) {
	if s.plan == nil || s.limits == nil {
		return 0, nil
	}
	plan, err := s.plan(ctx, tenantID)
	if err != nil {
		return 0, err
	}
	return s.limits[plan], nil
}
