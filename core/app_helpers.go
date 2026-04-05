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
		return contract.WrapError(fmt.Errorf("cannot %s after app has started", action), operation, "core", nil)
	}
	if frozen {
		return contract.WrapError(fmt.Errorf("cannot %s after app has been initialized", action), operation, "core", nil)
	}
	return nil
}

func (a *App) freezeConfig() {
	a.mu.Lock()
	a.configFrozen = true
	a.mu.Unlock()
}

func (a *App) ensureRouter() *router.Router {
	a.mu.Lock()
	if a.router == nil {
		a.router = router.NewRouter()
	}
	r := a.router
	logger := a.logger
	hasMethodNotAllowed := a.hasRouterMethodNotAllowed
	methodNotAllowed := a.routerMethodNotAllowed
	a.mu.Unlock()

	a.syncRouterConfig(r, logger, hasMethodNotAllowed, methodNotAllowed)
	return r
}

func (a *App) syncRouterConfig(r *router.Router, logger log.StructuredLogger, hasMethodNotAllowed bool, methodNotAllowed bool) {
	if r == nil {
		return
	}
	if logger != nil {
		r.SetLogger(logger)
	}
	if hasMethodNotAllowed {
		r.SetMethodNotAllowed(methodNotAllowed)
	}
}

func (a *App) ensureMiddlewareChain() *middleware.Chain {
	a.mu.Lock()
	if a.middlewareChain == nil {
		a.middlewareChain = middleware.NewChain()
	}
	chain := a.middlewareChain
	a.mu.Unlock()
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
