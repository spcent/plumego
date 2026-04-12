package core

import (
	"fmt"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

const methodAny = "ANY"

// registerRoute is the single implementation for all route registration.
func (a *App) registerRoute(method, path string, handler http.Handler, opts ...router.RouteOption) error {
	if handler == nil {
		params := map[string]any{"method": method, "path": path}
		return wrapCoreError(contract.ErrHandlerNil, "add_route", params)
	}

	if err := a.ensureMutable("add_route", "register route"); err != nil {
		return err
	}

	r := a.ensureRouter()
	if r == nil {
		return wrapCoreError(fmt.Errorf("router not configured"), "add_route", nil)
	}

	return r.AddRoute(method, path, handler, opts...)
}

func (a *App) addRoute(method, path string, handler http.Handler) error {
	return a.registerRoute(method, path, handler)
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
func (a *App) AddRoute(method, path string, handler http.Handler, opts ...router.RouteOption) error {
	return a.registerRoute(method, path, handler, opts...)
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
	return a.addRoute(http.MethodGet, path, handler)
}

// Post registers a POST route with the given handler.
func (a *App) Post(path string, handler http.Handler) error {
	return a.addRoute(http.MethodPost, path, handler)
}

// Put registers a PUT route with the given handler.
func (a *App) Put(path string, handler http.Handler) error {
	return a.addRoute(http.MethodPut, path, handler)
}

// Delete registers a DELETE route with the given handler.
func (a *App) Delete(path string, handler http.Handler) error {
	return a.addRoute(http.MethodDelete, path, handler)
}

// Patch registers a PATCH route with the given handler.
func (a *App) Patch(path string, handler http.Handler) error {
	return a.addRoute(http.MethodPatch, path, handler)
}

// Any registers a route for any HTTP method with the given handler.
func (a *App) Any(path string, handler http.Handler) error {
	return a.addRoute(methodAny, path, handler)
}

// GetWithName registers a named GET route with the given handler.
func (a *App) GetWithName(path, name string, handler http.Handler) error {
	return a.registerRoute(http.MethodGet, path, handler, router.WithRouteName(name))
}

// PostWithName registers a named POST route with the given handler.
func (a *App) PostWithName(path, name string, handler http.Handler) error {
	return a.registerRoute(http.MethodPost, path, handler, router.WithRouteName(name))
}

// PutWithName registers a named PUT route with the given handler.
func (a *App) PutWithName(path, name string, handler http.Handler) error {
	return a.registerRoute(http.MethodPut, path, handler, router.WithRouteName(name))
}

// DeleteWithName registers a named DELETE route with the given handler.
func (a *App) DeleteWithName(path, name string, handler http.Handler) error {
	return a.registerRoute(http.MethodDelete, path, handler, router.WithRouteName(name))
}

// PatchWithName registers a named PATCH route with the given handler.
func (a *App) PatchWithName(path, name string, handler http.Handler) error {
	return a.registerRoute(http.MethodPatch, path, handler, router.WithRouteName(name))
}

// AnyWithName registers a named route for any HTTP method with the given handler.
func (a *App) AnyWithName(path, name string, handler http.Handler) error {
	return a.registerRoute(methodAny, path, handler, router.WithRouteName(name))
}
