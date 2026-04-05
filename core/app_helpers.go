package core

import (
	"fmt"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

func (a *App) ensureMutable(operation, action string) error {
	a.mu.RLock()
	state := a.preparationState
	a.mu.RUnlock()

	if state != PreparationStateMutable {
		return wrapCoreError(fmt.Errorf("cannot %s after app has been prepared", action), operation, nil)
	}
	return nil
}

func wrapCoreError(err error, operation string, params map[string]any) error {
	return contract.WrapError(err, operation, "core", params)
}

func (a *App) freezeConfig() {
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

func (a *App) logError(message string, err error, fields log.Fields) {
	if err == nil {
		return
	}
	a.mu.RLock()
	logger := a.logger
	a.mu.RUnlock()
	if logger == nil {
		return
	}
	if fields == nil {
		fields = log.Fields{}
	}
	if _, ok := fields["error"]; !ok {
		fields["error"] = err
	}
	logger.Error(message, fields)
}
