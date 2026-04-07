package app

import (
	"net/http"

	"github.com/spcent/plumego/reference/standard-service/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the reference application.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	health := handler.HealthHandler{ServiceName: "plumego-reference"}

	if err := a.Core.Get("/", http.HandlerFunc(api.Hello)); err != nil {
		return err
	}
	if err := a.Core.Get("/healthz", http.HandlerFunc(health.Live)); err != nil {
		return err
	}
	if err := a.Core.Get("/readyz", http.HandlerFunc(health.Ready)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/hello", http.HandlerFunc(api.Hello)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/status", http.HandlerFunc(api.Status)); err != nil {
		return err
	}
	return nil
}
