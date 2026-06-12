// Package tenantspace owns the workspace (tenant) and membership model.
package tenantspace

import (
	"errors"
	"time"

	"mini-saas-api/internal/domain/access"
)

// Sentinel errors translated to HTTP status codes by the handler layer.
var (
	ErrSlugTaken     = errors.New("tenantspace: slug already taken")
	ErrInvalidSlug   = errors.New("tenantspace: slug must be 3-64 lowercase letters, digits, or hyphens")
	ErrNotFound      = errors.New("tenantspace: not found")
	ErrAlreadyMember = errors.New("tenantspace: user is already a member")
	// ErrLastOwner blocks demoting or removing a tenant's only owner so every
	// tenant always has at least one owner.
	ErrLastOwner   = errors.New("tenantspace: cannot demote or remove the last owner")
	ErrInvalidRole = errors.New("tenantspace: invalid role")
)

// Plan labels — plain strings, no billing semantics. The project-count limit
// per plan lives in the project service configuration.
const (
	PlanFree = "free"
	PlanTeam = "team"
)

// Tenant is a workspace that owns memberships and projects.
type Tenant struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Plan      string    `json:"plan"`
	CreatedAt time.Time `json:"created_at"`
}

// Membership binds a user to a tenant with a role.
type Membership struct {
	ID        string      `json:"id"`
	TenantID  string      `json:"tenant_id"`
	UserID    string      `json:"user_id"`
	Role      access.Role `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
}
