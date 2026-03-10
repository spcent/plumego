package core

import (
	"net/http"
	"sync"
	"time"

	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/observability"
	"github.com/spcent/plumego/router"
)

// App represents the main application instance.
type App struct {
	// Core components (immutable after construction)
	config        *AppConfig           // Application configuration
	router        *router.Router       // HTTP router
	middlewareReg *middleware.Registry // Middleware registry for all routes
	logger        log.StructuredLogger // Logger instance

	// Runtime state (protected by mutex)
	mu           sync.RWMutex
	started      bool // Whether the app has started
	envLoaded    bool // Whether environment variables have been loaded
	configFrozen bool // Whether configuration has been frozen

	// Server components
	httpServer  *http.Server       // HTTP server instance
	connTracker *connectionTracker // Connection tracker for WebSocket
	handler     http.Handler       // Combined handler with middleware applied
	handlerOnce sync.Once          // Ensures handler initialization happens once, can be reset for testing

	// Optional components
	metricsCollector metrics.MetricsCollector
	tracer           observability.Tracer

	// Component management
	components        []Component
	startedComponents []Component
	componentStopOnce sync.Once
	componentsMounted bool

	runners        []Runner
	startedRunners []Runner
	runnerStopOnce sync.Once

	shutdownHooks []ShutdownHook
	shutdownOnce  sync.Once
}

// Option defines a function type for configuring the App.
type Option func(*App)

// New creates a new App instance with the provided options.
func New(options ...Option) *App {
	defaultConfig := &AppConfig{
		Addr:              ":8080",
		EnvFile:           ".env",
		TLS:               TLSConfig{Enabled: false},
		Debug:             false,
		ShutdownTimeout:   5 * time.Second,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MiB
		EnableHTTP2:       true,
		DrainInterval:     500 * time.Millisecond,
	}

	app := &App{
		config:        defaultConfig,
		router:        router.NewRouter(),
		middlewareReg: middleware.NewRegistry(),
		logger:        log.NewGLogger(),
	}

	for _, opt := range options {
		opt(app)
	}

	if app.router != nil {
		app.router.SetLogger(app.logger)
	}

	return app
}

// Router returns the underlying router for advanced configuration.
func (a *App) Router() *router.Router {
	return a.router
}

// Logger returns the configured application logger.
func (a *App) Logger() log.StructuredLogger {
	return a.logger
}
