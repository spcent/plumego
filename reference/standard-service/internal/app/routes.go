package app

import (
	"net/http"

	"standard-service/internal/domain/item"
	"standard-service/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the reference application.
// One method, one path, one handler per line — no scanning, no annotations.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	// HealthHandler receives no Checkers here because the reference has no real
	// dependencies to probe. In a production service, pass one ReadinessChecker
	// per dependency (database, cache, downstream) so /readyz reflects real state.
	health := handler.HealthHandler{ServiceName: "plumego-reference"}
	// ItemHandler demonstrates constructor injection: the concrete domain store
	// is created here and passed through the interface the handler declared.
	items := handler.ItemHandler{Repo: item.NewMemoryStore()}

	// Root landing — minimal identity. Full endpoint listing is at /api/hello.
	if err := a.Core.Get("/", http.HandlerFunc(api.Root)); err != nil {
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
	// Collection: list (GET) and create (POST) share the same path, different methods.
	// GET  /api/v1/items              → 200 [items…]
	// POST /api/v1/items {"name":"…"} → 201 item
	if err := a.Core.Get("/api/v1/items", http.HandlerFunc(items.List)); err != nil {
		return err
	}
	if err := a.Core.Post("/api/v1/items", http.HandlerFunc(items.Create)); err != nil {
		return err
	}
	// Member: fetch (GET) and delete (DELETE) share the same path, different methods.
	// GET    /api/v1/items/:id → 200 item or 404 TypeNotFound
	// DELETE /api/v1/items/:id → 204      or 404 TypeNotFound
	if err := a.Core.Get("/api/v1/items/:id", http.HandlerFunc(items.GetByID)); err != nil {
		return err
	}
	if err := a.Core.Delete("/api/v1/items/:id", http.HandlerFunc(items.Delete)); err != nil {
		return err
	}
	return nil
}
