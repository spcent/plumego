package app

import (
	"github.com/spcent/plumego/reference/standard-service/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the reference application.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	health := handler.HealthHandler{ServiceName: "plumego-reference"}

	if err := a.Core.Get("/", api.Hello); err != nil {
		return err
	}
	if err := a.Core.Get("/healthz", health.Live); err != nil {
		return err
	}
	if err := a.Core.Get("/readyz", health.Ready); err != nil {
		return err
	}
	if err := a.Core.Get("/api/hello", api.Hello); err != nil {
		return err
	}
	if err := a.Core.Get("/api/status", api.Status); err != nil {
		return err
	}
	return nil
}
