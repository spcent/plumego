package middleware

import "net/http"

// Middleware is the canonical middleware type used across Plumego.
//
// Usage with core.App:
//
//	app.Use(observability.RequestID())
//
// Usage with a standalone handler:
//
//	h := middleware.Apply(finalHandler, observability.RequestID(), recovery.RecoveryMiddleware)
type Middleware func(http.Handler) http.Handler

// Chain composes middleware in registration order.
//
// Given middlewares A then B, execution is: A -> B -> handler.
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a chain from the provided middlewares.
func NewChain(middlewares ...Middleware) *Chain {
	mws := make([]Middleware, len(middlewares), len(middlewares)+4)
	copy(mws, middlewares)
	return &Chain{middlewares: mws}
}

// Use appends middleware to the chain.
func (c *Chain) Use(middleware Middleware) *Chain {
	c.middlewares = append(c.middlewares, middleware)
	return c
}

// Len returns the number of middlewares in the chain.
func (c *Chain) Len() int {
	return len(c.middlewares)
}

// Apply wraps h so middleware executes in registration order.
func (c *Chain) Apply(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

// Apply wraps h with middlewares in registration order.
func Apply(h http.Handler, middlewares ...Middleware) http.Handler {
	return NewChain(middlewares...).Apply(h)
}
