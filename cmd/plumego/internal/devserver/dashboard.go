package devserver

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
	"github.com/spcent/plumego/x/devtools"
	"github.com/spcent/plumego/x/frontend"
	"github.com/spcent/plumego/x/pubsub"
	"github.com/spcent/plumego/x/websocket"
)

const dashboardRoom = "dashboard"

const (
	devserverCodeConfigEditPathInvalid = "CONFIG_EDIT_PATH_INVALID"
	devserverCodeConfigEditReadFailed  = "CONFIG_EDIT_READ_FAILED"
	devserverCodeConfigEditWriteFailed = "CONFIG_EDIT_WRITE_FAILED"
	devserverCodeAppRebuildFailed      = "APP_REBUILD_FAILED"
	devserverCodeDependencyGraphFailed = "DEPENDENCY_GRAPH_FAILED"
	devserverCodeBuildFailed           = "BUILD_FAILED"
	devserverCodeAppRestartFailed      = "APP_RESTART_FAILED"
	devserverCodeAppStopFailed         = "APP_STOP_FAILED"
	devserverCodeAppNotRunning         = "APP_NOT_RUNNING"
	devserverCodeConfigFetchFailed     = "CONFIG_FETCH_FAILED"
	devserverCodeMetricsClearFailed    = "METRICS_CLEAR_FAILED"
	devserverCodeInvalidPprofRequest   = "INVALID_PPROF_REQUEST"
	devserverCodePprofFetchFailed      = "PPROF_FETCH_FAILED"
	devserverCodeAPITestFailed         = "API_TEST_FAILED"
)

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
	appCfg := core.DefaultConfig()
	appCfg.Addr = cfg.DashboardAddr
	app := core.New(appCfg, core.AppDependencies{Logger: plog.NewLogger()})
	if err := app.Use(
		requestid.Middleware(),
		mwtracing.Middleware(nil),
		httpmetrics.Middleware(nil),
		accesslog.Middleware(app.Logger(), nil, nil),
		recovery.Recovery(app.Logger()),
		cors.Middleware(cors.CORSOptions{}),
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
	if err := d.registerRoutes(cfg.UIPath); err != nil {
		return nil, fmt.Errorf("register dashboard routes: %w", err)
	}

	// Subscribe to events and broadcast to WebSocket
	d.subscribeEvents()

	return d, nil
}

// registerRoutes sets up HTTP routes.
func (d *Dashboard) registerRoutes(uiPath string) error {
	// WebSocket endpoint for real-time events
	if err := d.app.Get("/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use ServeWSWithAuth to handle WebSocket upgrade
		websocket.ServeWSWithAuth(
			w, r,
			d.hub,
			nil, // No auth
			32,  // Queue size
			5*time.Second,
			websocket.SendBlock, // Block on send
		)
	})); err != nil {
		return fmt.Errorf("register /ws: %w", err)
	}

	// API endpoints (without Group - register directly)
	if err := d.app.Get("/api/info", http.HandlerFunc(d.handleInfo)); err != nil {
		return fmt.Errorf("register /api/info: %w", err)
	}
	if err := d.app.Get("/api/status", http.HandlerFunc(d.handleStatus)); err != nil {
		return fmt.Errorf("register /api/status: %w", err)
	}
	if err := d.app.Get("/api/health", http.HandlerFunc(d.handleHealth)); err != nil {
		return fmt.Errorf("register /api/health: %w", err)
	}
	if err := d.app.Get("/api/routes", http.HandlerFunc(d.handleRoutes)); err != nil {
		return fmt.Errorf("register /api/routes: %w", err)
	}
	if err := d.app.Get("/api/config", http.HandlerFunc(d.handleConfig)); err != nil {
		return fmt.Errorf("register /api/config: %w", err)
	}
	if err := d.app.Get("/api/config/edit", http.HandlerFunc(d.handleConfigEditGet)); err != nil {
		return fmt.Errorf("register /api/config/edit GET: %w", err)
	}
	if err := d.app.Post("/api/config/edit", http.HandlerFunc(d.handleConfigEditSave)); err != nil {
		return fmt.Errorf("register /api/config/edit POST: %w", err)
	}
	if err := d.app.Get("/api/metrics", http.HandlerFunc(d.handleMetrics)); err != nil {
		return fmt.Errorf("register /api/metrics: %w", err)
	}
	if err := d.app.Post("/api/metrics/clear", http.HandlerFunc(d.handleMetricsClear)); err != nil {
		return fmt.Errorf("register /api/metrics/clear: %w", err)
	}
	if err := d.app.Get("/api/deps", http.HandlerFunc(d.handleDeps)); err != nil {
		return fmt.Errorf("register /api/deps: %w", err)
	}
	if err := d.app.Get("/api/pprof/types", http.HandlerFunc(d.handlePprofTypes)); err != nil {
		return fmt.Errorf("register /api/pprof/types: %w", err)
	}
	if err := d.app.Get("/api/pprof/raw", http.HandlerFunc(d.handlePprofRaw)); err != nil {
		return fmt.Errorf("register /api/pprof/raw: %w", err)
	}
	if err := d.app.Post("/api/test", http.HandlerFunc(d.handleAPITest)); err != nil {
		return fmt.Errorf("register /api/test: %w", err)
	}
	if err := d.app.Post("/api/build", http.HandlerFunc(d.handleBuild)); err != nil {
		return fmt.Errorf("register /api/build: %w", err)
	}
	if err := d.app.Post("/api/restart", http.HandlerFunc(d.handleRestart)); err != nil {
		return fmt.Errorf("register /api/restart: %w", err)
	}
	if err := d.app.Post("/api/stop", http.HandlerFunc(d.handleStop)); err != nil {
		return fmt.Errorf("register /api/stop: %w", err)
	}

	// Try embedded UI first, then fallback to disk
	if HasEmbeddedUI() {
		uiFS, err := GetUIFS()
		if err == nil {
			if err := frontend.RegisterFS(d.app, http.FS(uiFS),
				frontend.WithPrefix("/"),
				frontend.WithIndex("index.html"),
			); err != nil {
				return fmt.Errorf("register embedded ui: %w", err)
			}
		}
	} else if uiPath != "" {
		if err := frontend.RegisterFromDir(d.app, uiPath,
			frontend.WithPrefix("/"),
			frontend.WithIndex("index.html"),
		); err != nil {
			return fmt.Errorf("register ui dir: %w", err)
		}
	}

	return nil
}

// subscribeEvents subscribes to all events and broadcasts to WebSocket
func (d *Dashboard) subscribeEvents() {
	// Subscribe to all events using wildcard (with default options)
	sub, _ := d.pubsub.Subscribe(context.Background(), "*", pubsub.SubOptions{})

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

	if err := d.app.Prepare(); err != nil {
		return fmt.Errorf("prepare dashboard app: %w", err)
	}
	srv, err := d.app.Server()
	if err != nil {
		return fmt.Errorf("get dashboard server: %w", err)
	}

	// Start the dashboard server in background
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
		sub, err := d.pubsub.Subscribe(context.Background(), pattern, pubsub.SubOptions{})
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
		if err := d.runner.Stop(); err != nil {
			return fmt.Errorf("stop dashboard runner: %w", err)
		}
	}

	// Stop WebSocket hub
	d.hub.Stop()

	if err := d.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown dashboard app: %w", err)
	}

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

type dashboardActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type dashboardStatusResponse struct {
	Dashboard dashboardStatusDashboard `json:"dashboard"`
	App       dashboardStatusApp       `json:"app"`
	Project   dashboardStatusProject   `json:"project"`
}

type dashboardStatusDashboard struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	Uptime  string `json:"uptime"`
}

type dashboardStatusApp struct {
	URL     string `json:"url"`
	Running bool   `json:"running"`
	PID     int    `json:"pid"`
}

type dashboardStatusProject struct {
	Dir       string `json:"dir"`
	GoVersion string `json:"go_version"`
}

type dashboardHealthResponse struct {
	Healthy bool                  `json:"healthy"`
	Checks  dashboardHealthChecks `json:"checks"`
}

type dashboardHealthChecks struct {
	App string `json:"app"`
}

type dashboardRoutesResponse struct {
	Routes []RouteInfo `json:"routes"`
	Count  int         `json:"count,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type dashboardMetricsResponse struct {
	Dashboard  dashboardMetricsDashboard `json:"dashboard"`
	App        dashboardMetricsApp       `json:"app"`
	Alerts     []RequestAlert            `json:"alerts"`
	Thresholds RequestAlertThresholds    `json:"thresholds"`
}

type dashboardMetricsDashboard struct {
	Uptime    float64 `json:"uptime"`
	StartTime string  `json:"startTime"`
}

type dashboardMetricsApp struct {
	Running       bool                      `json:"running"`
	PID           int                       `json:"pid"`
	Healthy       *bool                     `json:"healthy,omitempty"`
	HealthDetails map[string]any            `json:"healthDetails,omitempty"`
	Requests      *devtools.DevHTTPSnapshot `json:"requests,omitempty"`
	DB            *devtools.DevDBSnapshot   `json:"db,omitempty"`
	RequestsError string                    `json:"requests_error,omitempty"`
}

type dashboardPprofTypesResponse struct {
	Types []pprofProfile `json:"types"`
}

type dashboardPprofPreviewResponse struct {
	Type        string `json:"type"`
	Seconds     int    `json:"seconds"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
	PreviewHex  string `json:"preview_hex"`
	DownloadURL string `json:"download_url"`
}

func writeDashboardActionResponse(w http.ResponseWriter, r *http.Request, message string) {
	_ = contract.WriteResponse(w, r, http.StatusOK, dashboardActionResponse{
		Success: true,
		Message: message,
	}, nil)
}

// HTTP Handlers

func (d *Dashboard) handleInfo(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, d.getDashboardInfo(), nil)
}

func (d *Dashboard) handleStatus(w http.ResponseWriter, r *http.Request) {
	info := d.getDashboardInfo()
	status := dashboardStatusResponse{
		Dashboard: dashboardStatusDashboard{
			Version: info.Version,
			URL:     info.DashboardURL,
			Uptime:  info.Uptime,
		},
		App: dashboardStatusApp{
			URL:     info.AppURL,
			Running: info.AppRunning,
			PID:     info.AppPID,
		},
		Project: dashboardStatusProject{
			Dir:       info.ProjectDir,
			GoVersion: info.GoVersion,
		},
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, status, nil)
}

func (d *Dashboard) handleHealth(w http.ResponseWriter, r *http.Request) {
	healthy := d.runner.IsRunning()
	appStatus := "stopped"
	if healthy {
		appStatus = "running"
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, dashboardHealthResponse{
		Healthy: healthy,
		Checks:  dashboardHealthChecks{App: appStatus},
	}, nil)
}

func (d *Dashboard) handleBuild(w http.ResponseWriter, r *http.Request) {
	if err := d.builder.Build(); err != nil {
		writeDevserverError(w, r, contract.TypeInternal, devserverCodeBuildFailed, "build failed")
		return
	}

	writeDashboardActionResponse(w, r, "Build completed successfully")
}

func (d *Dashboard) handleRestart(w http.ResponseWriter, r *http.Request) {
	bgCtx := context.Background()

	if err := d.Rebuild(bgCtx); err != nil {
		writeDevserverError(w, r, contract.TypeInternal, devserverCodeAppRestartFailed, "application restart failed")
		return
	}

	writeDashboardActionResponse(w, r, "Application restarted successfully")
}

func (d *Dashboard) handleStop(w http.ResponseWriter, r *http.Request) {
	if err := d.runner.Stop(); err != nil {
		writeDevserverError(w, r, contract.TypeInternal, devserverCodeAppStopFailed, "application stop failed")
		return
	}

	writeDashboardActionResponse(w, r, "Application stopped successfully")
}

func (d *Dashboard) handleRoutes(w http.ResponseWriter, r *http.Request) {
	if !d.runner.IsRunning() {
		writeDevserverError(w, r, contract.TypeUnavailable, devserverCodeAppNotRunning, "application is not running")
		return
	}

	// Try to fetch routes from the running app
	routes, err := d.analyzer.GetRoutes()
	if err != nil {
		// Fallback to probing if debug endpoint is not available
		routes = d.analyzer.ProbeEndpoints()
		if len(routes) == 0 {
			_ = contract.WriteResponse(w, r, http.StatusOK, dashboardRoutesResponse{
				Routes: []RouteInfo{},
				Error:  "could not fetch routes",
			}, nil)
			return
		}
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, dashboardRoutesResponse{
		Routes: routes,
		Count:  len(routes),
	}, nil)
}

func (d *Dashboard) handleConfig(w http.ResponseWriter, r *http.Request) {
	if !d.runner.IsRunning() {
		writeDevserverError(w, r, contract.TypeUnavailable, devserverCodeAppNotRunning, "application is not running")
		return
	}

	snapshot, err := d.analyzer.GetAppSnapshot()
	if err != nil {
		writeDevserverError(w, r, contract.TypeInternal, devserverCodeConfigFetchFailed, "config unavailable")
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, snapshot, nil)
}

func (d *Dashboard) handleMetrics(w http.ResponseWriter, r *http.Request) {
	appMetrics := dashboardMetricsApp{
		Running: d.runner.IsRunning(),
		PID:     d.getAppPID(),
	}

	alerts, thresholds := evaluateRequestAlerts(nil)

	// If app is running, try to get health info
	if d.runner.IsRunning() {
		healthy, details, err := d.analyzer.HealthCheck()
		if err == nil {
			appMetrics.Healthy = &healthy
			appMetrics.HealthDetails = details
		}

		if devMetrics, err := d.analyzer.GetDevMetrics(); err == nil {
			appMetrics.Requests = &devMetrics.HTTP
			appMetrics.DB = &devMetrics.DB
			alerts, thresholds = evaluateRequestAlerts(&devMetrics.HTTP)
		} else {
			appMetrics.RequestsError = "request metrics unavailable"
		}
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, dashboardMetricsResponse{
		Dashboard: dashboardMetricsDashboard{
			Uptime:    time.Since(d.startTime).Seconds(),
			StartTime: d.startTime.Format(time.RFC3339),
		},
		App:        appMetrics,
		Alerts:     alerts,
		Thresholds: thresholds,
	}, nil)
}

func (d *Dashboard) handleMetricsClear(w http.ResponseWriter, r *http.Request) {
	if !d.runner.IsRunning() {
		writeDevserverError(w, r, contract.TypeUnavailable, devserverCodeAppNotRunning, "application is not running")
		return
	}

	if err := d.analyzer.ClearDevMetrics(); err != nil {
		writeDevserverError(w, r, contract.TypeInternal, devserverCodeMetricsClearFailed, "metrics could not be cleared")
		return
	}

	writeDashboardActionResponse(w, r, "")
}

func (d *Dashboard) handlePprofTypes(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, dashboardPprofTypesResponse{Types: pprofProfiles()}, nil)
}

func (d *Dashboard) handlePprofRaw(w http.ResponseWriter, r *http.Request) {
	if !d.runner.IsRunning() {
		writeDevserverError(w, r, contract.TypeUnavailable, devserverCodeAppNotRunning, "application is not running")
		return
	}

	profileType, seconds, err := parsePprofRequest(r)
	if err != nil {
		writeDevserverError(w, r, contract.TypeValidation, devserverCodeInvalidPprofRequest, "invalid pprof request")
		return
	}

	payload, contentType, err := d.analyzer.FetchPprof(profileType, seconds)
	if err != nil {
		writeDevserverError(w, r, contract.TypeInternal, devserverCodePprofFetchFailed, "pprof data unavailable")
		return
	}

	if download := strings.TrimSpace(r.URL.Query().Get("download")); download == "0" || strings.EqualFold(download, "false") {
		_ = contract.WriteResponse(w, r, http.StatusOK, dashboardPprofPreviewResponse{
			Type:        profileType,
			Seconds:     seconds,
			ContentType: contentType,
			SizeBytes:   len(payload),
			PreviewHex:  previewHex(payload, 96),
			DownloadURL: fmt.Sprintf("/api/pprof/raw?type=%s&seconds=%d", profileType, seconds),
		}, nil)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pprof", profileType))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (d *Dashboard) handleAPITest(w http.ResponseWriter, r *http.Request) {
	if !d.runner.IsRunning() {
		writeDevserverError(w, r, contract.TypeUnavailable, devserverCodeAppNotRunning, "application is not running")
		return
	}

	var req APITestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Category(contract.CategoryValidation).
			Code(contract.CodeInvalidJSON).
			Message("invalid request body").
			Build())
		return
	}

	resp, err := d.analyzer.DoAPITest(req)
	if err != nil {
		writeDevserverError(w, r, contract.TypeValidation, devserverCodeAPITestFailed, "api test failed")
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, resp, nil)
}

func writeDevserverError(w http.ResponseWriter, r *http.Request, errorType contract.ErrorType, code string, message string) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(errorType).
		Code(code).
		Message(message).
		Build())
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
