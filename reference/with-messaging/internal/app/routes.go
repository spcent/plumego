package app

import (
	"net/http"

	"with-messaging/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the with-messaging demo.
func (a *App) RegisterRoutes() error {
	logger := a.Core.Logger()
	reg := newRouteReg(a.Core)
	reg.get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.WriteHealthResponse(w, r, "with-messaging", logger)
	}))
	reg.post("/events/publish", http.HandlerFunc(a.Handler.Publish))
	return reg.err
}

// routeAdder is the minimal interface shared by *core.App and *core.RouteGroup.
type routeAdder interface {
	Get(path string, h http.Handler) error
	Post(path string, h http.Handler) error
}

// routeReg wraps a routeAdder and records the first registration error.
// This lets the route table be written one route per line without per-call
// error checks; inspect reg.err once after all registrations.
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
