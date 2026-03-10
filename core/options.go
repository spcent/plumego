package core

import (
	"time"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/observability"
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
		if logger != nil {
			a.logger = logger
		}
	}
}

// WithComponent mounts a Component into the application lifecycle.
func WithComponent(component Component) Option {
	return func(a *App) {
		if component != nil {
			a.components = append(a.components, component)
		}
	}
}

// WithComponents mounts multiple Components into the application lifecycle.
func WithComponents(components ...Component) Option {
	return func(a *App) {
		for _, component := range components {
			if component != nil {
				a.components = append(a.components, component)
			}
		}
	}
}

// WithMethodNotAllowed enables 405 responses with Allow headers for method mismatches.
func WithMethodNotAllowed(enabled bool) Option {
	return func(a *App) {
		r := a.ensureRouter()
		if r != nil {
			r.SetMethodNotAllowed(enabled)
		}
	}
}

// WithShutdownHook registers a shutdown hook.
func WithShutdownHook(hook ShutdownHook) Option {
	return func(a *App) {
		if hook != nil {
			a.shutdownHooks = append(a.shutdownHooks, hook)
		}
	}
}

// WithShutdownHooks registers multiple shutdown hooks.
func WithShutdownHooks(hooks ...ShutdownHook) Option {
	return func(a *App) {
		for _, hook := range hooks {
			if hook != nil {
				a.shutdownHooks = append(a.shutdownHooks, hook)
			}
		}
	}
}

// WithRunner registers a background runner in the application lifecycle.
func WithRunner(runner Runner) Option {
	return func(a *App) {
		if runner != nil {
			a.runners = append(a.runners, runner)
		}
	}
}

// WithRunners registers multiple background runners.
func WithRunners(runners ...Runner) Option {
	return func(a *App) {
		for _, runner := range runners {
			if runner != nil {
				a.runners = append(a.runners, runner)
			}
		}
	}
}

// WithMetricsCollector sets a metrics collector hook for observability middleware.
func WithMetricsCollector(collector metrics.MetricsCollector) Option {
	return func(a *App) {
		a.metricsCollector = collector
	}
}

// WithTracer sets a tracer hook for observability middleware.
func WithTracer(tracer observability.Tracer) Option {
	return func(a *App) {
		a.tracer = tracer
	}
}

// WithHealthManager attaches a HealthManager to the App.
// When set, the App will call MarkReady/MarkNotReady on it during lifecycle events.
func WithHealthManager(hm health.HealthManager) Option {
	return func(a *App) {
		a.healthManager = hm
	}
}
