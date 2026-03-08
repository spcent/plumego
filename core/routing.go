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

func (a *App) registerCtxRoute(method, path string, handler contract.CtxHandlerFunc) {
	if err := contract.ValidateCtxHandler(handler); err != nil {
		a.logError("route registration failed", err, log.Fields{"method": method, "path": path})
		return
	}

	a.mu.RLock()
	logger := a.logger
	a.mu.RUnlock()

	a.registerRoute(method, path, contract.AdaptCtxHandler(handler, logger))
}

// =========================================================
// Canonical handler registration — standard library style
//
// These are the primary route registration methods. Use
// http.HandlerFunc as the handler type for all new routes.
// This is the only style shown in quick-start guides and
// canonical examples.
// =========================================================

// HandleFunc registers a handler function for the given path (any HTTP method).
func (a *App) HandleFunc(pattern string, handler http.HandlerFunc) {
	a.registerRoute(router.ANY, pattern, handler)
}

// Handle registers a handler for the given path (any HTTP method).
func (a *App) Handle(pattern string, handler http.Handler) {
	a.registerRoute(router.ANY, pattern, handler)
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

// =========================================================
// Compatibility adapters — context-aware style
//
// These methods accept contract.CtxHandlerFunc and adapt it
// to the canonical http.Handler pipeline internally. They are
// provided for convenience but are not the recommended primary
// style for new code. Prefer the standard Get/Post/... family.
//
// Do not introduce additional *Ctx registration variants.
// =========================================================

// GetCtx registers a GET route with a context-aware handler.
//
// Deprecated: use Get with http.HandlerFunc for new code.
func (a *App) GetCtx(path string, handler contract.CtxHandlerFunc) {
	a.registerCtxRoute(router.GET, path, handler)
}

// PostCtx registers a POST route with a context-aware handler.
//
// Deprecated: use Post with http.HandlerFunc for new code.
func (a *App) PostCtx(path string, handler contract.CtxHandlerFunc) {
	a.registerCtxRoute(router.POST, path, handler)
}

// PutCtx registers a PUT route with a context-aware handler.
//
// Deprecated: use Put with http.HandlerFunc for new code.
func (a *App) PutCtx(path string, handler contract.CtxHandlerFunc) {
	a.registerCtxRoute(router.PUT, path, handler)
}

// DeleteCtx registers a DELETE route with a context-aware handler.
//
// Deprecated: use Delete with http.HandlerFunc for new code.
func (a *App) DeleteCtx(path string, handler contract.CtxHandlerFunc) {
	a.registerCtxRoute(router.DELETE, path, handler)
}

// PatchCtx registers a PATCH route with a context-aware handler.
//
// Deprecated: use Patch with http.HandlerFunc for new code.
func (a *App) PatchCtx(path string, handler contract.CtxHandlerFunc) {
	a.registerCtxRoute(router.PATCH, path, handler)
}

// AnyCtx registers a route for any HTTP method with a context-aware handler.
//
// Deprecated: use Any with http.HandlerFunc for new code.
func (a *App) AnyCtx(path string, handler contract.CtxHandlerFunc) {
	a.registerCtxRoute(router.ANY, path, handler)
}

// =========================================================
// Compatibility adapters — router.Handler style
//
// router.Handler is an alias for http.Handler. These methods
// exist for legacy call sites that explicitly reference the
// router.Handler type. Prefer the standard Get/Post/... family
// with http.HandlerFunc for all new code.
//
// Do not introduce additional *Handler registration variants.
// =========================================================

// GetHandler registers a GET route with an http.Handler.
//
// Deprecated: use Get with http.HandlerFunc for new code.
func (a *App) GetHandler(path string, handler router.Handler) {
	a.registerRoute(router.GET, path, handler)
}

// PostHandler registers a POST route with an http.Handler.
//
// Deprecated: use Post with http.HandlerFunc for new code.
func (a *App) PostHandler(path string, handler router.Handler) {
	a.registerRoute(router.POST, path, handler)
}

// PutHandler registers a PUT route with an http.Handler.
//
// Deprecated: use Put with http.HandlerFunc for new code.
func (a *App) PutHandler(path string, handler router.Handler) {
	a.registerRoute(router.PUT, path, handler)
}

// DeleteHandler registers a DELETE route with an http.Handler.
//
// Deprecated: use Delete with http.HandlerFunc for new code.
func (a *App) DeleteHandler(path string, handler router.Handler) {
	a.registerRoute(router.DELETE, path, handler)
}

// PatchHandler registers a PATCH route with an http.Handler.
//
// Deprecated: use Patch with http.HandlerFunc for new code.
func (a *App) PatchHandler(path string, handler router.Handler) {
	a.registerRoute(router.PATCH, path, handler)
}

// AnyHandler registers a route for any HTTP method with an http.Handler.
//
// Deprecated: use Any with http.HandlerFunc for new code.
func (a *App) AnyHandler(path string, handler router.Handler) {
	a.registerRoute(router.ANY, path, handler)
}
