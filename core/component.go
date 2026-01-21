package core

import (
	"context"
	"reflect"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

// Component describes a pluggable unit that can contribute routes,
// middleware, and lifecycle hooks to the application.
type Component interface {
	RegisterRoutes(r *router.Router)
	RegisterMiddleware(m *middleware.Registry)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health() (name string, status health.HealthStatus)
	// Dependencies returns a list of component types that this component depends on.
	Dependencies() []reflect.Type
}

// BaseComponent provides a default implementation for Component interface
// with empty dependencies list and health check.
type BaseComponent struct{}

// RegisterRoutes implements Component.RegisterRoutes
func (b *BaseComponent) RegisterRoutes(r *router.Router) {}

// RegisterMiddleware implements Component.RegisterMiddleware
func (b *BaseComponent) RegisterMiddleware(m *middleware.Registry) {}

// Dependencies implements Component.Dependencies
func (b *BaseComponent) Dependencies() []reflect.Type {
	return nil
}

// Health implements Component.Health
func (b *BaseComponent) Health() (name string, status health.HealthStatus) {
	return "base-component", health.HealthStatus{Status: health.StatusHealthy}
}
