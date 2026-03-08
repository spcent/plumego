package router

import (
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
)

// Get registers a GET route with the given path and handler.
func (r *Router) Get(path string, handler Handler) error { return r.AddRoute(GET, path, handler) }

// Post registers a POST route with the given path and handler.
func (r *Router) Post(path string, handler Handler) error { return r.AddRoute(POST, path, handler) }

// Put registers a PUT route with the given path and handler.
func (r *Router) Put(path string, handler Handler) error { return r.AddRoute(PUT, path, handler) }

// Delete registers a DELETE route with the given path and handler.
func (r *Router) Delete(path string, handler Handler) error { return r.AddRoute(DELETE, path, handler) }

// Patch registers a PATCH route with the given path and handler.
func (r *Router) Patch(path string, handler Handler) error { return r.AddRoute(PATCH, path, handler) }

// Any registers a route that accepts any HTTP method with the given path and handler.
func (r *Router) Any(path string, handler Handler) error { return r.AddRoute(ANY, path, handler) }

// Options registers an OPTIONS route with the given path and handler.
func (r *Router) Options(path string, handler Handler) error {
	return r.AddRoute(OPTIONS, path, handler)
}

// Head registers a HEAD route with the given path and handler.
func (r *Router) Head(path string, handler Handler) error { return r.AddRoute(HEAD, path, handler) }

// GetCtx registers a GET route with a context-aware handler.
func (r *Router) GetCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(GET, path, handler)
}

// PostCtx registers a POST route with a context-aware handler.
func (r *Router) PostCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(POST, path, handler)
}

// PutCtx registers a PUT route with a context-aware handler.
func (r *Router) PutCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(PUT, path, handler)
}

// DeleteCtx registers a DELETE route with a context-aware handler.
func (r *Router) DeleteCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(DELETE, path, handler)
}

// PatchCtx registers a PATCH route with a context-aware handler.
func (r *Router) PatchCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(PATCH, path, handler)
}

// AnyCtx registers a route that accepts any HTTP method with a context-aware handler.
func (r *Router) AnyCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(ANY, path, handler)
}

// GetFunc registers a GET route with a standard http.HandlerFunc.
func (r *Router) GetFunc(path string, h http.HandlerFunc) error { return r.HandleFunc(GET, path, h) }

// PostFunc registers a POST route with a standard http.HandlerFunc.
func (r *Router) PostFunc(path string, h http.HandlerFunc) error { return r.HandleFunc(POST, path, h) }

// PutFunc registers a PUT route with a standard http.HandlerFunc.
func (r *Router) PutFunc(path string, h http.HandlerFunc) error { return r.HandleFunc(PUT, path, h) }

// DeleteFunc registers a DELETE route with a standard http.HandlerFunc.
func (r *Router) DeleteFunc(path string, h http.HandlerFunc) error {
	return r.HandleFunc(DELETE, path, h)
}

// PatchFunc registers a PATCH route with a standard http.HandlerFunc.
func (r *Router) PatchFunc(path string, h http.HandlerFunc) error {
	return r.HandleFunc(PATCH, path, h)
}

// AnyFunc registers a route for any HTTP method with a standard http.HandlerFunc.
func (r *Router) AnyFunc(path string, h http.HandlerFunc) error { return r.HandleFunc(ANY, path, h) }

// Resource registers REST-style routes for a resource.
func (r *Router) Resource(path string, c ResourceController) error {
	path = strings.TrimSuffix(path, "/")

	if err := r.Get(path, http.HandlerFunc(c.Index)); err != nil {
		return err
	}
	if err := r.Post(path, http.HandlerFunc(c.Create)); err != nil {
		return err
	}
	if err := r.Get(path+"/:id", http.HandlerFunc(c.Show)); err != nil {
		return err
	}
	if err := r.Put(path+"/:id", http.HandlerFunc(c.Update)); err != nil {
		return err
	}
	if err := r.Delete(path+"/:id", http.HandlerFunc(c.Delete)); err != nil {
		return err
	}
	if err := r.Patch(path+"/:id", http.HandlerFunc(c.Patch)); err != nil {
		return err
	}
	if err := r.Options(path, http.HandlerFunc(c.Options)); err != nil {
		return err
	}
	if err := r.Options(path+"/:id", http.HandlerFunc(c.Options)); err != nil {
		return err
	}
	if err := r.Head(path, http.HandlerFunc(c.Head)); err != nil {
		return err
	}
	if err := r.Head(path+"/:id", http.HandlerFunc(c.Head)); err != nil {
		return err
	}
	return nil
}
