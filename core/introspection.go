package core

import (
	"fmt"
	"strings"
)

// RuntimeSnapshot returns the stable runtime/config snapshot consumed by
// first-party tooling.
func (a *App) RuntimeSnapshot() RuntimeSnapshot {
	if a == nil || a.config == nil {
		return RuntimeSnapshot{}
	}

	a.mu.RLock()
	cfg := *a.config
	started := a.started
	configFrozen := a.configFrozen
	serverPrepared := a.httpServer != nil
	a.mu.RUnlock()

	snapshot := projectRuntimeSnapshot(cfg)
	snapshot.Started = started
	snapshot.ConfigFrozen = configFrozen
	snapshot.ServerPrepared = serverPrepared
	return snapshot
}

// MiddlewareNames returns the registered middleware type names.
func (a *App) MiddlewareNames() []string {
	if a == nil || a.middlewareChain == nil {
		return nil
	}

	middlewares := a.middlewareChain.Snapshot()
	list := make([]string, 0, len(middlewares))
	for _, mw := range middlewares {
		name := fmt.Sprintf("%T", mw)
		name = strings.TrimPrefix(name, "*")
		list = append(list, name)
	}
	return list
}
