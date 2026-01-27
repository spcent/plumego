package middleware

// Registry collects middleware functions in registration order.
//
// Components can use the Registry to contribute middleware to the
// application pipeline without needing to know how the chain is
// assembled. This is useful for modular applications where different
// components need to add their own middleware.
//
// Registry is not goroutine-safe; register middleware during startup
// before the HTTP server begins serving requests.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	registry := middleware.NewRegistry()
//
//	// Add middleware in registration order
//	registry.Use(middleware.Logging)
//	registry.Use(middleware.Recovery)
//	registry.Use(middleware.CORS)
//
//	// Prepend middleware (executes first)
//	registry.Prepend(middleware.SecurityHeaders)
//
//	// Get all middleware
//	middlewares := registry.Middlewares()
//
//	// Apply to handler
//	handler := middleware.Apply(myHandler, middlewares...)
//
// Middleware execution order:
//   - Prepend() adds middleware to the beginning (executes first)
//   - Use() adds middleware to the end (executes last)
//   - When applied, middlewares execute in reverse order (last added runs first)
type Registry struct {
	middlewares []Middleware
}

// NewRegistry creates an empty middleware registry.
func NewRegistry() *Registry {
	return &Registry{middlewares: make([]Middleware, 0)}
}

// Use appends middleware to the end of the registry.
// The middleware will execute after any previously registered middleware.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	registry := middleware.NewRegistry()
//	registry.Use(middleware.Logging)
//	registry.Use(middleware.Recovery)
func (r *Registry) Use(middlewares ...Middleware) {
	if r == nil {
		return
	}

	r.middlewares = append(r.middlewares, middlewares...)
}

// Prepend inserts middleware at the start of the registry.
// The provided middleware will execute before any previously
// registered middleware.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	registry := middleware.NewRegistry()
//	registry.Use(middleware.Logging)
//	registry.Prepend(middleware.SecurityHeaders) // Executes first
func (r *Registry) Prepend(middlewares ...Middleware) {
	if r == nil || len(middlewares) == 0 {
		return
	}

	stack := make([]Middleware, 0, len(middlewares)+len(r.middlewares))
	stack = append(stack, middlewares...)
	stack = append(stack, r.middlewares...)

	r.middlewares = stack
}

// Middlewares returns a copy of the registered middleware slice to
// prevent callers from mutating internal state.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	registry := middleware.NewRegistry()
//	registry.Use(middleware.Logging)
//	registry.Use(middleware.Recovery)
//
//	middlewares := registry.Middlewares()
//	handler := middleware.Apply(myHandler, middlewares...)
func (r *Registry) Middlewares() []Middleware {
	if r == nil {
		return nil
	}

	out := make([]Middleware, len(r.middlewares))
	copy(out, r.middlewares)
	return out
}
