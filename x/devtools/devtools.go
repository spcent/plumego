package devtools

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/internal/config"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/x/ops/healthhttp"
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

type DevTools struct {
	debug   bool
	logger  log.StructuredLogger
	envFile string
	hooks   Hooks

	watchOnce sync.Once
	watchStop context.CancelFunc
	watchWg   sync.WaitGroup

	devMetrics *DevCollector
}

// TLSSnapshot is the devtools-owned debug payload for TLS runtime state.
type TLSSnapshot struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// RuntimeSnapshot is the devtools-owned debug payload for core runtime state.
type RuntimeSnapshot struct {
	Addr              string                `json:"addr"`
	ReadTimeout       time.Duration         `json:"read_timeout"`
	ReadHeaderTimeout time.Duration         `json:"read_header_timeout"`
	WriteTimeout      time.Duration         `json:"write_timeout"`
	IdleTimeout       time.Duration         `json:"idle_timeout"`
	MaxHeaderBytes    int                   `json:"max_header_bytes"`
	HTTP2Enabled      bool                  `json:"http2_enabled"`
	DrainInterval     time.Duration         `json:"drain_interval"`
	TLS               TLSSnapshot           `json:"tls"`
	PreparationState  core.PreparationState `json:"preparation_state"`
}

type Hooks struct {
	RuntimeSnapshot  func() RuntimeSnapshot
	MiddlewareList   func() []string
	AttachDevMetrics func(*DevCollector)
}

type Options struct {
	Debug   bool
	Logger  log.StructuredLogger
	EnvFile string
	Hooks   Hooks
}

// ConfigSnapshot exposes the devtools-owned config/runtime payload returned by
// the debug config endpoint.
type ConfigSnapshot struct {
	Debug   bool   `json:"debug"`
	EnvFile string `json:"env_file"`
	RuntimeSnapshot
}

type routeRegistrar interface {
	AddRoute(method, path string, handler http.Handler) error
	Routes() []router.RouteInfo
}

func New(opts Options) *DevTools {
	if opts.Logger == nil {
		opts.Logger = log.NewNoOpLogger()
	}
	return &DevTools{
		debug:      opts.Debug,
		logger:     opts.Logger,
		envFile:    opts.EnvFile,
		hooks:      opts.Hooks,
		devMetrics: NewDevCollector(DefaultDevCollectorConfig()),
	}
}

func (c *DevTools) RegisterRoutes(r routeRegistrar) error {
	if !c.debug {
		return nil
	}

	// Legacy aliases for CLI compatibility.
	if err := r.AddRoute(http.MethodGet, "/_routes", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]any{
			"routes": r.Routes(),
		}
		_ = contract.WriteResponse(w, req, http.StatusOK, payload, nil)
	})); err != nil {
		return err
	}

	if err := r.AddRoute(http.MethodGet, "/_config", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_ = contract.WriteResponse(w, req, http.StatusOK, c.snapshotMap(), nil)
	})); err != nil {
		return err
	}

	if err := r.AddRoute(http.MethodGet, "/_info", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]any{
			"config": c.snapshotMap(),
			"build":  healthhttp.GetBuildInfo(),
		}
		_ = contract.WriteResponse(w, req, http.StatusOK, payload, nil)
	})); err != nil {
		return err
	}

	if err := r.AddRoute(http.MethodGet, DevToolsRoutesPath, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(renderRoutesText(r.Routes())))
	})); err != nil {
		return err
	}

	if err := r.AddRoute(http.MethodGet, DevToolsRoutesJSONPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]any{
			"routes": r.Routes(),
		}
		_ = contract.WriteResponse(w, req, http.StatusOK, payload, nil)
	})); err != nil {
		return err
	}

	if err := r.AddRoute(http.MethodGet, DevToolsMiddlewarePath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload := map[string]any{
			"middlewares": c.middlewareList(),
		}
		_ = contract.WriteResponse(w, req, http.StatusOK, payload, nil)
	})); err != nil {
		return err
	}

	if err := r.AddRoute(http.MethodGet, DevToolsConfigPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_ = contract.WriteResponse(w, req, http.StatusOK, c.snapshotMap(), nil)
	})); err != nil {
		return err
	}

	if err := r.AddRoute(http.MethodGet, DevToolsMetricsPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if c.devMetrics == nil {
			_ = contract.WriteResponse(w, req, http.StatusOK, map[string]any{
				"enabled": false,
			}, nil)
			return
		}

		_ = contract.WriteResponse(w, req, http.StatusOK, map[string]any{
			"enabled": true,
			"http":    c.devMetrics.Snapshot(),
			"db":      c.devMetrics.DBSnapshot(),
		}, nil)
	})); err != nil {
		return err
	}

	if err := r.AddRoute(http.MethodPost, DevToolsMetricsClear, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if c.devMetrics != nil {
			c.devMetrics.Clear()
		}
		_ = contract.WriteResponse(w, req, http.StatusOK, map[string]any{
			"status": "ok",
		}, nil)
	})); err != nil {
		return err
	}

	// pprof endpoints (debug-only)
	if err := r.AddRoute(http.MethodGet, DevToolsPprofBasePath, http.HandlerFunc(pprof.Index)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, DevToolsPprofCmdline, http.HandlerFunc(pprof.Cmdline)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, DevToolsPprofProfile, http.HandlerFunc(pprof.Profile)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, DevToolsPprofSymbol, http.HandlerFunc(pprof.Symbol)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodPost, DevToolsPprofSymbol, http.HandlerFunc(pprof.Symbol)); err != nil {
		return err
	}
	if err := r.AddRoute(http.MethodGet, DevToolsPprofTrace, http.HandlerFunc(pprof.Trace)); err != nil {
		return err
	}

	for _, name := range []string{"allocs", "block", "goroutine", "heap", "mutex", "threadcreate"} {
		path := DevToolsPprofBasePath + "/" + name
		if err := r.AddRoute(http.MethodGet, path, pprof.Handler(name)); err != nil {
			return err
		}
	}

	if err := r.AddRoute(http.MethodPost, DevToolsReloadPath, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := c.reloadEnv(req.Context()); err != nil {
			contract.WriteError(w, req, contract.NewErrorBuilder().
				Status(http.StatusBadRequest).
				Code("env_reload_failed").
				Message(err.Error()).
				Category(contract.CategoryClient).
				Build())
			return
		}

		_ = contract.WriteResponse(w, req, http.StatusOK, map[string]any{
			"status": "ok",
		}, nil)
	})); err != nil {
		return err
	}
	return nil
}

func renderRoutesText(routes []router.RouteInfo) string {
	var b strings.Builder
	b.WriteString("Registered Routes:\n")
	for _, route := range routes {
		label := route.Path
		if strings.Contains(route.Path, "/*") {
			label += "   [wildcard]"
		}
		fmt.Fprintf(&b, "%-6s %s\n", route.Method, label)
	}
	return b.String()
}

func (c *DevTools) AttachMetrics() {
	c.attachDevMetrics()
}

func (c *DevTools) Start(ctx context.Context) error {
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

func (c *DevTools) Stop(_ context.Context) error {
	if c.watchStop != nil {
		c.watchStop()
	}
	c.watchWg.Wait()
	return nil
}

func (c *DevTools) Health() (string, health.HealthStatus) {
	status := health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"enabled": c.debug}}
	if !c.debug {
		status.Status = health.StatusDegraded
		status.Message = "devtools disabled"
	}
	return "devtools", status
}

func (c *DevTools) reloadEnv(ctx context.Context) error {
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

func (c *DevTools) attachDevMetrics() {
	if c.devMetrics == nil || c.hooks.AttachDevMetrics == nil {
		return
	}
	c.hooks.AttachDevMetrics(c.devMetrics)
}

func (c *DevTools) watchEnvFile(ctx context.Context) {
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

func (c *DevTools) middlewareList() []string {
	if c.hooks.MiddlewareList == nil {
		return nil
	}
	return c.hooks.MiddlewareList()
}

func (c *DevTools) snapshot() ConfigSnapshot {
	snapshot := ConfigSnapshot{
		Debug:   c.debug,
		EnvFile: c.envFile,
	}
	if c.hooks.RuntimeSnapshot != nil {
		snapshot.RuntimeSnapshot = c.hooks.RuntimeSnapshot()
	}
	return snapshot
}

func (c *DevTools) snapshotMap() map[string]any {
	snapshot := c.snapshot()
	return map[string]any{
		"addr":                snapshot.Addr,
		"debug":               snapshot.Debug,
		"env_file":            snapshot.EnvFile,
		"read_timeout":        snapshot.ReadTimeout,
		"read_header_timeout": snapshot.ReadHeaderTimeout,
		"write_timeout":       snapshot.WriteTimeout,
		"idle_timeout":        snapshot.IdleTimeout,
		"max_header_bytes":    snapshot.MaxHeaderBytes,
		"http2_enabled":       snapshot.HTTP2Enabled,
		"drain_interval":      snapshot.DrainInterval,
		"preparation_state":   snapshot.PreparationState,
		"tls": map[string]any{
			"enabled":   snapshot.TLS.Enabled,
			"cert_file": snapshot.TLS.CertFile,
			"key_file":  snapshot.TLS.KeyFile,
		},
	}
}
