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
	started := a.started
	frozen := a.configFrozen
	a.mu.RUnlock()

	if started {
		return wrapCoreError(fmt.Errorf("cannot %s after app has started", action), operation, nil)
	}
	if frozen {
		return wrapCoreError(fmt.Errorf("cannot %s after app has been initialized", action), operation, nil)
	}
	return nil
}

func wrapCoreError(err error, operation string, params map[string]any) error {
	return contract.WrapError(err, operation, "core", params)
}

func (a *App) freezeConfig() {
	a.mu.Lock()
	a.configFrozen = true
	a.mu.Unlock()
}

func (a *App) ensureRouter() *router.Router {
	if a == nil {
		return nil
	}

	a.mu.RLock()
	r := a.router
	hasMethodNotAllowed := a.hasRouterMethodNotAllowed
	methodNotAllowed := a.routerMethodNotAllowed
	a.mu.RUnlock()

	a.syncRouterConfig(r, hasMethodNotAllowed, methodNotAllowed)
	return r
}

func (a *App) syncRouterConfig(r *router.Router, hasMethodNotAllowed bool, methodNotAllowed bool) {
	if r == nil {
		return
	}
	if hasMethodNotAllowed {
		r.SetMethodNotAllowed(methodNotAllowed)
	}
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
