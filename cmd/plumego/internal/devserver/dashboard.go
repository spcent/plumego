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

	"github.com/spcent/plumego"
	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/frontend"
	"github.com/spcent/plumego/net/websocket"
	"github.com/spcent/plumego/pubsub"
)

const dashboardRoom = "dashboard"

// Dashboard is the development dashboard server
type Dashboard struct {
	app       *core.App
	hub       *websocket.Hub
	pubsub    *pubsub.InProcPubSub
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
		core.WithRecommendedMiddleware(),
	)

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
	d.app.GetCtx("/api/info", d.handleInfo)
	d.app.GetCtx("/api/status", d.handleStatus)
	d.app.GetCtx("/api/health", d.handleHealth)
	d.app.GetCtx("/api/routes", d.handleRoutes)
	d.app.GetCtx("/api/config", d.handleConfig)
	d.app.GetCtx("/api/config/edit", d.handleConfigEditGet)
	d.app.PostCtx("/api/config/edit", d.handleConfigEditSave)
	d.app.GetCtx("/api/metrics", d.handleMetrics)
	d.app.PostCtx("/api/metrics/clear", d.handleMetricsClear)
	d.app.GetCtx("/api/deps", d.handleDeps)
	d.app.GetCtx("/api/pprof/types", d.handlePprofTypes)
	d.app.GetCtx("/api/pprof/raw", d.handlePprofRaw)
	d.app.PostCtx("/api/test", d.handleAPITest)
	d.app.PostCtx("/api/build", d.handleBuild)
	d.app.PostCtx("/api/restart", d.handleRestart)
	d.app.PostCtx("/api/stop", d.handleStop)

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

func (d *Dashboard) handleInfo(ctx *plumego.Context) {
	ctx.JSON(http.StatusOK, d.getDashboardInfo())
}

func (d *Dashboard) handleStatus(ctx *plumego.Context) {
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

func (d *Dashboard) handleHealth(ctx *plumego.Context) {
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

func (d *Dashboard) handleBuild(ctx *plumego.Context) {
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

func (d *Dashboard) handleRestart(ctx *plumego.Context) {
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

func (d *Dashboard) handleStop(ctx *plumego.Context) {
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

func (d *Dashboard) handleRoutes(ctx *plumego.Context) {
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

func (d *Dashboard) handleConfig(ctx *plumego.Context) {
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

func (d *Dashboard) handleMetrics(ctx *plumego.Context) {
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

func (d *Dashboard) handleMetricsClear(ctx *plumego.Context) {
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

func (d *Dashboard) handlePprofTypes(ctx *plumego.Context) {
	ctx.JSON(http.StatusOK, map[string]any{
		"types": pprofProfiles(),
	})
}

func (d *Dashboard) handlePprofRaw(ctx *plumego.Context) {
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

func (d *Dashboard) handleAPITest(ctx *plumego.Context) {
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
func (d *Dashboard) GetPubSub() *pubsub.InProcPubSub {
	return d.pubsub
}

// GetBuilder returns the Builder instance
func (d *Dashboard) GetBuilder() BuilderAPI {
	return d.builder
}

// GetRunner returns the AppRunner instance
func (d *Dashboard) GetRunner() RunnerAPI {
	return d.runner
}
