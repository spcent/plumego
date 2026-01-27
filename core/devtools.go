package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/spcent/plumego/config"
	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

const (
	devToolsBasePath       = "/_debug"
	devToolsRoutesPath     = devToolsBasePath + "/routes"
	devToolsRoutesJSONPath = devToolsBasePath + "/routes.json"
	devToolsMiddlewarePath = devToolsBasePath + "/middleware"
	devToolsConfigPath     = devToolsBasePath + "/config"
	devToolsReloadPath     = devToolsBasePath + "/reload"
)

type devToolsComponent struct {
	BaseComponent
	app     *App
	debug   bool
	logger  log.StructuredLogger
	envFile string

	watchOnce sync.Once
	watchStop context.CancelFunc
	watchWg   sync.WaitGroup
}

func newDevToolsComponent(app *App) *devToolsComponent {
	app.mu.RLock()
	debug := app.config.Debug
	envFile := app.config.EnvFile
	logger := app.logger
	app.mu.RUnlock()

	return &devToolsComponent{
		app:     app,
		debug:   debug,
		logger:  logger,
		envFile: envFile,
	}
}

func (c *devToolsComponent) RegisterRoutes(r *router.Router) {
	if !c.debug {
		return
	}

	r.Get(devToolsRoutesPath, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		r.Print(w)
	}))

	r.Get(devToolsRoutesJSONPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]any{
			"routes": r.Routes(),
		}
		writeHTTPResponse(w, req, http.StatusOK, payload)
	}))

	r.Get(devToolsMiddlewarePath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]any{
			"middlewares": c.middlewareList(),
		}
		writeHTTPResponse(w, req, http.StatusOK, payload)
	}))

	r.Get(devToolsConfigPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		writeHTTPResponse(w, req, http.StatusOK, c.configSnapshot())
	}))

	r.Post(devToolsReloadPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := c.reloadEnv(req.Context()); err != nil {
			contract.WriteError(w, req, contract.APIError{
				Status:   http.StatusBadRequest,
				Code:     "env_reload_failed",
				Category: contract.CategoryClient,
				Message:  err.Error(),
			})
			return
		}

		writeHTTPResponse(w, req, http.StatusOK, map[string]any{
			"status": "ok",
		})
	}))
}

func (c *devToolsComponent) RegisterMiddleware(_ *middleware.Registry) {}

func (c *devToolsComponent) Start(ctx context.Context) error {
	if !c.debug || c.envFile == "" {
		return nil
	}

	if _, err := os.Stat(c.envFile); err != nil {
		return nil
	}

	c.watchOnce.Do(func() {
		watchCtx, cancel := context.WithCancel(ctx)
		c.watchStop = cancel
		c.watchWg.Add(1)
		go c.watchEnvFile(watchCtx)
	})

	return nil
}

func (c *devToolsComponent) Stop(_ context.Context) error {
	if c.watchStop != nil {
		c.watchStop()
	}
	c.watchWg.Wait()
	return nil
}

func (c *devToolsComponent) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"enabled": c.debug}}
	if !c.debug {
		status.Status = health.StatusDegraded
		status.Message = "devtools disabled"
	}
	return "devtools", status
}

func (c *devToolsComponent) reloadEnv(ctx context.Context) error {
	_ = ctx
	if c.envFile == "" {
		return fmt.Errorf("env file not configured")
	}

	if _, err := os.Stat(c.envFile); err != nil {
		return fmt.Errorf("env file not found")
	}

	if err := config.LoadEnv(c.envFile, true); err != nil {
		return err
	}

	c.logger.Info("Reloaded .env file", log.Fields{"path": c.envFile})
	return nil
}

func (c *devToolsComponent) watchEnvFile(ctx context.Context) {
	defer c.watchWg.Done()

	fileSource := config.NewFileSource(c.envFile, config.FormatEnv, true)
	if _, err := fileSource.Load(ctx); err != nil {
		c.logger.Warn("Devtools env watch load failed", log.Fields{"error": err})
	}

	updates, errs := fileSource.Watch(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-errs:
			if !ok {
				return
			}
			if err != nil {
				c.logger.Warn("Devtools env watch error", log.Fields{"error": err})
			}
		case _, ok := <-updates:
			if !ok {
				return
			}
			if err := c.reloadEnv(ctx); err != nil {
				c.logger.Warn("Devtools env reload failed", log.Fields{"error": err})
			}
		}
	}
}

func (c *devToolsComponent) middlewareList() []string {
	if c.app == nil || c.app.middlewareReg == nil {
		return nil
	}

	middlewares := c.app.middlewareReg.Middlewares()
	list := make([]string, 0, len(middlewares))
	for _, mw := range middlewares {
		list = append(list, formatType(mw))
	}
	return list
}

func (c *devToolsComponent) configSnapshot() map[string]any {
	if c.app == nil {
		return map[string]any{"debug": c.debug}
	}

	c.app.mu.RLock()
	cfg := *c.app.config
	c.app.mu.RUnlock()

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
