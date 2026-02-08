package middleware

import (
	"net/http"
)

// Middleware defines a function that wraps an http.Handler to add functionality.
// This is the primary middleware type, compatible with standard http.Handler.
//
// Example:
//
//	func LoggingMiddleware(next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			log.Printf("Request: %s %s", r.Method, r.URL.Path)
//			next.ServeHTTP(w, r)
//		})
//	}
//
// Usage:
//
//	handler := middleware.Apply(myHandler, LoggingMiddleware)
type Middleware func(http.Handler) http.Handler

// Deprecated: UseFuncMiddleware is deprecated, use FromFuncMiddleware instead.
// This is kept for backward compatibility with existing tests.
var Use = func(middleware Middleware) {
	// This is a no-op for backward compatibility
}

// Deprecated: ApplyGlobal is deprecated, use Apply instead.
// This is kept for backward compatibility with existing tests.
func ApplyGlobal(h http.HandlerFunc) http.HandlerFunc {
	return ApplyFunc(h)
}

// Chain represents a chain of middlewares that can be applied to a handler.
// Middlewares are applied in the order they are added, but executed in reverse order
// (last added runs first).
//
// Example:
//
//	chain := middleware.NewChain().
//		Use(observability.Logging(log.NewGLogger(), nil, nil)).
//		Use(recovery.RecoveryMiddleware).
//		Use(cors.CORS)
//	handler := chain.Apply(myHandler)
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain.
// The initial middlewares are copied into a slice with extra capacity
// to reduce allocations when additional middlewares are added via Use.
func NewChain(middlewares ...Middleware) *Chain {
	mws := make([]Middleware, len(middlewares), len(middlewares)+4)
	copy(mws, middlewares)
	return &Chain{
		middlewares: mws,
	}
}

// Use adds a middleware to the chain.
// Returns the chain for method chaining.
func (c *Chain) Use(middleware Middleware) *Chain {
	c.middlewares = append(c.middlewares, middleware)
	return c
}

// Len returns the number of middlewares in the chain.
func (c *Chain) Len() int {
	return len(c.middlewares)
}

// Apply applies the middleware chain to an http.Handler.
// The middlewares are applied in reverse order so they execute in
// registration order (first added wraps outermost).
func (c *Chain) Apply(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

// ApplyFunc applies the middleware chain to a http.HandlerFunc.
// The middlewares are applied in reverse order (last added runs first).
func (c *Chain) ApplyFunc(h http.HandlerFunc) http.HandlerFunc {
	wrapped := c.Apply(h)
	// Avoid an extra closure when the result is already an http.HandlerFunc.
	if hf, ok := wrapped.(http.HandlerFunc); ok {
		return hf
	}
	return wrapped.ServeHTTP
}

// FromFuncMiddleware converts a func(http.HandlerFunc) http.HandlerFunc to Middleware.
// This function is provided for convenience when working with function handlers.
//
// Example:
//
//	func myFuncMiddleware(next http.HandlerFunc) http.HandlerFunc {
//		return func(w http.ResponseWriter, r *http.Request) {
//			// middleware logic
//			next(w, r)
//		}
//	}
//
//	middleware := middleware.FromFuncMiddleware(myFuncMiddleware)
func FromFuncMiddleware(fm func(http.HandlerFunc) http.HandlerFunc) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fm(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})(w, r)
		})
	}
}

// Apply applies middlewares to an http.Handler.
// Middlewares are applied in the order they are provided.
//
// Example:
//
//	handler := middleware.Apply(myHandler, LoggingMiddleware, RecoveryMiddleware, CORS)
func Apply(h http.Handler, middlewares ...Middleware) http.Handler {
	return NewChain(middlewares...).Apply(h)
}

// ApplyFunc applies middlewares to a http.HandlerFunc.
// Middlewares are applied in the order they are provided.
//
// Example:
//
//	handler := middleware.ApplyFunc(myHandlerFunc, LoggingMiddleware, RecoveryMiddleware, CORS)
func ApplyFunc(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	return NewChain(middlewares...).ApplyFunc(h)
}
