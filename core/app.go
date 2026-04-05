package core

import (
	"net/http"
	"sync"

	"github.com/spcent/plumego/log"
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

	// Runtime state (protected by mutex)
	mu           sync.RWMutex
	started      bool // Whether runtime hooks have started
	configFrozen bool // Whether configuration has been frozen

	// Server components
	httpServer  *http.Server       // HTTP server instance
	connTracker *connectionTracker // Connection tracker for WebSocket
	handler     http.Handler       // Combined handler with middleware applied
	handlerOnce sync.Once          // Ensures handler initialization happens once, can be reset for testing
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

	app.syncRouterConfig(app.router)

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
