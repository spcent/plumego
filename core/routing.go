package core

import (
	"fmt"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

// registerRoute is the single implementation for all route registration.
// An empty name registers the route without a name.
func (a *App) registerRoute(method, path, name string, handler http.Handler) error {
	if handler == nil {
		params := map[string]any{"method": method, "path": path}
		if name != "" {
			params["name"] = name
		}
		return wrapCoreError(contract.ErrHandlerNil, "add_route", params)
	}

	if err := a.ensureMutable("add_route", "register route"); err != nil {
		return err
	}

	r := a.ensureRouter()
	if r == nil {
		return wrapCoreError(fmt.Errorf("router not configured"), "add_route", nil)
	}

	if name != "" {
		return r.AddRouteWithName(method, path, name, handler)
	}
	return r.AddRoute(method, path, handler)
}

func (a *App) addRoute(method, path string, handler http.Handler) error {
	return a.registerRoute(method, path, "", handler)
}

// =========================================================
// Canonical handler registration — standard library style
//
// These are the primary route registration methods. Use
// http.Handler as the handler type for all new routes.
// This is the only style shown in quick-start guides and
// canonical examples, and registration failures must be
// surfaced by the caller.
// =========================================================

// AddRoute registers a route and returns explicit errors for invalid registration.
// This is useful in strict boot wiring where route registration failures must be surfaced.
func (a *App) AddRoute(method, path string, handler http.Handler) error {
	return a.addRoute(method, path, handler)
}

// AddRouteWithName registers a named route and returns explicit registration errors.
func (a *App) AddRouteWithName(method, path, name string, handler http.Handler) error {
	return a.registerRoute(method, path, name, handler)
}

// URL resolves a named route against the owned app router.
func (a *App) URL(name string, params ...string) string {
	r := a.ensureRouter()
	if r == nil {
		return ""
	}
	return r.URL(name, params...)
}

// Get registers a GET route with the given handler.
func (a *App) Get(path string, handler http.Handler) error {
	return a.addRoute(router.GET, path, handler)
}

// Post registers a POST route with the given handler.
func (a *App) Post(path string, handler http.Handler) error {
	return a.addRoute(router.POST, path, handler)
}

// Put registers a PUT route with the given handler.
func (a *App) Put(path string, handler http.Handler) error {
	return a.addRoute(router.PUT, path, handler)
}

// Delete registers a DELETE route with the given handler.
func (a *App) Delete(path string, handler http.Handler) error {
	return a.addRoute(router.DELETE, path, handler)
}

// Patch registers a PATCH route with the given handler.
func (a *App) Patch(path string, handler http.Handler) error {
	return a.addRoute(router.PATCH, path, handler)
}

// Any registers a route for any HTTP method with the given handler.
func (a *App) Any(path string, handler http.Handler) error {
	return a.addRoute(router.ANY, path, handler)
}
