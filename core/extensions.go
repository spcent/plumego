package core

import (
	"fmt"
	"strings"

	"github.com/spcent/plumego/metrics"
)

// MountComponent appends a component before the app is prepared or started.
func (a *App) MountComponent(component Component) error {
	if component == nil {
		return nil
	}
	if err := a.ensureMutable("mount_component", "mount component"); err != nil {
		return err
	}

	a.mu.Lock()
	a.components = append(a.components, component)
	a.mu.Unlock()
	return nil
}

// MountComponents appends multiple components before the app is prepared or started.
func (a *App) MountComponents(components ...Component) error {
	for _, component := range components {
		if err := a.MountComponent(component); err != nil {
			return err
		}
	}
	return nil
}

// DebugEnabled reports whether debug mode is enabled.
func (a *App) DebugEnabled() bool {
	if a == nil || a.config == nil {
		return false
	}

	a.mu.RLock()
	debug := a.config.Debug
	a.mu.RUnlock()
	return debug
}

// EnvPath returns the configured env file path.
func (a *App) EnvPath() string {
	if a == nil || a.config == nil {
		return ""
	}

	a.mu.RLock()
	path := a.config.EnvFile
	a.mu.RUnlock()
	return path
}

// ConfigSnapshot exposes the current server-facing config values.
func (a *App) ConfigSnapshot() map[string]any {
	if a == nil || a.config == nil {
		return map[string]any{"debug": false}
	}

	a.mu.RLock()
	cfg := *a.config
	a.mu.RUnlock()

	return map[string]any{
		"addr":                cfg.Addr,
		"debug":               cfg.Debug,
		"env_file":            cfg.EnvFile,
		"read_timeout":        cfg.ReadTimeout,
		"read_header_timeout": cfg.ReadHeaderTimeout,
		"write_timeout":       cfg.WriteTimeout,
		"idle_timeout":        cfg.IdleTimeout,
		"max_header_bytes":    cfg.MaxHeaderBytes,
		"http2_enabled":       cfg.EnableHTTP2,
		"tls": map[string]any{
			"enabled":   cfg.TLS.Enabled,
			"cert_file": cfg.TLS.CertFile,
			"key_file":  cfg.TLS.KeyFile,
		},
	}
}

// MiddlewareNames returns the registered middleware type names.
func (a *App) MiddlewareNames() []string {
	if a == nil || a.middlewareReg == nil {
		return nil
	}

	middlewares := a.middlewareReg.Middlewares()
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
