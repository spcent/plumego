package core

import (
	"github.com/spcent/plumego/middleware"
)

// Use adds middleware to the application's middleware chain.
func (a *App) Use(middlewares ...middleware.Middleware) error {
	if err := a.ensureMutable("use_middleware", "add middleware"); err != nil {
		return err
	}

	chain := a.ensureMiddlewareChain()
	for _, mw := range middlewares {
		chain.Use(mw)
	}
	return nil
}

// buildHandler builds the combined handler with current middleware stack.
func (a *App) buildHandler() {
	chain := a.ensureMiddlewareChain()
	r := a.ensureRouter()
	handler := chain.Build(r)

	a.mu.Lock()
	a.handler = handler
	a.mu.Unlock()
}
