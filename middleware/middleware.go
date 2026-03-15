// Package middleware provides the [Chain] type for composing [http.Handler] middleware.
//
// # Ordering semantics
//
// Registration order is the order middlewares are added via [NewChain], [Chain.Use],
// or [Chain.Prepend]. Execution order matches registration order: the first registered
// middleware runs outermost — it executes first on the request path and last on the
// response path. [Chain.Prepend] inserts at the front so prepended middlewares always
// run before anything added via [Chain.Use].
//
// # Canonical example
//
//	chain := middleware.NewChain(
//	    security.SecurityHeaders(nil),          // [1] outermost — runs first on request
//	    observability.Logging(logger, nil, nil), // [2]
//	    recovery.RecoveryMiddleware,             // [3] innermost before handler
//	)
//	handler := chain.Apply(myHandler)
//
// Request path:  security → logging → recovery → handler
// Response path: handler → recovery → logging → security
//
// # Immutable output
//
// [Chain.Apply] and [Chain.ApplyFunc] snapshot the registered middleware slice at call
// time. The returned handler does not hold a reference back to the [Chain], so
// subsequent [Chain.Use] or [Chain.Prepend] calls do not affect already-built handlers.
// Use [Chain.Snapshot] when you need a copy of the registered slice before building.
package middleware

import (
	"net/http"
)

// Middleware is a function that wraps an [http.Handler] to add transport-level
// functionality. Middleware must address only cross-cutting concerns (auth checks,
// logging, rate limiting, etc.) and must not contain business logic.
//
// Example:
//
//	func LoggingMiddleware(next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			log.Printf("%s %s", r.Method, r.URL.Path)
//			next.ServeHTTP(w, r)
//		})
//	}
type Middleware func(http.Handler) http.Handler

// Chain is a mutable, ordered list of [Middleware] values that can be applied to an
// [http.Handler]. Build the chain during startup, then call [Chain.Apply] once to
// produce the final handler. The chain is not goroutine-safe; register all middleware
// before the HTTP server starts serving requests.
//
// Example:
//
//	chain := middleware.NewChain(
//	    security.SecurityHeaders(nil),
//	    observability.Logging(logger, nil, nil),
//	    recovery.RecoveryMiddleware,
//	)
//	http.ListenAndServe(":8080", chain.Apply(mux))
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a Chain pre-loaded with the given middlewares in the order provided.
func NewChain(middlewares ...Middleware) *Chain {
	mws := make([]Middleware, len(middlewares), len(middlewares)+4)
	copy(mws, middlewares)
	return &Chain{middlewares: mws}
}

// Use appends one or more middlewares to the end of the chain.
// Appended middlewares execute after all previously registered middlewares.
// Returns the chain to allow method chaining.
func (c *Chain) Use(middlewares ...Middleware) *Chain {
	c.middlewares = append(c.middlewares, middlewares...)
	return c
}

// Prepend inserts one or more middlewares at the front of the chain.
// Prepended middlewares execute before any previously registered middlewares.
// When multiple middlewares are prepended at once, they are inserted together
// preserving their relative order at the front.
// Returns the chain to allow method chaining.
func (c *Chain) Prepend(middlewares ...Middleware) *Chain {
	if len(middlewares) == 0 {
		return c
	}
	stack := make([]Middleware, 0, len(middlewares)+len(c.middlewares))
	stack = append(stack, middlewares...)
	stack = append(stack, c.middlewares...)
	c.middlewares = stack
	return c
}

// Len returns the number of middlewares currently in the chain.
func (c *Chain) Len() int {
	return len(c.middlewares)
}

// Snapshot returns an independent copy of the current middleware slice.
// The returned slice is not affected by subsequent [Chain.Use] or [Chain.Prepend]
// calls. Use this when you need to inspect or pass the registered list elsewhere
// before calling [Chain.Apply].
func (c *Chain) Snapshot() []Middleware {
	out := make([]Middleware, len(c.middlewares))
	copy(out, c.middlewares)
	return out
}

// Apply wraps h with all registered middlewares and returns the resulting
// [http.Handler]. Middlewares execute in registration order (first registered =
// outermost). The returned handler does not reference the Chain; later [Chain.Use]
// or [Chain.Prepend] calls do not affect it.
func (c *Chain) Apply(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

// ApplyFunc wraps h with all registered middlewares and returns an [http.HandlerFunc].
func (c *Chain) ApplyFunc(h http.HandlerFunc) http.HandlerFunc {
	wrapped := c.Apply(h)
	if hf, ok := wrapped.(http.HandlerFunc); ok {
		return hf
	}
	return wrapped.ServeHTTP
}

// Apply applies middlewares to an [http.Handler] without constructing a Chain.
// Middlewares execute in the order provided (first = outermost).
//
// Example:
//
//	handler := middleware.Apply(myHandler, LoggingMiddleware, RecoveryMiddleware)
func Apply(h http.Handler, middlewares ...Middleware) http.Handler {
	return NewChain(middlewares...).Apply(h)
}

// ApplyFunc applies middlewares to an [http.HandlerFunc].
// Middlewares execute in the order provided (first = outermost).
func ApplyFunc(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	return NewChain(middlewares...).ApplyFunc(h)
}

// FromFuncMiddleware converts a func(http.HandlerFunc) http.HandlerFunc to [Middleware].
//
// Example:
//
//	mw := middleware.FromFuncMiddleware(func(next http.HandlerFunc) http.HandlerFunc {
//		return func(w http.ResponseWriter, r *http.Request) {
//			// pre-handler logic
//			next(w, r)
//		}
//	})
func FromFuncMiddleware(fm func(http.HandlerFunc) http.HandlerFunc) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fm(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})(w, r)
		})
	}
}
