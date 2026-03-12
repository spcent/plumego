package frontend

import (
	"context"
	"net/http"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

type frontendComponent struct {
	core.BaseComponent
	register func(r *router.Router) error
	name     string
	result   struct {
		registered bool
		err        error
	}
}

// NewFrontendComponentFromFS mounts a frontend bundle served from the provided
// filesystem as a pluggable component.
func NewFrontendComponentFromFS(fs http.FileSystem, opts ...Option) core.Component {
	return &frontendComponent{
		register: func(r *router.Router) error { return RegisterFS(r, fs, opts...) },
		name:     "frontend_fs",
	}
}

// NewFrontendComponentFromDir mounts a frontend bundle from a directory path as
// a pluggable component.
func NewFrontendComponentFromDir(dir string, opts ...Option) core.Component {
	return &frontendComponent{
		register: func(r *router.Router) error { return RegisterFromDir(r, dir, opts...) },
		name:     "frontend_dir",
	}
}

func (c *frontendComponent) RegisterRoutes(r *router.Router) {
	if c.register == nil {
		return
	}

	if err := c.register(r); err != nil {
		c.result.err = err
		return
	}

	c.result.registered = true
}

func (c *frontendComponent) RegisterMiddleware(_ *middleware.Registry) {}

func (c *frontendComponent) Start(_ context.Context) error { return nil }

func (c *frontendComponent) Stop(_ context.Context) error { return nil }

func (c *frontendComponent) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{
		Status:  health.StatusHealthy,
		Details: map[string]any{"registered": c.result.registered},
	}

	if c.result.err != nil {
		status.Status = health.StatusUnhealthy
		status.Message = c.result.err.Error()
		status.Details["error"] = c.result.err.Error()
	}

	name := c.name
	if name == "" {
		name = "frontend"
	}

	return name, status
}
