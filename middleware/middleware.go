package middleware

import (
	"net/http"
)

// Handler defines the standard HTTP handler interface
// It's an alias for http.Handler to improve code readability
type Handler http.Handler

// Middleware defines a function that wraps an http.Handler to add functionality
// This is the primary middleware type, compatible with http.Handler

type Middleware func(Handler) Handler

// FuncMiddleware defines a middleware that works with http.HandlerFunc
// This type is provided for convenience when working with function handlers

type FuncMiddleware func(http.HandlerFunc) http.HandlerFunc

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

// Apply applies the middleware chain to a Handler
func (c *Chain) Apply(h Handler) Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

// ApplyFunc applies the middleware chain to a http.HandlerFunc
func (c *Chain) ApplyFunc(h http.HandlerFunc) http.HandlerFunc {
	// Convert HandlerFunc to Handler
	handler := http.HandlerFunc(h)
	// Apply chain
	wrapped := c.Apply(handler)
	// Convert back to HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		wrapped.ServeHTTP(w, r)
	}
}

// HandlerFunc converts a http.HandlerFunc to a Handler
func HandlerFunc(h http.HandlerFunc) Handler {
	return http.HandlerFunc(h)
}

// FromFuncMiddleware converts a FuncMiddleware to a Middleware
func FromFuncMiddleware(fm FuncMiddleware) Middleware {
	return func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fm(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})(w, r)
		})
	}
}

// FromHTTPHandlerMiddleware converts a func(http.Handler) http.Handler to a Middleware
// This function is useful for converting standard http.Handler middlewares to the Middleware type
func FromHTTPHandlerMiddleware(mw func(http.Handler) http.Handler) Middleware {
	return func(next Handler) Handler {
		// Convert our Handler to http.Handler, apply the middleware, then convert back
		return Handler(mw(http.Handler(next)))
	}
}

// Apply applies middlewares to a Handler
// This is a generic version that works with both Middleware and FuncMiddleware
func Apply(h any, m ...any) Handler {
	// Convert h to Handler
	var handler Handler
	switch v := h.(type) {
	case http.HandlerFunc:
		// Handle HandlerFunc type first
		handler = Handler(v)
	case http.Handler:
		// Handle http.Handler type
		handler = Handler(v)
	default:
		panic("invalid handler type")
	}

	// Apply middlewares
	for i := len(m) - 1; i >= 0; i-- {
		mw := m[i]
		switch v := mw.(type) {
		case Middleware:
			handler = v(handler)
		case FuncMiddleware:
			// Convert Handler to http.HandlerFunc, apply FuncMiddleware, then convert back
			hf := func(w http.ResponseWriter, r *http.Request) {
				handler.ServeHTTP(w, r)
			}
			hf = v(hf)
			// Convert back to http.Handler first, then to Handler
			handler = Handler(http.HandlerFunc(hf))
		default:
			panic("invalid middleware type")
		}
	}

	return handler
}

// ApplyFunc applies middlewares to a http.HandlerFunc
func ApplyFunc(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	return NewChain(middlewares...).ApplyFunc(h)
}

// ApplyFuncMiddleware applies FuncMiddleware to a http.HandlerFunc
func ApplyFuncMiddleware(h http.HandlerFunc, middlewares ...FuncMiddleware) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// ApplyLegacy applies legacy FuncMiddleware to a http.HandlerFunc
// This is for backward compatibility with existing tests
func ApplyLegacy(h http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
