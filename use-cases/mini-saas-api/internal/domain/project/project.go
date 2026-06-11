// Package project owns the project model, tenant-scoped repository, and service.
package project

import (
	"errors"
	"time"
)

// Sentinel errors translated to HTTP status codes by the handler layer.
var (
	// ErrNotFound is returned for both missing projects and projects that
	// belong to another tenant, so responses never leak cross-tenant existence.
	ErrNotFound      = errors.New("project: not found")
	ErrInvalidStatus = errors.New("project: invalid status")
	ErrNameRequired  = errors.New("project: name is required")
	// ErrLimitReached is returned when the tenant's plan project quota is full.
	ErrLimitReached = errors.New("project: plan project limit reached")
)

// Status is a project lifecycle state.
type Status string

const (
	StatusActive   Status = "active"
	StatusPaused   Status = "paused"
	StatusArchived Status = "archived"
)

// Valid reports whether s is a recognized status.
func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusPaused, StatusArchived:
		return true
	default:
		return false
	}
}

// Project is a tenant-owned resource.
type Project struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
