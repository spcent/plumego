package middleware

import (
	"net/http"
)

// Middleware defines a function that wraps an http.Handler to add functionality
// This is the primary middleware type, compatible with standard http.Handler
type Middleware func(http.Handler) http.Handler

// Deprecated: UseFuncMiddleware is deprecated, use FromFuncMiddleware instead
// This is kept for backward compatibility with existing tests
var Use = func(middleware Middleware) {
	// This is a no-op for backward compatibility
}

// Deprecated: ApplyGlobal is deprecated, use Apply instead
// This is kept for backward compatibility with existing tests
func ApplyGlobal(h http.HandlerFunc) http.HandlerFunc {
	return ApplyFunc(h)
}

// Chain represents a chain of middlewares that can be applied to a handler
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
	}
}

// Use adds a middleware to the chain
func (c *Chain) Use(middleware Middleware) *Chain {
	c.middlewares = append(c.middlewares, middleware)
	return c
}

// Apply applies the middleware chain to an http.Handler
func (c *Chain) Apply(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

// ApplyFunc applies the middleware chain to a http.HandlerFunc
func (c *Chain) ApplyFunc(h http.HandlerFunc) http.HandlerFunc {
	wrapped := c.Apply(h)
	return func(w http.ResponseWriter, r *http.Request) {
		wrapped.ServeHTTP(w, r)
	}
}

// FromFuncMiddleware converts a func(http.HandlerFunc) http.HandlerFunc to Middleware
// This function is provided for convenience when working with function handlers
func FromFuncMiddleware(fm func(http.HandlerFunc) http.HandlerFunc) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fm(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})(w, r)
		})
	}
}

// Apply applies middlewares to an http.Handler
func Apply(h http.Handler, middlewares ...Middleware) http.Handler {
	return NewChain(middlewares...).Apply(h)
}

// ApplyFunc applies middlewares to a http.HandlerFunc
func ApplyFunc(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	return NewChain(middlewares...).ApplyFunc(h)
}
