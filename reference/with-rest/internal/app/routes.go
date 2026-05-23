package app

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

// RegisterRoutes wires all HTTP routes for the with-rest demo.
func (a *App) RegisterRoutes() error {
	reg := newRouteReg(a.Core)
	reg.get("/api/hello", http.HandlerFunc(hello))
	// x/rest resource: list, get, create, update, delete
	reg.get("/api/users", http.HandlerFunc(a.Users.Index))
	reg.get("/api/users/:id", http.HandlerFunc(a.Users.Show))
	reg.post("/api/users", http.HandlerFunc(a.Users.Create))
	reg.post("/api/items", a.Items)
	reg.put("/api/users/:id", http.HandlerFunc(a.Users.Update))
	reg.delete("/api/users/:id", http.HandlerFunc(a.Users.Delete))
	return reg.err
}

func hello(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{
		"message": "hello from with-rest",
	}, nil)
}

// routeAdder is the minimal interface shared by *core.App and *core.RouteGroup.
type routeAdder interface {
	Get(path string, h http.Handler) error
	Post(path string, h http.Handler) error
	Put(path string, h http.Handler) error
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
func (r *routeReg) put(path string, h http.Handler)    { r.record(r.adder.Put(path, h)) }
func (r *routeReg) delete(path string, h http.Handler) { r.record(r.adder.Delete(path, h)) }
func (r *routeReg) record(err error) {
	if r.err == nil {
		r.err = err
	}
}
