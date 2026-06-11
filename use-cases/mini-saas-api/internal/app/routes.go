package app

import (
	"context"
	"net/http"
	"time"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
	tenantquota "github.com/spcent/plumego/x/tenant/quota"
	tenantratelimit "github.com/spcent/plumego/x/tenant/ratelimit"
	tenantresolve "github.com/spcent/plumego/x/tenant/resolve"
	"mini-saas-api/internal/domain/access"
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
	tenantH := handler.TenantHandler{
		Spaces:   a.Spaces,
		Projects: a.Projects,
		Audit:    a.Audit,
		Logger:   logger,
	}
	membersH := handler.MembersHandler{
		Users:  a.Users,
		Spaces: a.Spaces,
		Audit:  a.Audit,
		Logger: logger,
	}
	projectsC := handler.NewProjectsController(a.Projects, a.Audit, logger)

	// Per-route guards. authn is per-route (not global) so public endpoints
	// stay guard-free and the wiring remains visible at the route table.
	requireAuth := handler.RequireAuth(a.JWT, logger)
	requireAdmin := handler.RequireRole(access.RoleAdmin, logger)
	abuse := a.authGuard.Middleware()

	// Tenant chain for the authenticated API surface (x/tenant, beta):
	//   resolve   → lifts authn.Principal.TenantID into the tenant context
	//   ratelimit → per-tenant token bucket (sustained APP_TENANT_RPS, burst APP_TENANT_BURST)
	//   quota     → per-tenant fixed-window request quota (APP_TENANT_QUOTA_PER_MINUTE)
	// Enforcement state is per tenant; the config providers are uniform.
	tenantChain := middleware.NewChain(
		tenantresolve.Middleware(tenantresolve.Options{}),
		tenantratelimit.Middleware(tenantratelimit.Options{
			Limiter: tenantcore.NewTokenBucketRateLimiter(uniformRateLimits{
				rps:   a.Cfg.App.TenantRPS,
				burst: a.Cfg.App.TenantBurst,
			}),
		}),
		tenantquota.Middleware(tenantquota.Options{
			Manager: tenantcore.NewFixedWindowQuotaManager(uniformQuota{
				requestsPerMinute: a.Cfg.App.TenantQuotaPerMinute,
			}),
		}),
	)
	// authed wraps a handler with bearer-token auth followed by the tenant chain.
	authed := func(h http.Handler) http.Handler { return requireAuth(tenantChain.Build(h)) }
	// idem adds Idempotency-Key replay protection (stable store/idempotency);
	// a no-op when the header is absent. Must sit inside authed.
	idem := handler.Idempotent(a.Idem, 24*time.Hour, logger)

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
	// Authenticated surface: auth → tenant resolve → rate limit → quota.
	// Mutating routes additionally honor Idempotency-Key via idem.
	v1.get("/me", authed(http.HandlerFunc(meH.Get)))
	v1.get("/tenant", authed(http.HandlerFunc(tenantH.Get)))
	v1.patch("/tenant", authed(requireAdmin(idem(http.HandlerFunc(tenantH.Update)))))
	v1.get("/tenant/members", authed(http.HandlerFunc(membersH.List)))
	v1.post("/tenant/members", authed(requireAdmin(idem(http.HandlerFunc(membersH.Add)))))
	v1.patch("/tenant/members/:id", authed(requireAdmin(idem(http.HandlerFunc(membersH.ChangeRole)))))
	v1.delete("/tenant/members/:id", authed(requireAdmin(idem(http.HandlerFunc(membersH.Remove)))))
	// Projects CRUD through the x/rest resource controller (beta).
	v1.get("/projects", authed(http.HandlerFunc(projectsC.Index)))
	v1.post("/projects", authed(idem(http.HandlerFunc(projectsC.Create))))
	v1.get("/projects/:id", authed(http.HandlerFunc(projectsC.Show)))
	v1.put("/projects/:id", authed(idem(http.HandlerFunc(projectsC.Update))))
	v1.delete("/projects/:id", authed(requireAdmin(idem(http.HandlerFunc(projectsC.Delete)))))
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
