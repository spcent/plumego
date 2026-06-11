package app

import (
	"context"
	"net/http"

	"github.com/spcent/plumego/health"
	"mini-saas-api/internal/handler"
)

// RegisterRoutes wires all HTTP routes for mini-saas-api.
// One method, one path, one handler per line — no scanning, no annotations.
func (a *App) RegisterRoutes() error {
	logger := a.Core.Logger()
	healthH := handler.HealthHandler{
		ServiceName: a.Cfg.App.ServiceName,
		Logger:      logger,
		Checkers:    []health.ComponentChecker{appStateChecker{}},
	}
	authH := handler.AuthHandler{
		Users:  a.Users,
		Spaces: a.Spaces,
		Tokens: a.Issuer,
		Audit:  a.Audit,
		Logger: logger,
	}
	meH := handler.MeHandler{
		Users:  a.Users,
		Spaces: a.Spaces,
		Logger: logger,
	}

	// Per-route guards. authn is per-route (not global) so public endpoints
	// stay guard-free and the wiring remains visible at the route table.
	requireAuth := handler.RequireAuth(a.JWT, logger)
	abuse := a.authGuard.Middleware()

	root := newRouteReg(a.Core)
	root.get("/healthz", http.HandlerFunc(healthH.Live))
	root.get("/readyz", http.HandlerFunc(healthH.Ready))
	if root.err != nil {
		return root.err
	}

	v1 := newRouteReg(a.Core.Group("/api/v1"))
	// Public auth endpoints, brute-force-guarded by the abuse token bucket.
	v1.post("/auth/signup", abuse(http.HandlerFunc(authH.Signup)))
	v1.post("/auth/login", abuse(http.HandlerFunc(authH.Login)))
	v1.post("/auth/refresh", abuse(http.HandlerFunc(authH.Refresh)))
	// Authenticated surface.
	v1.get("/me", requireAuth(http.HandlerFunc(meH.Get)))
	return v1.err
}

// appStateChecker is a synthetic readiness probe.
// Replace with real dependency probes (store/kv ping, etc.) in subsequent cards.
type appStateChecker struct{}

func (appStateChecker) Name() string                  { return "app" }
func (appStateChecker) Check(_ context.Context) error { return nil }

// routeAdder is the minimal interface shared by *core.App and *core.RouteGroup.
type routeAdder interface {
	Get(path string, h http.Handler) error
	Post(path string, h http.Handler) error
	Put(path string, h http.Handler) error
	Patch(path string, h http.Handler) error
	Delete(path string, h http.Handler) error
}

// routeReg wraps a routeAdder and accumulates the first registration error.
type routeReg struct {
	adder routeAdder
	err   error
}

func newRouteReg(adder routeAdder) *routeReg { return &routeReg{adder: adder} }

func (r *routeReg) get(path string, h http.Handler)    { r.record(r.adder.Get(path, h)) }
func (r *routeReg) post(path string, h http.Handler)   { r.record(r.adder.Post(path, h)) }
func (r *routeReg) put(path string, h http.Handler)    { r.record(r.adder.Put(path, h)) }
func (r *routeReg) patch(path string, h http.Handler)  { r.record(r.adder.Patch(path, h)) }
func (r *routeReg) delete(path string, h http.Handler) { r.record(r.adder.Delete(path, h)) }

func (r *routeReg) record(err error) {
	if r.err == nil {
		r.err = err
	}
}
