package core

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/observability"
	"github.com/spcent/plumego/middleware/ratelimit"
	"github.com/spcent/plumego/middleware/tenant"
	tenantmw "github.com/spcent/plumego/tenant/middleware"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/security/headers"
	tenants "github.com/spcent/plumego/tenant"
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
func WithCORSOptions(opts cors.CORSOptions) Option {
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
func WithAbuseGuardConfig(cfg ratelimit.AbuseGuardConfig) Option {
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
func WithTracer(tracer observability.Tracer) Option {
	return func(a *App) {
		a.tracer = tracer
	}
}

// WithTenantConfigManager registers a tenant config manager as a component.
// This enables multi-tenancy support with per-tenant configuration.
func WithTenantConfigManager(manager tenants.ConfigManager) Option {
	return func(a *App) {
		component := &TenantConfigComponent{
			Name:    "tenant-config",
			Manager: manager,
		}
		a.components = append(a.components, component)
	}
}

// TenantMiddlewareOptions configures the tenant middleware chain registered by WithTenantMiddleware.
type TenantMiddlewareOptions struct {
	// ── Resolver ─────────────────────────────────────────────────────────────

	// HeaderName is the HTTP header to extract tenant ID from (default: "X-Tenant-ID").
	HeaderName string
	// Extractor is a custom function to extract tenant ID from the request.
	// When set, takes precedence over HeaderName (Principal check still runs first).
	Extractor tenants.TenantExtractor
	// AllowMissing allows requests without tenant ID (default: false).
	AllowMissing bool
	// DisablePrincipal disables tenant extraction from the authenticated Principal.
	DisablePrincipal bool
	// OnMissing is called when tenant ID is required but not found.
	OnMissing func(http.ResponseWriter, *http.Request)

	// ── Rate limiter ──────────────────────────────────────────────────────────

	// RateLimiter enforces per-tenant rate limits (optional).
	RateLimiter tenants.RateLimiter
	// RateLimitEstimator returns the token cost for a request (default: 1).
	RateLimitEstimator func(*http.Request) int64
	// OnRateLimitRejected is called when a request is rejected by the rate limiter.
	OnRateLimitRejected func(http.ResponseWriter, *http.Request, tenants.RateLimitResult)

	// ── Quota ─────────────────────────────────────────────────────────────────

	// QuotaManager enforces per-tenant quota limits (optional).
	QuotaManager tenants.QuotaManager
	// QuotaTokensHeader is the request header carrying the token count
	// (default: "X-Token-Count").
	QuotaTokensHeader string
	// QuotaEstimator returns the token cost for a request.
	// Takes precedence over QuotaTokensHeader when set.
	QuotaEstimator func(*http.Request) int
	// OnQuotaRejected is called when a request is rejected by the quota manager.
	OnQuotaRejected func(http.ResponseWriter, *http.Request, tenants.QuotaResult)

	// ── Policy ────────────────────────────────────────────────────────────────

	// PolicyEvaluator enforces per-tenant policies (optional).
	PolicyEvaluator tenants.PolicyEvaluator
	// PolicyModelHeader is the request header carrying the model name
	// (default: "X-Model").
	PolicyModelHeader string
	// PolicyToolHeader is the request header carrying the tool name
	// (default: "X-Tool").
	PolicyToolHeader string
	// OnPolicyDenied is called when a request is denied by the policy evaluator.
	OnPolicyDenied func(http.ResponseWriter, *http.Request, tenants.PolicyResult)

	// ── Hooks ─────────────────────────────────────────────────────────────────

	// Hooks provides callbacks for tenant resolution, rate limit, quota, and policy events.
	Hooks tenants.Hooks
}

// WithTenantMiddleware adds the full tenant middleware chain:
// resolver → rate limiter → quota → policy.
// Each stage is only added when the relevant option is non-nil.
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
		// 1. Tenant resolver — always added.
		a.Use(tenant.TenantResolver(tenant.TenantResolverOptions{
			HeaderName:       options.HeaderName,
			Extractor:        options.Extractor,
			AllowMissing:     options.AllowMissing,
			DisablePrincipal: options.DisablePrincipal,
			OnMissing:        options.OnMissing,
			Hooks:            options.Hooks,
		}))

		// 2. Rate limiter (optional).
		if options.RateLimiter != nil {
			a.Use(tenant.TenantRateLimit(tenant.TenantRateLimitOptions{
				Limiter:    options.RateLimiter,
				Estimator:  options.RateLimitEstimator,
				Hooks:      options.Hooks,
				OnRejected: options.OnRateLimitRejected,
			}))
		}

		// 3. Quota enforcement (optional).
		if options.QuotaManager != nil {
			a.Use(tenantmw.TenantQuota(tenantmw.TenantQuotaOptions{
				Manager:      options.QuotaManager,
				TokensHeader: options.QuotaTokensHeader,
				Estimator:    options.QuotaEstimator,
				Hooks:        options.Hooks,
				OnRejected:   options.OnQuotaRejected,
			}))
		}

		// 4. Policy enforcement (optional).
		if options.PolicyEvaluator != nil {
			a.Use(tenantmw.TenantPolicy(tenantmw.TenantPolicyOptions{
				Evaluator:   options.PolicyEvaluator,
				ModelHeader: options.PolicyModelHeader,
				ToolHeader:  options.PolicyToolHeader,
				Hooks:       options.Hooks,
				OnDenied:    options.OnPolicyDenied,
			}))
		}
	}
}
