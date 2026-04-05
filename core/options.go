package core

import (
	"time"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/router"
)

// WithRouter sets the router for the App.
func WithRouter(router *router.Router) Option {
	return func(a *App) {
		a.router = router
	}
}

// WithAddr sets the server address.
func WithAddr(address string) Option {
	return func(a *App) {
		a.config.Addr = address
	}
}

// WithEnvPath sets the path to the .env file.
func WithEnvPath(path string) Option {
	return func(a *App) {
		a.config.EnvFile = path
	}
}

// WithShutdownTimeout sets graceful shutdown timeout.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(a *App) {
		a.config.ShutdownTimeout = timeout
	}
}

// WithServerTimeouts configures HTTP server timeouts for read, write and idle connections.
func WithServerTimeouts(read, readHeader, write, idle time.Duration) Option {
	return func(a *App) {
		a.config.ReadTimeout = read
		a.config.ReadHeaderTimeout = readHeader
		a.config.WriteTimeout = write
		a.config.IdleTimeout = idle
	}
}

// WithMaxHeaderBytes sets the maximum header size accepted by the server.
func WithMaxHeaderBytes(bytes int) Option {
	return func(a *App) {
		a.config.MaxHeaderBytes = bytes
	}
}

// WithHTTP2 enables or disables HTTP/2 support when TLS is configured.
func WithHTTP2(enabled bool) Option {
	return func(a *App) {
		a.config.EnableHTTP2 = enabled
	}
}

// WithTLS configures TLS for the app.
func WithTLS(certFile, keyFile string) Option {
	return func(a *App) {
		a.config.TLS = TLSConfig{
			Enabled:  true,
			CertFile: certFile,
			KeyFile:  keyFile,
		}
	}
}

// WithTLSConfig sets the TLS configuration for the App.
func WithTLSConfig(tlsConfig TLSConfig) Option {
	return func(a *App) {
		a.config.TLS = tlsConfig
	}
}

// WithDebug enables debug mode for the app.
func WithDebug() Option {
	return func(a *App) {
		a.config.Debug = true
	}
}

// WithLogger sets a custom logger for the App.
func WithLogger(logger log.StructuredLogger) Option {
	return func(a *App) {
		if logger == nil {
			panic("core logger cannot be nil")
		}
		a.logger = logger
	}
}

// WithMethodNotAllowed enables 405 responses with Allow headers for method mismatches.
func WithMethodNotAllowed(enabled bool) Option {
	return func(a *App) {
		a.hasRouterMethodNotAllowed = true
		a.routerMethodNotAllowed = enabled
	}
}

// WithHTTPMetrics sets the HTTP metrics observer used by the app.
func WithHTTPMetrics(observer metrics.HTTPObserver) Option {
	return func(a *App) {
		a.httpMetrics = observer
	}
}

// WithHealthManager attaches a HealthManager to the App.
// When set, the App will call MarkReady/MarkNotReady on it during lifecycle events.
func WithHealthManager(hm health.HealthManager) Option {
	return func(a *App) {
		a.healthManager = hm
	}
}
