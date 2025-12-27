package core

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

// HandleFunc registers a handler function for the given path.
func (a *App) HandleFunc(pattern string, handler http.HandlerFunc) {
	a.router.HandleFunc(router.ANY, pattern, handler)
}

// Handle registers a handler for the given path.
func (a *App) Handle(pattern string, handler http.Handler) {
	a.router.Handle(router.ANY, pattern, handler)
}

// Get registers a GET route with the given handler.
func (a *App) Get(path string, handler http.HandlerFunc) {
	a.Router().GetFunc(path, handler)
}

// Post registers a POST route with the given handler.
func (a *App) Post(path string, handler http.HandlerFunc) {
	a.Router().PostFunc(path, handler)
}

// Put registers a PUT route with the given handler.
func (a *App) Put(path string, handler http.HandlerFunc) {
	a.Router().PutFunc(path, handler)
}

// Delete registers a DELETE route with the given handler.
func (a *App) Delete(path string, handler http.HandlerFunc) {
	a.Router().DeleteFunc(path, handler)
}

// Patch registers a PATCH route with the given handler.
func (a *App) Patch(path string, handler http.HandlerFunc) {
	a.Router().PatchFunc(path, handler)
}

// Any registers a route for any HTTP method with the given handler.
func (a *App) Any(path string, handler http.HandlerFunc) {
	a.Router().AnyFunc(path, handler)
}

// Context-aware registration helpers keep compatibility with net/http while exposing the unified Ctx.
func (a *App) GetCtx(path string, handler contract.CtxHandlerFunc) {
	a.Router().GetCtx(path, handler)
}

func (a *App) PostCtx(path string, handler contract.CtxHandlerFunc) {
	a.Router().PostCtx(path, handler)
}

func (a *App) PutCtx(path string, handler contract.CtxHandlerFunc) {
	a.Router().PutCtx(path, handler)
}

func (a *App) DeleteCtx(path string, handler contract.CtxHandlerFunc) {
	a.Router().DeleteCtx(path, handler)
}

func (a *App) PatchCtx(path string, handler contract.CtxHandlerFunc) {
	a.Router().PatchCtx(path, handler)
}

func (a *App) AnyCtx(path string, handler contract.CtxHandlerFunc) {
	a.Router().AnyCtx(path, handler)
}

// GetHandler registers a GET route with the router's Handler type.
func (a *App) GetHandler(path string, handler router.Handler) {
	a.Router().Get(path, handler)
}

// PostHandler registers a POST route with the router's Handler type.
func (a *App) PostHandler(path string, handler router.Handler) {
	a.Router().Post(path, handler)
}

// PutHandler registers a PUT route with the router's Handler type.
func (a *App) PutHandler(path string, handler router.Handler) {
	a.Router().Put(path, handler)
}

// DeleteHandler registers a DELETE route with the router's Handler type.
func (a *App) DeleteHandler(path string, handler router.Handler) {
	a.Router().Delete(path, handler)
}

// PatchHandler registers a PATCH route with the router's Handler type.
func (a *App) PatchHandler(path string, handler router.Handler) {
	a.Router().Patch(path, handler)
}

// AnyHandler registers a route for any HTTP method with the router's Handler type.
func (a *App) AnyHandler(path string, handler router.Handler) {
	a.Router().Any(path, handler)
}
