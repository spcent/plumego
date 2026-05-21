package app

import (
	"net/http"

	"standard-service/internal/domain/item"
	"standard-service/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the reference application.
// One method, one path, one handler per line — no scanning, no annotations.
// Routes that share a common path prefix are registered through a RouteGroup
// so the prefix is declared once and not repeated on every line.
func (a *App) RegisterRoutes() error {
	api := handler.APIHandler{}
	// HealthHandler receives no Checkers here because the reference has no real
	// dependencies to probe. In a production service, pass one ReadinessChecker
	// per dependency (database, cache, downstream) so /readyz reflects real state.
	health := handler.HealthHandler{ServiceName: a.Cfg.App.ServiceName}
	// ItemHandler demonstrates constructor injection: the concrete domain store
	// is created here and passed through the interface the handler declared.
	items := handler.ItemHandler{Repo: item.NewMemoryStore()}

	// Top-level routes registered directly on the app.
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

	// Versioned API — all routes under /api/v1 share this group prefix.
	// Query-param binding: GET /api/v1/greet?name=Alice → 200
	//                      GET /api/v1/greet            → 400 TypeRequired
	// Collection:          GET  /api/v1/items              → 200 [items…]
	//                      POST /api/v1/items {"name":"…"} → 201 item
	// Member:              GET    /api/v1/items/:id → 200 item or 404
	//                      DELETE /api/v1/items/:id → 204      or 404
	v1 := a.Core.Group("/api/v1")
	if err := v1.Get("/greet", http.HandlerFunc(api.Greet)); err != nil {
		return err
	}
	if err := v1.Get("/items", http.HandlerFunc(items.List)); err != nil {
		return err
	}
	if err := v1.Post("/items", http.HandlerFunc(items.Create)); err != nil {
		return err
	}
	if err := v1.Get("/items/:id", http.HandlerFunc(items.GetByID)); err != nil {
		return err
	}
	if err := v1.Delete("/items/:id", http.HandlerFunc(items.Delete)); err != nil {
		return err
	}
	return nil
}
