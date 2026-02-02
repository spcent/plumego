package devserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
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
	}

	// Create builder and runner
	d.builder = NewBuilder(absDir, ps)
	d.runner = NewAppRunner(absDir, ps)

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
	d.app.GetCtx("/api/status", d.handleStatus)
	d.app.GetCtx("/api/health", d.handleHealth)
	d.app.PostCtx("/api/build", d.handleBuild)
	d.app.PostCtx("/api/restart", d.handleRestart)
	d.app.PostCtx("/api/stop", d.handleStop)

	// Static UI files
	if uiPath != "" {
		router := d.app.Router()
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
			event := map[string]interface{}{
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
	// Start WebSocket hub
	go d.hub.Run()

	// Start the dashboard server in background
	go func() {
		if err := d.app.Boot(); err != nil {
			fmt.Printf("Dashboard server error: %v\n", err)
		}
	}()

	// Wait a bit for server to start
	time.Sleep(500 * time.Millisecond)

	return nil
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

func (d *Dashboard) handleStatus(ctx *plumego.Context) {
	status := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"url":    fmt.Sprintf("http://localhost%s", d.dashboardAddr),
			"uptime": time.Since(d.startTime).String(),
		},
		"app": map[string]interface{}{
			"url":     fmt.Sprintf("http://localhost%s", d.appAddr),
			"running": d.runner.IsRunning(),
			"pid":     d.getAppPID(),
		},
		"project": map[string]interface{}{
			"dir": d.projectDir,
		},
	}

	ctx.JSON(http.StatusOK, status)
}

func (d *Dashboard) handleHealth(ctx *plumego.Context) {
	healthy := d.runner.IsRunning()

	ctx.JSON(http.StatusOK, map[string]interface{}{
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
		ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Build completed successfully",
	})
}

func (d *Dashboard) handleRestart(ctx *plumego.Context) {
	bgCtx := context.Background()

	if err := d.Rebuild(bgCtx); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Application restarted successfully",
	})
}

func (d *Dashboard) handleStop(ctx *plumego.Context) {
	if err := d.runner.Stop(); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Application stopped successfully",
	})
}

// Helper methods

func (d *Dashboard) getDashboardInfo() DashboardInfo {
	return DashboardInfo{
		Version:      "0.1.0",
		DashboardURL: fmt.Sprintf("http://localhost%s", d.dashboardAddr),
		AppURL:       fmt.Sprintf("http://localhost%s", d.appAddr),
		Uptime:       time.Since(d.startTime).String(),
	}
}

func (d *Dashboard) getAppPID() int {
	if d.runner.process != nil {
		return d.runner.process.Pid
	}
	return 0
}

// PublishEvent publishes an event to the event bus
func (d *Dashboard) PublishEvent(eventType string, data interface{}) {
	d.pubsub.Publish(eventType, data)
}

// GetPubSub returns the PubSub instance for external use
func (d *Dashboard) GetPubSub() *pubsub.InProcPubSub {
	return d.pubsub
}

// GetBuilder returns the Builder instance
func (d *Dashboard) GetBuilder() *Builder {
	return d.builder
}

// GetRunner returns the AppRunner instance
func (d *Dashboard) GetRunner() *AppRunner {
	return d.runner
}
