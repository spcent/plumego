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
	mu               sync.RWMutex
	preparationState PreparationState // Tracks mutation and preparation phase

	// Server components
	httpServer  *http.Server       // HTTP server instance
	connTracker *connectionTracker // Connection tracker for WebSocket
	handler     http.Handler       // Combined handler with middleware applied
	handlerOnce sync.Once          // Ensures handler initialization happens once, can be reset for testing
}

// New creates a new App instance with typed config and typed dependencies.
func New(cfg AppConfig, dependencies AppDependencies) *App {
	config := cfg
	app := &App{
		config:           &config,
		router:           router.NewRouter(),
		middlewareChain:  middleware.NewChain(),
		logger:           resolveLogger(dependencies),
		preparationState: PreparationStateMutable,
	}

	// Apply typed router policy once at construction.
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
