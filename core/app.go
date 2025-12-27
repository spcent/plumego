package core

import (
	"net/http"
	"time"

	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/net/webhookin"
	ws "github.com/spcent/plumego/net/websocket"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
)

// App represents the main application instance.
type App struct {
	config        AppConfig               // Application configuration
	router        *router.Router          // HTTP router
	wsHub         *ws.Hub                 // WebSocket hub
	started       bool                    // Whether the app has started
	envLoaded     bool                    // Whether environment variables have been loaded
	httpServer    *http.Server            // HTTP server instance
	middlewares   []middleware.Middleware // Stored middleware for all routes
	handler       http.Handler            // Combined handler with middleware applied
	connTracker   *connectionTracker      // Connection tracker for WebSocket
	guardsApplied bool                    // Whether guards have been applied

	logger           log.StructuredLogger
	metricsCollector middleware.MetricsCollector
	tracer           middleware.Tracer

	pub              pubsub.PubSub
	webhookInDeduper *webhookin.Deduper
}

// Option defines a function type for configuring the App.
type Option func(*App)

// New creates a new App instance with the provided options.
func New(options ...Option) *App {
	app := &App{
		config: AppConfig{
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
			MaxBodyBytes:      10 << 20, // 10 MiB
			MaxConcurrency:    256,
			QueueDepth:        512,
			QueueTimeout:      250 * time.Millisecond,
		},
		router: router.NewRouter(),
		logger: log.NewGLogger(),
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
