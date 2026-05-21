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
	root := newRouteReg(a.Core)
	root.get("/", http.HandlerFunc(api.Root))
	root.get("/healthz", http.HandlerFunc(health.Live))
	root.get("/readyz", http.HandlerFunc(health.Ready))
	root.get("/api/hello", http.HandlerFunc(api.Hello))
	root.get("/api/status", http.HandlerFunc(api.Status))
	if root.err != nil {
		return root.err
	}

	// Versioned API — all routes under /api/v1 share this group prefix.
	// Query-param binding: GET /api/v1/greet?name=Alice → 200
	//                      GET /api/v1/greet            → 400 TypeRequired
	// Collection:          GET  /api/v1/items              → 200 {items:[…],total:N,…}
	//                      POST /api/v1/items {"name":"…"} → 201 item
	// Member:              GET    /api/v1/items/:id → 200 item or 404
	//                      DELETE /api/v1/items/:id → 204      or 404
	v1 := newRouteReg(a.Core.Group("/api/v1"))
	v1.get("/greet", http.HandlerFunc(api.Greet))
	v1.get("/items", http.HandlerFunc(items.List))
	v1.post("/items", http.HandlerFunc(items.Create))
	v1.get("/items/:id", http.HandlerFunc(items.GetByID))
	v1.delete("/items/:id", http.HandlerFunc(items.Delete))
	return v1.err
}

// routeAdder is the minimal interface shared by *core.App and *core.RouteGroup,
// allowing newRouteReg to work with both without type-switching.
type routeAdder interface {
	Get(path string, h http.Handler) error
	Post(path string, h http.Handler) error
	Delete(path string, h http.Handler) error
}

// routeReg wraps a routeAdder and records the first registration error.
// This lets the route table be written one route per line without per-call
// error checks; inspect reg.err once after all registrations.
type routeReg struct {
	adder routeAdder
	err   error
}

func newRouteReg(adder routeAdder) *routeReg { return &routeReg{adder: adder} }

func (r *routeReg) get(path string, h http.Handler)    { r.record(r.adder.Get(path, h)) }
func (r *routeReg) post(path string, h http.Handler)   { r.record(r.adder.Post(path, h)) }
func (r *routeReg) delete(path string, h http.Handler) { r.record(r.adder.Delete(path, h)) }
func (r *routeReg) record(err error) {
	if r.err == nil {
		r.err = err
	}
}
