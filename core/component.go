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
