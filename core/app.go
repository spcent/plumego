package core

import (
	"io"
	"net/http"
	"sync"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

// App represents the main application instance.
type App struct {
	// Core components (immutable after construction)
	config          *AppConfig           // Application configuration
	router          *router.Router       // HTTP router
	middlewareChain *middleware.Chain    // Middleware pipeline for all routes
	logger          log.StructuredLogger // Logger instance
	// Declarative router option state applied to the owned app router.
	hasRouterMethodNotAllowed bool
	routerMethodNotAllowed    bool

	// Runtime state (protected by mutex)
	mu            sync.RWMutex
	started       bool // Whether runtime hooks have started
	configFrozen  bool // Whether configuration has been frozen
	loggerStarted bool

	// Server components
	httpServer  *http.Server       // HTTP server instance
	connTracker *connectionTracker // Connection tracker for WebSocket
	handler     http.Handler       // Combined handler with middleware applied
	handlerOnce sync.Once          // Ensures handler initialization happens once, can be reset for testing

	// Optional components
	httpMetrics   metrics.HTTPObserver
	healthManager health.HealthManager
}

// Option defines a function type for configuring non-config app dependencies.
type Option func(*App)

// New creates a new App instance with the provided typed config and options.
func New(cfg AppConfig, options ...Option) *App {
	config := cfg
	app := &App{
		config:          &config,
		router:          router.NewRouter(),
		middlewareChain: middleware.NewChain(),
		logger:          log.NewNoOpLogger(),
	}

	for _, opt := range options {
		opt(app)
	}

	app.syncRouterConfig(app.router, app.hasRouterMethodNotAllowed, app.routerMethodNotAllowed)

	return app
}

// Logger returns the configured application logger.
func (a *App) Logger() log.StructuredLogger {
	return a.logger
}

// Routes returns the owned route table snapshot.
func (a *App) Routes() []router.RouteInfo {
	r := a.ensureRouter()
	if r == nil {
		return nil
	}
	return r.Routes()
}

// Print writes the owned route table to w.
func (a *App) Print(w io.Writer) {
	r := a.ensureRouter()
	if r == nil {
		return
	}
	r.Print(w)
}
