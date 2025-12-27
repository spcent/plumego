package middleware

// Registry collects middleware functions in registration order.
//
// Components can use the Registry to contribute middleware to the
// application pipeline without needing to know how the chain is
// assembled.
type Registry struct {
	middlewares []Middleware
}

// NewRegistry creates an empty middleware registry.
func NewRegistry() *Registry {
	return &Registry{middlewares: make([]Middleware, 0)}
}

// Use appends middleware to the end of the registry.
func (r *Registry) Use(middlewares ...Middleware) {
	if r == nil {
		return
	}

	r.middlewares = append(r.middlewares, middlewares...)
}

// Prepend inserts middleware at the start of the registry.
// The provided middleware will execute before any previously
// registered middleware.
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
func (r *Registry) Middlewares() []Middleware {
	if r == nil {
		return nil
	}

	out := make([]Middleware, len(r.middlewares))
	copy(out, r.middlewares)
	return out
}
