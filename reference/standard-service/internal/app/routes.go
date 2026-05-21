package app

import (
	"net/http"

	"standard-service/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the reference application.
// One method, one path, one handler per line — no scanning, no annotations.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	health := handler.HealthHandler{ServiceName: "plumego-reference"}
	// ItemHandler demonstrates constructor injection: the concrete store is
	// created here and passed through the interface the handler declared.
	items := handler.ItemHandler{Repo: handler.NewMemoryItemStore()}

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
	// Query-param binding and structured error response.
	// GET /api/v1/greet?name=Alice → 200   GET /api/v1/greet → 400 TypeRequired
	if err := a.Core.Get("/api/v1/greet", http.HandlerFunc(api.Greet)); err != nil {
		return err
	}
	// Request body decode (POST) and path parameter extraction (GET /:id).
	// POST /api/v1/items {"name":"widget"} → 201
	// GET  /api/v1/items/:id              → 200 or 404 TypeNotFound
	if err := a.Core.Post("/api/v1/items", http.HandlerFunc(items.Create)); err != nil {
		return err
	}
	if err := a.Core.Get("/api/v1/items/:id", http.HandlerFunc(items.GetByID)); err != nil {
		return err
	}
	return nil
}
