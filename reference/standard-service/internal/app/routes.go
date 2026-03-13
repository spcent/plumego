package app

import (
	"github.com/spcent/plumego/reference/standard-service/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the reference application.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	health := handler.HealthHandler{ServiceName: "plumego-reference"}

	a.Core.Get("/", api.Hello)
	a.Core.Get("/healthz", health.Live)
	a.Core.Get("/readyz", health.Ready)
	a.Core.Get("/api/hello", api.Hello)
	a.Core.Get("/api/status", api.Status)
	return nil
}
