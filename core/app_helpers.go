package core

import (
	"fmt"

	"github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
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
	a.mu.Unlock()

	if r != nil && logger != nil {
		r.SetLogger(logger)
	}
	return r
}

func (a *App) ensureMiddlewareRegistry() *middleware.Registry {
	a.mu.Lock()
	if a.middlewareReg == nil {
		a.middlewareReg = middleware.NewRegistry()
	}
	reg := a.middlewareReg
	a.mu.Unlock()
	return reg
}

func (a *App) configSnapshot() AppConfig {
	a.mu.RLock()
	if a.config == nil {
		a.mu.RUnlock()
		return AppConfig{}
	}
	cfg := *a.config
	a.mu.RUnlock()
	return cfg
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
