package core

import (
	"fmt"
	"strings"

	"github.com/spcent/plumego/core/components/devtools"
	"github.com/spcent/plumego/metrics"
)

func newDevToolsComponent(app *App) *devtools.DevToolsComponent {
	app.mu.RLock()
	debug := app.config.Debug
	envFile := app.config.EnvFile
	logger := app.logger
	app.mu.RUnlock()

	return devtools.NewComponent(devtools.Options{
		Debug:   debug,
		Logger:  logger,
		EnvFile: envFile,
		Hooks: devtools.Hooks{
			ConfigSnapshot: func() map[string]any {
				return devtoolsConfigSnapshot(app)
			},
			MiddlewareList: func() []string {
				return devtoolsMiddlewareList(app)
			},
			AttachDevMetrics: func(dev *metrics.DevCollector) {
				attachDevMetrics(app, dev)
			},
		},
	})
}

func attachDevMetrics(app *App, dev *metrics.DevCollector) {
	if app == nil || dev == nil {
		return
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	if app.httpMetrics == nil {
		app.httpMetrics = dev
		return
	}

	app.httpMetrics = metrics.NewMultiHTTPObserver(app.httpMetrics, dev)
}

func devtoolsMiddlewareList(app *App) []string {
	if app == nil || app.middlewareReg == nil {
		return nil
	}

	middlewares := app.middlewareReg.Middlewares()
	list := make([]string, 0, len(middlewares))
	for _, mw := range middlewares {
		name := fmt.Sprintf("%T", mw)
		name = strings.TrimPrefix(name, "*")
		list = append(list, name)
	}
	return list
}

func devtoolsConfigSnapshot(app *App) map[string]any {
	if app == nil {
		return map[string]any{"debug": false}
	}

	app.mu.RLock()
	cfg := *app.config
	app.mu.RUnlock()

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
