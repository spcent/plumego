package core

import (
	"fmt"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
)

func (a *App) registerRoute(method, path string, handler http.Handler) {
	if handler == nil {
		err := contract.WrapError(contract.ErrHandlerNil, "add_route", "core", map[string]any{
			"method": method,
			"path":   path,
		})
		a.logError("route registration failed", err, log.Fields{"method": method, "path": path})
		return
	}

	if err := a.ensureMutable("add_route", "register route"); err != nil {
		a.logError("route registration failed", err, log.Fields{"method": method, "path": path})
		return
	}

	r := a.ensureRouter()
	if r == nil {
		err := contract.WrapError(fmt.Errorf("router not configured"), "add_route", "core", nil)
		a.logError("route registration failed", err, log.Fields{"method": method, "path": path})
		return
	}

	if err := r.AddRoute(method, path, handler); err != nil {
		a.logError("route registration failed", err, log.Fields{"method": method, "path": path})
	}
}

// Get registers a GET route with the given handler.
func (a *App) Get(path string, handler http.HandlerFunc) {
	a.registerRoute(router.GET, path, handler)
}

// Post registers a POST route with the given handler.
func (a *App) Post(path string, handler http.HandlerFunc) {
	a.registerRoute(router.POST, path, handler)
}

// Put registers a PUT route with the given handler.
func (a *App) Put(path string, handler http.HandlerFunc) {
	a.registerRoute(router.PUT, path, handler)
}

// Delete registers a DELETE route with the given handler.
func (a *App) Delete(path string, handler http.HandlerFunc) {
	a.registerRoute(router.DELETE, path, handler)
}

// Patch registers a PATCH route with the given handler.
func (a *App) Patch(path string, handler http.HandlerFunc) {
	a.registerRoute(router.PATCH, path, handler)
}

// Any registers a route for any HTTP method with the given handler.
func (a *App) Any(path string, handler http.HandlerFunc) {
	a.registerRoute(router.ANY, path, handler)
}
