package core

import (
	"fmt"
	"strings"

	"github.com/spcent/plumego/core/components/devtools"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
)

const (
	devToolsBasePath       = devtools.DevToolsBasePath
	devToolsRoutesPath     = devtools.DevToolsRoutesPath
	devToolsRoutesJSONPath = devtools.DevToolsRoutesJSONPath
	devToolsMiddlewarePath = devtools.DevToolsMiddlewarePath
	devToolsConfigPath     = devtools.DevToolsConfigPath
	devToolsMetricsPath    = devtools.DevToolsMetricsPath
	devToolsMetricsClear   = devtools.DevToolsMetricsClear
	devToolsPprofBasePath  = devtools.DevToolsPprofBasePath
	devToolsPprofIndexPath = devtools.DevToolsPprofIndexPath
	devToolsPprofCmdline   = devtools.DevToolsPprofCmdline
	devToolsPprofProfile   = devtools.DevToolsPprofProfile
	devToolsPprofSymbol    = devtools.DevToolsPprofSymbol
	devToolsPprofTrace     = devtools.DevToolsPprofTrace
	devToolsReloadPath     = devtools.DevToolsReloadPath
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

	if app.metricsCollector == nil {
		app.metricsCollector = dev
		return
	}

	if app.metricsCollector == dev {
		return
	}

	app.metricsCollector = metrics.NewMultiCollector(app.metricsCollector, dev)
}

func devtoolsMiddlewareList(app *App) []string {
	if app == nil || app.middlewareReg == nil {
		return nil
	}

	middlewares := app.middlewareReg.Middlewares()
	list := make([]string, 0, len(middlewares))
	for _, mw := range middlewares {
		list = append(list, formatType(mw))
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

	abuseConfig := cfg.AbuseGuardConfig
	if abuseConfig == nil {
		defaults := middleware.DefaultAbuseGuardConfig()
		abuseConfig = &defaults
	}

	return map[string]any{
		"addr":                cfg.Addr,
		"debug":               cfg.Debug,
		"env_file":            cfg.EnvFile,
		"read_timeout":        cfg.ReadTimeout,
		"read_header_timeout": cfg.ReadHeaderTimeout,
		"write_timeout":       cfg.WriteTimeout,
		"idle_timeout":        cfg.IdleTimeout,
		"max_header_bytes":    cfg.MaxHeaderBytes,
		"max_body_bytes":      cfg.MaxBodyBytes,
		"max_concurrency":     cfg.MaxConcurrency,
		"queue_depth":         cfg.QueueDepth,
		"queue_timeout":       cfg.QueueTimeout,
		"http2_enabled":       cfg.EnableHTTP2,
		"tls": map[string]any{
			"enabled":   cfg.TLS.Enabled,
			"cert_file": cfg.TLS.CertFile,
			"key_file":  cfg.TLS.KeyFile,
		},
		"security": map[string]any{
			"headers_enabled":      cfg.EnableSecurityHeaders,
			"policy_custom":        cfg.SecurityHeadersPolicy != nil,
			"abuse_guard_enabled":  cfg.EnableAbuseGuard,
			"abuse_guard_rate":     abuseConfig.Rate,
			"abuse_guard_capacity": abuseConfig.Capacity,
		},
		"pubsub_debug": map[string]any{
			"enabled": cfg.PubSub.Enabled,
			"path":    cfg.PubSub.Path,
		},
		"webhook_out": map[string]any{
			"enabled":           cfg.WebhookOut.Enabled,
			"base_path":         cfg.WebhookOut.BasePath,
			"trigger_token_set": cfg.WebhookOut.TriggerToken != "",
			"allow_empty_token": cfg.WebhookOut.AllowEmptyToken,
		},
		"webhook_in": map[string]any{
			"enabled":           cfg.WebhookIn.Enabled,
			"github_path":       cfg.WebhookIn.GitHubPath,
			"stripe_path":       cfg.WebhookIn.StripePath,
			"github_secret_set": cfg.WebhookIn.GitHubSecret != "",
			"stripe_secret_set": cfg.WebhookIn.StripeSecret != "",
		},
	}
}

func formatType(v any) string {
	name := fmt.Sprintf("%T", v)
	name = strings.TrimPrefix(name, "*")
	return name
}
