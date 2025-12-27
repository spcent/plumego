package core

import (
	"context"

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
}

// ComponentFunc is a helper that allows building Components from
// individual callbacks without defining a concrete type.
type ComponentFunc struct {
	RegisterRoutesFn     func(r *router.Router)
	RegisterMiddlewareFn func(m *middleware.Registry)
	StartFn              func(ctx context.Context) error
	StopFn               func(ctx context.Context) error
	HealthFn             func() (string, health.HealthStatus)
}

func (c ComponentFunc) RegisterRoutes(r *router.Router) {
	if c.RegisterRoutesFn != nil {
		c.RegisterRoutesFn(r)
	}
}

func (c ComponentFunc) RegisterMiddleware(m *middleware.Registry) {
	if c.RegisterMiddlewareFn != nil {
		c.RegisterMiddlewareFn(m)
	}
}

func (c ComponentFunc) Start(ctx context.Context) error {
	if c.StartFn == nil {
		return nil
	}
	return c.StartFn(ctx)
}

func (c ComponentFunc) Stop(ctx context.Context) error {
	if c.StopFn == nil {
		return nil
	}
	return c.StopFn(ctx)
}

func (c ComponentFunc) Health() (string, health.HealthStatus) {
	if c.HealthFn == nil {
		return "component", health.HealthStatus{Status: health.StatusDegraded, Message: "health check not provided"}
	}
	return c.HealthFn()
}
