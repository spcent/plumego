package devtools

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"reflect"
	"sync"

	"github.com/spcent/plumego/config"
	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core/internal/contractio"
	"github.com/spcent/plumego/health"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

const (
	DevToolsBasePath       = "/_debug"
	DevToolsRoutesPath     = DevToolsBasePath + "/routes"
	DevToolsRoutesJSONPath = DevToolsBasePath + "/routes.json"
	DevToolsMiddlewarePath = DevToolsBasePath + "/middleware"
	DevToolsConfigPath     = DevToolsBasePath + "/config"
	DevToolsMetricsPath    = DevToolsBasePath + "/metrics"
	DevToolsMetricsClear   = DevToolsMetricsPath + "/clear"
	DevToolsPprofBasePath  = DevToolsBasePath + "/pprof"
	DevToolsPprofIndexPath = DevToolsPprofBasePath + "/"
	DevToolsPprofCmdline   = DevToolsPprofBasePath + "/cmdline"
	DevToolsPprofProfile   = DevToolsPprofBasePath + "/profile"
	DevToolsPprofSymbol    = DevToolsPprofBasePath + "/symbol"
	DevToolsPprofTrace     = DevToolsPprofBasePath + "/trace"
	DevToolsReloadPath     = DevToolsBasePath + "/reload"
)

type DevToolsComponent struct {
	debug   bool
	logger  log.StructuredLogger
	envFile string
	hooks   Hooks

	watchOnce sync.Once
	watchStop context.CancelFunc
	watchWg   sync.WaitGroup

	devMetrics *metrics.DevCollector
}

type Hooks struct {
	ConfigSnapshot   func() map[string]any
	MiddlewareList   func() []string
	AttachDevMetrics func(*metrics.DevCollector)
}

type Options struct {
	Debug   bool
	Logger  log.StructuredLogger
	EnvFile string
	Hooks   Hooks
}

func NewComponent(opts Options) *DevToolsComponent {
	if opts.Logger == nil {
		opts.Logger = log.NewGLogger()
	}
	return &DevToolsComponent{
		debug:      opts.Debug,
		logger:     opts.Logger,
		envFile:    opts.EnvFile,
		hooks:      opts.Hooks,
		devMetrics: metrics.NewDevCollector(metrics.DefaultDevCollectorConfig()),
	}
}

func (c *DevToolsComponent) RegisterRoutes(r *router.Router) {
	if !c.debug {
		return
	}

	// Legacy aliases for CLI compatibility.
	r.Get("/_routes", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]any{
			"routes": r.Routes(),
		}
		contractio.WriteHTTPResponse(w, req, http.StatusOK, payload)
	}))

	r.Get("/_config", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		contractio.WriteHTTPResponse(w, req, http.StatusOK, c.configSnapshot())
	}))

	r.Get("/_info", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]any{
			"config": c.configSnapshot(),
			"build":  health.GetBuildInfo(),
		}
		contractio.WriteHTTPResponse(w, req, http.StatusOK, payload)
	}))

	r.Get(DevToolsRoutesPath, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		r.Print(w)
	}))

	r.Get(DevToolsRoutesJSONPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]any{
			"routes": r.Routes(),
		}
		contractio.WriteHTTPResponse(w, req, http.StatusOK, payload)
	}))

	r.Get(DevToolsMiddlewarePath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]any{
			"middlewares": c.middlewareList(),
		}
		contractio.WriteHTTPResponse(w, req, http.StatusOK, payload)
	}))

	r.Get(DevToolsConfigPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		contractio.WriteHTTPResponse(w, req, http.StatusOK, c.configSnapshot())
	}))

	r.Get(DevToolsMetricsPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if c.devMetrics == nil {
			contractio.WriteHTTPResponse(w, req, http.StatusOK, map[string]any{
				"enabled": false,
			})
			return
		}

		contractio.WriteHTTPResponse(w, req, http.StatusOK, map[string]any{
			"enabled": true,
			"http":    c.devMetrics.Snapshot(),
			"db":      c.devMetrics.DBSnapshot(),
		})
	}))

	r.Post(DevToolsMetricsClear, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if c.devMetrics != nil {
			c.devMetrics.Clear()
		}
		contractio.WriteHTTPResponse(w, req, http.StatusOK, map[string]any{
			"status": "ok",
		})
	}))

	// pprof endpoints (debug-only)
	r.Get(DevToolsPprofBasePath, http.HandlerFunc(pprof.Index))
	r.Get(DevToolsPprofIndexPath, http.HandlerFunc(pprof.Index))
	r.Get(DevToolsPprofCmdline, http.HandlerFunc(pprof.Cmdline))
	r.Get(DevToolsPprofProfile, http.HandlerFunc(pprof.Profile))
	r.Get(DevToolsPprofSymbol, http.HandlerFunc(pprof.Symbol))
	r.Post(DevToolsPprofSymbol, http.HandlerFunc(pprof.Symbol))
	r.Get(DevToolsPprofTrace, http.HandlerFunc(pprof.Trace))

	for _, name := range []string{"allocs", "block", "goroutine", "heap", "mutex", "threadcreate"} {
		path := DevToolsPprofBasePath + "/" + name
		r.Get(path, pprof.Handler(name))
	}

	r.Post(DevToolsReloadPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := c.reloadEnv(req.Context()); err != nil {
			contract.WriteError(w, req, contract.APIError{
				Status:   http.StatusBadRequest,
				Code:     "env_reload_failed",
				Category: contract.CategoryClient,
				Message:  err.Error(),
			})
			return
		}

		contractio.WriteHTTPResponse(w, req, http.StatusOK, map[string]any{
			"status": "ok",
		})
	}))
}

func (c *DevToolsComponent) RegisterMiddleware(_ *middleware.Registry) {
	c.attachDevMetrics()
}

func (c *DevToolsComponent) Start(ctx context.Context) error {
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

func (c *DevToolsComponent) Stop(_ context.Context) error {
	if c.watchStop != nil {
		c.watchStop()
	}
	c.watchWg.Wait()
	return nil
}

func (c *DevToolsComponent) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"enabled": c.debug}}
	if !c.debug {
		status.Status = health.StatusDegraded
		status.Message = "devtools disabled"
	}
	return "devtools", status
}

func (c *DevToolsComponent) Dependencies() []reflect.Type { return nil }

func (c *DevToolsComponent) reloadEnv(ctx context.Context) error {
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

func (c *DevToolsComponent) attachDevMetrics() {
	if c.devMetrics == nil || c.hooks.AttachDevMetrics == nil {
		return
	}
	c.hooks.AttachDevMetrics(c.devMetrics)
}

func (c *DevToolsComponent) watchEnvFile(ctx context.Context) {
	defer c.watchWg.Done()

	fileSource := config.NewFileSource(c.envFile, config.FormatEnv, true)
	if _, err := fileSource.Load(ctx); err != nil {
		c.logger.Warn("Devtools env watch load failed", log.Fields{"error": err})
	}

	results := fileSource.Watch(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-results:
			if !ok {
				return
			}
			if result.Err != nil {
				c.logger.Warn("Devtools env watch error", log.Fields{"error": result.Err})
			} else if result.Data != nil {
				if err := c.reloadEnv(ctx); err != nil {
					c.logger.Warn("Devtools env reload failed", log.Fields{"error": err})
				}
			}
		}
	}
}

func (c *DevToolsComponent) middlewareList() []string {
	if c.hooks.MiddlewareList == nil {
		return nil
	}
	return c.hooks.MiddlewareList()
}

func (c *DevToolsComponent) configSnapshot() map[string]any {
	if c.hooks.ConfigSnapshot == nil {
		return map[string]any{"debug": c.debug}
	}
	return c.hooks.ConfigSnapshot()
}
