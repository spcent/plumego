package devserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	plog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/observability"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/x/frontend"
	"github.com/spcent/plumego/x/pubsub"
	"github.com/spcent/plumego/x/websocket"
)

const dashboardRoom = "dashboard"

// Dashboard is the development dashboard server
type Dashboard struct {
	app       *core.App
	hub       *websocket.Hub
	pubsub    *pubsub.InProcBroker
	builder   *Builder
	runner    *AppRunner
	analyzer  *Analyzer
	depsCache *depsCache
	startTime time.Time

	// Configuration
	dashboardAddr string
	appAddr       string
	projectDir    string
}

// Config holds dashboard configuration
type Config struct {
	DashboardAddr string
	AppAddr       string
	ProjectDir    string
	UIPath        string

	// Optional custom build command. Empty means use the default go build.
	CustomBuildCmd  string
	CustomBuildArgs []string

	// Optional custom run command. Empty means run the built binary.
	CustomRunCmd  string
	CustomRunArgs []string

	// OutputPassthrough forwards app stdout/stderr to the CLI output.
	OutputPassthrough bool
}

// NewDashboard creates a new development dashboard
func NewDashboard(cfg Config) (*Dashboard, error) {
	// Validate project directory
	absDir, err := filepath.Abs(cfg.ProjectDir)
	if err != nil {
		return nil, fmt.Errorf("invalid project directory: %w", err)
	}

	// Create plumego app for dashboard
	app := core.New(
		core.WithAddr(cfg.DashboardAddr),
		core.WithDebug(),
		core.WithLogger(plog.NewGLogger()),
	)
	if err := app.Use(
		observability.RequestID(),
		observability.Tracing(nil),
		observability.HTTPMetrics(nil),
		observability.AccessLog(app.Logger()),
		recovery.Recovery(app.Logger()),
		cors.CORS,
	); err != nil {
		return nil, fmt.Errorf("register dashboard middleware: %w", err)
	}

	// Create WebSocket hub (4 workers, queue size 100)
	hub := websocket.NewHub(4, 100)

	// Create PubSub for event coordination
	ps := pubsub.New()

	d := &Dashboard{
		app:           app,
		hub:           hub,
		pubsub:        ps,
		dashboardAddr: cfg.DashboardAddr,
		appAddr:       cfg.AppAddr,
		projectDir:    absDir,
		startTime:     time.Now(),
		depsCache:     newDepsCache(),
	}

	// Create builder, runner, and analyzer
	d.builder = NewBuilder(absDir, ps)
	d.runner = NewAppRunner(absDir, ps)
	d.analyzer = NewAnalyzer(fmt.Sprintf("http://localhost%s", cfg.AppAddr))

	// Set app address for runner
	d.runner.SetEnv("APP_ADDR", cfg.AppAddr)
	d.runner.SetEnv("APP_DEBUG", "true")

	// Apply optional custom commands from config
	if cfg.CustomBuildCmd != "" {
		d.builder.SetCustomBuild(cfg.CustomBuildCmd, cfg.CustomBuildArgs)
	}
	if cfg.CustomRunCmd != "" {
		d.runner.SetCustomCommand(cfg.CustomRunCmd, cfg.CustomRunArgs)
	}
	d.runner.SetOutputPassthrough(cfg.OutputPassthrough)

	// Register routes
	d.registerRoutes(cfg.UIPath)

	// Subscribe to events and broadcast to WebSocket
	d.subscribeEvents()

	return d, nil
}

// registerRoutes sets up HTTP routes
func (d *Dashboard) registerRoutes(uiPath string) {
	// WebSocket endpoint for real-time events
	d.app.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Use ServeWSWithAuth to handle WebSocket upgrade
		websocket.ServeWSWithAuth(
			w, r,
			d.hub,
			nil, // No auth
			32,  // Queue size
			5*time.Second,
			websocket.SendBlock, // Block on send
		)
	})

	// API endpoints (without Group - register directly)
	adaptCtx := func(handler contract.CtxHandlerFunc) http.HandlerFunc {
		return contract.AdaptCtxHandler(handler, d.app.Logger()).ServeHTTP
	}

	d.app.Get("/api/info", adaptCtx(d.handleInfo))
	d.app.Get("/api/status", adaptCtx(d.handleStatus))
	d.app.Get("/api/health", adaptCtx(d.handleHealth))
	d.app.Get("/api/routes", adaptCtx(d.handleRoutes))
	d.app.Get("/api/config", adaptCtx(d.handleConfig))
	d.app.Get("/api/config/edit", adaptCtx(d.handleConfigEditGet))
	d.app.Post("/api/config/edit", adaptCtx(d.handleConfigEditSave))
	d.app.Get("/api/metrics", adaptCtx(d.handleMetrics))
	d.app.Post("/api/metrics/clear", adaptCtx(d.handleMetricsClear))
	d.app.Get("/api/deps", adaptCtx(d.handleDeps))
	d.app.Get("/api/pprof/types", adaptCtx(d.handlePprofTypes))
	d.app.Get("/api/pprof/raw", adaptCtx(d.handlePprofRaw))
	d.app.Post("/api/test", adaptCtx(d.handleAPITest))
	d.app.Post("/api/build", adaptCtx(d.handleBuild))
	d.app.Post("/api/restart", adaptCtx(d.handleRestart))
	d.app.Post("/api/stop", adaptCtx(d.handleStop))

	// Static UI files
	router := d.app.Router()

	// Try embedded UI first, then fallback to disk
	if HasEmbeddedUI() {
		uiFS, err := GetUIFS()
		if err == nil {
			frontend.RegisterFS(router, http.FS(uiFS),
				frontend.WithPrefix("/"),
				frontend.WithIndex("index.html"),
			)
		}
	} else if uiPath != "" {
		frontend.RegisterFromDir(router, uiPath,
			frontend.WithPrefix("/"),
			frontend.WithIndex("index.html"),
		)
	}
}

// subscribeEvents subscribes to all events and broadcasts to WebSocket
func (d *Dashboard) subscribeEvents() {
	// Subscribe to all events using wildcard (with default options)
	sub, _ := d.pubsub.Subscribe("*", pubsub.SubOptions{})

	// Start a goroutine to forward events to WebSocket
	go func() {
		for msg := range sub.C() {
			// Create a simple event message
			event := map[string]any{
				"type":      msg.Topic,
				"timestamp": time.Now().Format(time.RFC3339),
				"data":      msg.Data,
			}

			// Convert to JSON
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			// Broadcast to all WebSocket clients in dashboard room
			// Note: OpText is 0x01 for text frames
			d.hub.BroadcastRoom(dashboardRoom, 0x01, data)
		}
	}()
}

// Start starts the dashboard server
func (d *Dashboard) Start(ctx context.Context) error {
	// Note: WebSocket hub workers are automatically started in NewHub()

	// Start the dashboard server in background
	go func() {
		if err := d.app.Boot(); err != nil {
			fmt.Printf("Dashboard server error: %v\n", err)
		}
	}()

	// Wait a bit for server to start
	time.Sleep(500 * time.Millisecond)

	// Publish initial dashboard info event
	d.publishDashboardInfo()

	// Re-publish dashboard info on app lifecycle changes
	d.subscribeLifecycleForInfo()

	return nil
}

// publishDashboardInfo publishes the current dashboard info via pubsub.
func (d *Dashboard) publishDashboardInfo() {
	d.pubsub.Publish(EventDashboard, pubsub.Message{
		Topic: EventDashboard,
		Data:  d.getDashboardInfo(),
	})
}

// subscribeLifecycleForInfo re-publishes dashboard info when app state changes.
func (d *Dashboard) subscribeLifecycleForInfo() {
	patterns := []string{EventAppStart, EventAppStop, EventAppRestart}
	for _, pattern := range patterns {
		sub, err := d.pubsub.Subscribe(pattern, pubsub.SubOptions{})
		if err != nil {
			continue
		}
		go func() {
			for range sub.C() {
				// Small delay to let the runner state settle
				time.Sleep(100 * time.Millisecond)
				d.publishDashboardInfo()
			}
		}()
	}
}

// Stop stops the dashboard server
func (d *Dashboard) Stop(ctx context.Context) error {
	// Stop the runner if running
	if d.runner.IsRunning() {
		d.runner.Stop()
	}

	// Stop WebSocket hub
	d.hub.Stop()

	// Note: core.App doesn't have Shutdown method, server stops when main exits
	return nil
}

// BuildAndRun builds and runs the application
func (d *Dashboard) BuildAndRun(ctx context.Context) error {
	// Verify build environment
	if err := d.builder.Verify(); err != nil {
		return fmt.Errorf("build verification failed: %w", err)
	}

	// Build
	if err := d.builder.Build(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if d.builder.HasCustomBuild() && !d.runner.HasCustomCommand() {
		if _, err := os.Stat(d.builder.OutputPath()); err != nil {
			return fmt.Errorf("custom build command must output %s or set --run-cmd", d.builder.OutputPath())
		}
	}

	// Start the application
	if err := d.runner.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Rebuild rebuilds and restarts the application
func (d *Dashboard) Rebuild(ctx context.Context) error {
	// Stop if running
	if d.runner.IsRunning() {
		if err := d.runner.Stop(); err != nil {
			return fmt.Errorf("failed to stop: %w", err)
		}
	}

	// Build and run
	return d.BuildAndRun(ctx)
}

// HTTP Handlers

func (d *Dashboard) handleInfo(ctx *contract.Ctx) {
	ctx.JSON(http.StatusOK, d.getDashboardInfo())
}

func (d *Dashboard) handleStatus(ctx *contract.Ctx) {
	info := d.getDashboardInfo()
	status := map[string]any{
		"dashboard": map[string]any{
			"version": info.Version,
			"url":     info.DashboardURL,
			"uptime":  info.Uptime,
		},
		"app": map[string]any{
			"url":     info.AppURL,
			"running": info.AppRunning,
			"pid":     info.AppPID,
		},
		"project": map[string]any{
			"dir":        info.ProjectDir,
			"go_version": info.GoVersion,
		},
	}

	ctx.JSON(http.StatusOK, status)
}

func (d *Dashboard) handleHealth(ctx *contract.Ctx) {
	healthy := d.runner.IsRunning()

	ctx.JSON(http.StatusOK, map[string]any{
		"healthy": healthy,
		"checks": map[string]string{
			"app": func() string {
				if healthy {
					return "running"
				}
				return "stopped"
			}(),
		},
	})
}

func (d *Dashboard) handleBuild(ctx *contract.Ctx) {
	if err := d.builder.Build(); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "Build completed successfully",
	})
}

func (d *Dashboard) handleRestart(ctx *contract.Ctx) {
	bgCtx := context.Background()

	if err := d.Rebuild(bgCtx); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "Application restarted successfully",
	})
}

func (d *Dashboard) handleStop(ctx *contract.Ctx) {
	if err := d.runner.Stop(); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "Application stopped successfully",
	})
}

func (d *Dashboard) handleRoutes(ctx *contract.Ctx) {
	if !d.runner.IsRunning() {
		ctx.JSON(http.StatusServiceUnavailable, map[string]any{
			"error": "Application is not running",
		})
		return
	}

	// Try to fetch routes from the running app
	routes, err := d.analyzer.GetRoutes()
	if err != nil {
		// Fallback to probing if debug endpoint is not available
		routes = d.analyzer.ProbeEndpoints()
		if len(routes) == 0 {
			ctx.JSON(http.StatusOK, map[string]any{
				"routes": []RouteInfo{},
				"error":  "Could not fetch routes: " + err.Error(),
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"routes": routes,
		"count":  len(routes),
	})
}

func (d *Dashboard) handleConfig(ctx *contract.Ctx) {
	if !d.runner.IsRunning() {
		ctx.JSON(http.StatusServiceUnavailable, map[string]any{
			"error": "Application is not running",
		})
		return
	}

	config, err := d.analyzer.GetAppInfo()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Could not fetch config: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, config)
}

func (d *Dashboard) handleMetrics(ctx *contract.Ctx) {
	metrics := map[string]any{
		"dashboard": map[string]any{
			"uptime":    time.Since(d.startTime).Seconds(),
			"startTime": d.startTime.Format(time.RFC3339),
		},
		"app": map[string]any{
			"running": d.runner.IsRunning(),
			"pid":     d.getAppPID(),
		},
	}

	alerts, thresholds := evaluateRequestAlerts(nil)

	// If app is running, try to get health info
	if d.runner.IsRunning() {
		healthy, details, err := d.analyzer.HealthCheck()
		if err == nil {
			metrics["app"].(map[string]any)["healthy"] = healthy
			metrics["app"].(map[string]any)["healthDetails"] = details
		}

		if devMetrics, err := d.analyzer.GetDevMetrics(); err == nil {
			metrics["app"].(map[string]any)["requests"] = devMetrics.HTTP
			metrics["app"].(map[string]any)["db"] = devMetrics.DB
			alerts, thresholds = evaluateRequestAlerts(&devMetrics.HTTP)
		} else {
			metrics["app"].(map[string]any)["requests_error"] = err.Error()
		}
	}

	metrics["alerts"] = alerts
	metrics["thresholds"] = thresholds

	ctx.JSON(http.StatusOK, metrics)
}

func (d *Dashboard) handleMetricsClear(ctx *contract.Ctx) {
	if !d.runner.IsRunning() {
		ctx.JSON(http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "Application is not running",
		})
		return
	}

	if err := d.analyzer.ClearDevMetrics(); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
	})
}

func (d *Dashboard) handlePprofTypes(ctx *contract.Ctx) {
	ctx.JSON(http.StatusOK, map[string]any{
		"types": pprofProfiles(),
	})
}

func (d *Dashboard) handlePprofRaw(ctx *contract.Ctx) {
	if !d.runner.IsRunning() {
		ctx.JSON(http.StatusServiceUnavailable, map[string]any{
			"error": "Application is not running",
		})
		return
	}

	profileType, seconds, err := parsePprofRequest(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	payload, contentType, err := d.analyzer.FetchPprof(profileType, seconds)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
		return
	}

	if download := strings.TrimSpace(ctx.Query.Get("download")); download == "0" || strings.EqualFold(download, "false") {
		ctx.JSON(http.StatusOK, map[string]any{
			"type":         profileType,
			"seconds":      seconds,
			"content_type": contentType,
			"size_bytes":   len(payload),
			"preview_hex":  previewHex(payload, 96),
			"download_url": fmt.Sprintf("/api/pprof/raw?type=%s&seconds=%d", profileType, seconds),
		})
		return
	}

	ctx.W.Header().Set("Content-Type", contentType)
	ctx.W.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.W.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pprof", profileType))
	ctx.W.WriteHeader(http.StatusOK)
	_, _ = ctx.W.Write(payload)
}

func (d *Dashboard) handleAPITest(ctx *contract.Ctx) {
	if !d.runner.IsRunning() {
		ctx.JSON(http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "Application is not running",
		})
		return
	}

	var req APITestRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	resp, err := d.analyzer.DoAPITest(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// Helper methods

func (d *Dashboard) getDashboardInfo() DashboardInfo {
	return DashboardInfo{
		Version:      "0.1.0",
		DashboardURL: fmt.Sprintf("http://localhost%s", d.dashboardAddr),
		AppURL:       fmt.Sprintf("http://localhost%s", d.appAddr),
		Uptime:       time.Since(d.startTime).String(),
		UptimeMS:     time.Since(d.startTime).Milliseconds(),
		StartTime:    d.startTime.Format(time.RFC3339),
		ProjectDir:   d.projectDir,
		GoVersion:    runtime.Version(),
		AppRunning:   d.runner.IsRunning(),
		AppPID:       d.getAppPID(),
	}
}

func (d *Dashboard) getAppPID() int {
	if d.runner.process != nil {
		return d.runner.process.Pid
	}
	return 0
}

// PublishEvent publishes an event to the event bus
func (d *Dashboard) PublishEvent(eventType string, data any) {
	d.pubsub.Publish(eventType, pubsub.Message{
		Topic: eventType,
		Data:  data,
	})
}

// GetPubSub returns the PubSub instance for external use
func (d *Dashboard) GetPubSub() *pubsub.InProcBroker {
	return d.pubsub
}

// GetBuilder exposes the builder behind the dashboard.
func (d *Dashboard) GetBuilder() BuilderAPI {
	return d.builder
}

// GetRunner exposes the runner behind the dashboard.
func (d *Dashboard) GetRunner() RunnerAPI {
	return d.runner
}

// SetOutputPassthrough controls whether app stdout/stderr is forwarded to the CLI.
func (d *Dashboard) SetOutputPassthrough(enabled bool) {
	d.runner.SetOutputPassthrough(enabled)
}
