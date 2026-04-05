package core

import (
	"fmt"
	"strings"

	"github.com/spcent/plumego/metrics"
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

	return RuntimeSnapshot{
		Addr:              cfg.Addr,
		Debug:             cfg.Debug,
		EnvFile:           cfg.EnvFile,
		ShutdownTimeout:   cfg.ShutdownTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		HTTP2Enabled:      cfg.EnableHTTP2,
		DrainInterval:     cfg.DrainInterval,
		TLS: RuntimeTLSSnapshot{
			Enabled:  cfg.TLS.Enabled,
			CertFile: cfg.TLS.CertFile,
			KeyFile:  cfg.TLS.KeyFile,
		},
		Started:        started,
		ConfigFrozen:   configFrozen,
		ServerPrepared: serverPrepared,
	}
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

// AttachHTTPObserver fans out HTTP metrics to an additional observer.
func (a *App) AttachHTTPObserver(observer metrics.HTTPObserver) {
	if a == nil || observer == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.httpMetrics == nil {
		a.httpMetrics = observer
		return
	}

	a.httpMetrics = metrics.NewMultiHTTPObserver(a.httpMetrics, observer)
}
