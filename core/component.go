package core

import (
	"context"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

// Component is a compatibility interface for legacy extension wrappers.
// It is not part of Plumego's canonical application path.
type Component interface {
	RegisterRoutes(r *router.Router)
	RegisterMiddleware(m *middleware.Registry)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health() (name string, status health.HealthStatus)
}

// BaseComponent provides a default implementation for the compatibility
// Component interface.
type BaseComponent struct{}

// RegisterRoutes implements Component.RegisterRoutes
func (b *BaseComponent) RegisterRoutes(r *router.Router) {}

// RegisterMiddleware implements Component.RegisterMiddleware
func (b *BaseComponent) RegisterMiddleware(m *middleware.Registry) {}

// Health implements Component.Health
func (b *BaseComponent) Health() (name string, status health.HealthStatus) {
	return "base-component", health.HealthStatus{Status: health.StatusHealthy}
}
