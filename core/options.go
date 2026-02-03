package core

import (
	"time"

	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/security/headers"
	"github.com/spcent/plumego/tenant"
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

// WithRecovery enables the recovery middleware.
func WithRecovery() Option {
	return func(a *App) {
		_ = a.enableRecovery()
	}
}

// WithLogging enables the logging middleware.
func WithLogging() Option {
	return func(a *App) {
		_ = a.enableLogging()
	}
}

// WithRequestID enables the request id middleware.
func WithRequestID() Option {
	return func(a *App) {
		_ = a.enableRequestID()
	}
}

// WithRecommendedMiddleware enables the default production-safe middleware chain.
// It includes RequestID, Logging, and Recovery in the recommended order.
func WithRecommendedMiddleware() Option {
	return func(a *App) {
		_ = a.enableRequestID()
		_ = a.enableLogging()
		_ = a.enableRecovery()
	}
}

// WithCORS enables the default CORS middleware.
func WithCORS() Option {
	return func(a *App) {
		_ = a.enableCORS(nil)
	}
}

// WithCORSOptions enables CORS with custom options.
func WithCORSOptions(opts middleware.CORSOptions) Option {
	return func(a *App) {
		_ = a.enableCORS(&opts)
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
func WithTracer(tracer middleware.Tracer) Option {
	return func(a *App) {
		a.tracer = tracer
	}
}

// WithTenantConfigManager registers a tenant config manager as a component.
// This enables multi-tenancy support with per-tenant configuration.
func WithTenantConfigManager(manager tenant.ConfigManager) Option {
	return func(a *App) {
		component := &TenantConfigComponent{
			Name:    "tenant-config",
			Manager: manager,
		}
		a.components = append(a.components, component)
	}
}

// TenantMiddlewareOptions configures tenant middleware chain.
type TenantMiddlewareOptions struct {
	// HeaderName is the HTTP header to extract tenant ID from (default: "X-Tenant-ID")
	HeaderName string
	// AllowMissing allows requests without tenant ID (default: false)
	AllowMissing bool
	// DisablePrincipal disables tenant extraction from Principal (default: false)
	DisablePrincipal bool
	// QuotaManager enforces per-tenant quota limits (optional)
	QuotaManager tenant.QuotaManager
	// PolicyEvaluator enforces per-tenant policies (optional)
	PolicyEvaluator tenant.PolicyEvaluator
	// Hooks provides callbacks for tenant resolution, quota, and policy events
	Hooks tenant.Hooks
}

// WithTenantMiddleware adds tenant resolution, quota, and policy middleware.
// This should be used after WithTenantConfigManager.
//
// Example:
//
//	app := core.New(
//	    core.WithTenantConfigManager(manager),
//	    core.WithTenantMiddleware(core.TenantMiddlewareOptions{
//	        QuotaManager:    quotaMgr,
//	        PolicyEvaluator: policyEval,
//	    }),
//	)
func WithTenantMiddleware(options TenantMiddlewareOptions) Option {
	return func(a *App) {
		// Add tenant resolver middleware
		a.Use(middleware.TenantResolver(middleware.TenantResolverOptions{
			HeaderName:       options.HeaderName,
			AllowMissing:     options.AllowMissing,
			DisablePrincipal: options.DisablePrincipal,
			Hooks:            options.Hooks,
		}))

		// Add quota enforcement if configured
		if options.QuotaManager != nil {
			a.Use(middleware.TenantQuota(middleware.TenantQuotaOptions{
				Manager: options.QuotaManager,
				Hooks:   options.Hooks,
			}))
		}

		// Add policy enforcement if configured
		if options.PolicyEvaluator != nil {
			a.Use(middleware.TenantPolicy(middleware.TenantPolicyOptions{
				Evaluator: options.PolicyEvaluator,
				Hooks:     options.Hooks,
			}))
		}
	}
}
