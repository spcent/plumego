package app

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spcent/plumego/middleware/auth"
	"github.com/spcent/plumego/x/frontend"

	"guardus/internal/handler"
	"guardus/web"
)

// RegisterRoutes wires HTTP routes for guardus.
//
// The route table mirrors gatus v5's public surface:
//   - /api/v1/config + /api/v1/endpoints/:key/* are public read endpoints.
//   - /api/v1/endpoints/statuses + /api/v1/endpoints/:key/statuses sit
//     behind Basic auth when security.basic is configured.
//   - /health, /metrics, and the SPA mount are non-API routes.
//
// Plumego's RouteGroup does not expose a Use() method, so the auth
// middleware is wrapped around each protected handler at registration time.
func (a *App) RegisterRoutes() error {
	v1 := newRouteReg(a.Core.Group("/api/v1"))

	// Public — these are intentionally outside the basic-auth gate so the
	// SPA can fetch /config and badges without credentials.
	v1.get("/config", handler.Config(a.Cfg))
	v1.get("/endpoints/:key/health/badge.svg", handler.HealthBadge(a.Store))
	v1.get("/endpoints/:key/health/badge.shields", handler.HealthBadgeShields(a.Store))
	v1.get("/endpoints/:key/uptimes/:duration", handler.UptimeRaw(a.Store))
	v1.get("/endpoints/:key/uptimes/:duration/badge.svg", handler.UptimeBadge(a.Store))
	v1.get("/endpoints/:key/response-times/:duration", handler.ResponseTimeRaw(a.Store))
	v1.get("/endpoints/:key/response-times/:duration/badge.svg", handler.ResponseTimeBadge(a.Cfg, a.Store))
	v1.get("/endpoints/:key/response-times/:duration/chart.svg", handler.ResponseTimeChart(a.Store))
	v1.get("/endpoints/:key/response-times/:duration/history", handler.ResponseTimeHistory(a.Store))
	v1.post("/endpoints/:key/external", handler.CreateExternalEndpointResult(a.Cfg, a.Store, a.Watchdog, a.Metrics, a.Core.Logger()))

	// Protected — when Security.Basic is unset, the gate is a no-op.
	statuses := http.Handler(handler.EndpointStatuses(a.Cfg, a.Store, a.Core.Logger()))
	statusByKey := http.Handler(handler.EndpointStatus(a.Cfg, a.Store, a.Core.Logger()))
	if a.Auth != nil {
		mw, err := auth.Authenticate(a.Auth, auth.WithAuthRealm("guardus"))
		if err != nil {
			return err
		}
		statuses = mw(statuses)
		statusByKey = mw(statusByKey)
	}
	v1.get("/endpoints/statuses", statuses)
	v1.get("/endpoints/:key/statuses", statusByKey)
	if v1.err != nil {
		return v1.err
	}

	root := newRouteReg(a.Core)
	root.get("/health", handler.Health())
	if a.Cfg.Metrics && a.Metrics != nil {
		root.get("/metrics", promhttp.HandlerFor(a.Metrics.Gatherer(), promhttp.HandlerOpts{}))
	}
	if root.err != nil {
		return root.err
	}

	// SPA mount comes last so the catch-all fallback doesn't shadow API
	// routes. fs.Sub strips the embed.FS "static/" prefix.
	staticFS, err := webStaticFS()
	if err != nil {
		return err
	}
	return frontend.RegisterFS(a.Core, staticFS,
		frontend.WithPrefix("/"),
		frontend.WithFallback(true),
	)
}

// webStaticFS returns the embedded SPA filesystem rooted at "static/".
func webStaticFS() (http.FileSystem, error) {
	sub, err := fsSub(web.FileSystem, web.RootPath)
	if err != nil {
		return nil, err
	}
	return http.FS(sub), nil
}

// routeAdder is the minimal interface shared by *core.App and *core.RouteGroup,
// allowing newRouteReg to work with both without type-switching.
type routeAdder interface {
	Get(path string, h http.Handler) error
	Post(path string, h http.Handler) error
	Put(path string, h http.Handler) error
	Delete(path string, h http.Handler) error
}

// routeReg wraps a routeAdder and records the first registration error.
type routeReg struct {
	adder routeAdder
	err   error
}

func newRouteReg(adder routeAdder) *routeReg { return &routeReg{adder: adder} }

func (r *routeReg) get(path string, h http.Handler)  { r.record(r.adder.Get(path, h)) }
func (r *routeReg) post(path string, h http.Handler) { r.record(r.adder.Post(path, h)) }

func (r *routeReg) record(err error) {
	if r.err == nil {
		r.err = err
	}
}
