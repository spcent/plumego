package core

import (
	"time"

	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/security/headers"
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

// WithMaxBodyBytes sets the default request body limit.
func WithMaxBodyBytes(bytes int64) Option {
	return func(a *App) {
		a.config.MaxBodyBytes = bytes
	}
}

// WithSecurityHeadersEnabled toggles default security header injection.
func WithSecurityHeadersEnabled(enabled bool) Option {
	return func(a *App) {
		a.config.EnableSecurityHeaders = enabled
	}
}

// WithSecurityHeadersPolicy sets a custom security header policy.
func WithSecurityHeadersPolicy(policy *headers.Policy) Option {
	return func(a *App) {
		a.config.SecurityHeadersPolicy = policy
		a.config.EnableSecurityHeaders = true
	}
}

// WithAbuseGuardEnabled toggles the abuse guard middleware.
func WithAbuseGuardEnabled(enabled bool) Option {
	return func(a *App) {
		a.config.EnableAbuseGuard = enabled
	}
}

// WithAbuseGuardConfig customizes the abuse guard middleware.
func WithAbuseGuardConfig(cfg middleware.AbuseGuardConfig) Option {
	return func(a *App) {
		a.config.AbuseGuardConfig = &cfg
		a.config.EnableAbuseGuard = true
	}
}

// WithPubSub configures the in-process pubsub implementation for built-in features.
func WithPubSub(ps pubsub.PubSub) Option {
	return func(a *App) {
		a.pub = ps
	}
}

// WithPubSubDebug enables the built-in pubsub debug endpoint.
func WithPubSubDebug(cfg PubSubConfig) Option {
	return func(a *App) {
		a.config.PubSub = cfg
	}
}

// WithWebhookOut enables the outbound webhook management APIs.
func WithWebhookOut(cfg WebhookOutConfig) Option {
	return func(a *App) {
		a.config.WebhookOut = cfg
	}
}

// WithWebhookIn enables the inbound webhook receiver APIs.
func WithWebhookIn(cfg WebhookInConfig) Option {
	return func(a *App) {
		a.config.WebhookIn = cfg
	}
}

// WithConcurrencyLimits sets the maximum concurrent requests and queueing strategy.
func WithConcurrencyLimits(maxConcurrent, queueDepth int, queueTimeout time.Duration) Option {
	return func(a *App) {
		a.config.MaxConcurrency = maxConcurrent
		a.config.QueueDepth = queueDepth
		a.config.QueueTimeout = queueTimeout
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

// WithMetricsCollector sets a metrics collector hook for observability middleware.
func WithMetricsCollector(collector metrics.MetricsCollector) Option {
	return func(a *App) {
		a.metricsCollector = collector
	}
}

// WithTracer sets a tracer hook for observability middleware.
func WithTracer(tracer middleware.Tracer) Option {
	return func(a *App) {
		a.tracer = tracer
	}
}
