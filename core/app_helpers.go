package core

import (
	"fmt"

	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

func (a *App) ensureMutable(operation, action string) error {
	if a == nil {
		return nilAppError(operation, nil)
	}
	a.mu.RLock()
	state := a.preparationState
	a.mu.RUnlock()

	if state != PreparationStateMutable {
		return wrapCoreError(fmt.Errorf("cannot %s after app has been prepared", action), operation, nil)
	}
	return nil
}

func nilAppError(operation string, params map[string]any) error {
	return wrapCoreError(fmt.Errorf("app is nil"), operation, params)
}

func wrapCoreError(err error, operation string, params map[string]any) error {
	if err == nil {
		return nil
	}
	if len(params) == 0 {
		return fmt.Errorf("core %s: %w", operation, err)
	}
	return fmt.Errorf("core %s %v: %w", operation, params, err)
}

func (a *App) freezeConfig() {
	if a == nil {
		return
	}
	a.mu.Lock()
	if a.preparationState == PreparationStateMutable {
		a.preparationState = PreparationStateHandlerPrepared
	}
	a.mu.Unlock()
}

func (a *App) ensureRouter() *router.Router {
	if a == nil {
		return nil
	}

	a.mu.RLock()
	r := a.router
	a.mu.RUnlock()
	return r
}

func (a *App) syncRouterConfig(r *router.Router) {
	if a == nil || r == nil {
		return
	}

	a.mu.RLock()
	cfg := a.config
	a.mu.RUnlock()
	if cfg == nil {
		return
	}

	r.SetMethodNotAllowed(cfg.Router.MethodNotAllowed)
}

func (a *App) ensureMiddlewareChain() *middleware.Chain {
	if a == nil {
		return nil
	}

	a.mu.RLock()
	chain := a.middlewareChain
	a.mu.RUnlock()
	return chain
}
