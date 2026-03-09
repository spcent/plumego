package core

import (
	"github.com/spcent/plumego/middleware"
)

// Use adds middleware to the application's middleware chain.
func (a *App) Use(middlewares ...middleware.Middleware) error {
	if err := a.ensureMutable("use_middleware", "add middleware"); err != nil {
		return err
	}

	reg := a.ensureMiddlewareRegistry()
	reg.Use(middlewares...)
	return nil
}

// buildHandler builds the combined handler with current middleware stack.
func (a *App) buildHandler() {
	reg := a.ensureMiddlewareRegistry()
	r := a.ensureRouter()
	handler := middleware.Apply(r, reg.Middlewares()...)

	a.mu.Lock()
	a.handler = handler
	a.mu.Unlock()
}
