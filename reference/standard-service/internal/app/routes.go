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
	// /api/v1/greet demonstrates: query-param binding, structured error response.
	// GET /api/v1/greet?name=Alice  → 200 {"message":"hello, Alice"}
	// GET /api/v1/greet             → 400 with structured APIError body
	if err := a.Core.Get("/api/v1/greet", http.HandlerFunc(api.Greet)); err != nil {
		return err
	}
	return nil
}
