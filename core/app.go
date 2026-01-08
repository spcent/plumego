package core

import (
	"net/http"
	"sync"
	"time"

	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/pubsub"
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
	mu            sync.RWMutex
	started       bool // Whether the app has started
	envLoaded     bool // Whether environment variables have been loaded
	guardsApplied bool // Whether guards have been applied
	configFrozen  bool // Whether configuration has been frozen

	// Server components
	httpServer  *http.Server       // HTTP server instance
	connTracker *connectionTracker // Connection tracker for WebSocket
	handler     http.Handler       // Combined handler with middleware applied
	handlerOnce sync.Once          // Ensures handler initialization happens once

	// Optional components
	metricsCollector middleware.MetricsCollector
	tracer           middleware.Tracer
	pub              pubsub.PubSub
	loggingEnabled   bool

	// Component management
	components        []Component
	startedComponents []Component
	componentStopOnce sync.Once
	componentsMounted bool
	
	// Dependency injection container
	diContainer       *DIContainer
}

// Option defines a function type for configuring the App.
type Option func(*App)

// New creates a new App instance with the provided options.
func New(options ...Option) *App {
	defaultConfig := &AppConfig{
		Addr:                  ":8080",
		EnvFile:               ".env",
		TLS:                   TLSConfig{Enabled: false},
		Debug:                 false,
		ShutdownTimeout:       5 * time.Second,
		ReadTimeout:           30 * time.Second,
		ReadHeaderTimeout:     5 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           60 * time.Second,
		MaxHeaderBytes:        1 << 20, // 1 MiB
		EnableHTTP2:           true,
		DrainInterval:         500 * time.Millisecond,
		MaxBodyBytes:          10 << 20, // 10 MiB
		MaxConcurrency:        256,
		QueueDepth:            512,
		QueueTimeout:          250 * time.Millisecond,
		EnableSecurityHeaders: true,
		EnableAbuseGuard:      true,
	}

	app := &App{
		config:        defaultConfig,
		router:        router.NewRouter(),
		middlewareReg: middleware.NewRegistry(),
		logger:        log.NewGLogger(),
		diContainer:   NewDIContainer(),
	}

	for _, opt := range options {
		opt(app)
	}

	if app.router != nil {
		app.router.SetLogger(app.logger)
	}

	// Register core services in DI container
	app.registerCoreServices()

	return app
}

// registerCoreServices registers core application services in the DI container.
func (a *App) registerCoreServices() {
	// Register app instance
	a.diContainer.Register(a)
	
	// Register router
	a.diContainer.Register(a.router)
	
	// Register middleware registry
	a.diContainer.Register(a.middlewareReg)
	
	// Register logger
	a.diContainer.Register(a.logger)
}

// Router returns the underlying router for advanced configuration.
func (a *App) Router() *router.Router {
	return a.router
}

// Logger returns the configured application logger.
func (a *App) Logger() log.StructuredLogger {
	return a.logger
}
