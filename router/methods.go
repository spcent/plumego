package router

import "net/http"

// Get registers a GET route. Panics on duplicate or frozen-router errors.
func (r *Router) Get(path string, handler http.Handler) { r.mustAddRoute(GET, path, handler) }

// Post registers a POST route. Panics on duplicate or frozen-router errors.
func (r *Router) Post(path string, handler http.Handler) { r.mustAddRoute(POST, path, handler) }

// Put registers a PUT route. Panics on duplicate or frozen-router errors.
func (r *Router) Put(path string, handler http.Handler) { r.mustAddRoute(PUT, path, handler) }

// Delete registers a DELETE route. Panics on duplicate or frozen-router errors.
func (r *Router) Delete(path string, handler http.Handler) { r.mustAddRoute(DELETE, path, handler) }

// Patch registers a PATCH route. Panics on duplicate or frozen-router errors.
func (r *Router) Patch(path string, handler http.Handler) { r.mustAddRoute(PATCH, path, handler) }

// Options registers an OPTIONS route. Panics on duplicate or frozen-router errors.
func (r *Router) Options(path string, handler http.Handler) { r.mustAddRoute(OPTIONS, path, handler) }

// Head registers a HEAD route. Panics on duplicate or frozen-router errors.
func (r *Router) Head(path string, handler http.Handler) { r.mustAddRoute(HEAD, path, handler) }

// Any registers a route that accepts any HTTP method. Panics on duplicate or frozen-router errors.
func (r *Router) Any(path string, handler http.Handler) { r.mustAddRoute(ANY, path, handler) }
