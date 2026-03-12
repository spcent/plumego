package plumego

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	compobs "github.com/spcent/plumego/core/components/observability"
	"github.com/spcent/plumego/store/db"
	"github.com/spcent/plumego/tenant"
)

// App is the main application entrypoint. It implements http.Handler so it can be
// passed directly to standard library servers.
type App = core.App

// Option represents functional options for configuring an App.
type Option = core.Option

// Context exposes the unified request helpers used by context-aware handlers.
type Context = contract.Ctx

// ContextHandlerFunc is the handler signature that receives a *Context.
type ContextHandlerFunc = contract.CtxHandlerFunc

// HandlerFunc is the standard net/http handler function signature.
// Re-exported for convenience so callers can use plumego.HandlerFunc
// without importing net/http directly.
type HandlerFunc = http.HandlerFunc

// Component hooks into routing, middleware, and lifecycle events.
type Component = core.Component

// Runner defines a background task with a lifecycle.
type Runner = core.Runner

// ShutdownHook is invoked during application shutdown.
type ShutdownHook = core.ShutdownHook

// TLSConfig re-exports the TLS settings for App configuration.
type TLSConfig = core.TLSConfig

// WebSocketConfig defines the WebSocket hub configuration.
type WebSocketConfig = core.WebSocketConfig

// MetricsConfig configures built-in metrics wiring.
type MetricsConfig = compobs.MetricsConfig

// TracingConfig configures built-in tracing wiring.
type TracingConfig = compobs.TracingConfig

// ObservabilityConfig bundles metrics + tracing settings.
type ObservabilityConfig = compobs.ObservabilityConfig

// New creates a new App instance.
func New(options ...Option) *App {
	return core.New(options...)
}

// Re-export common configuration helpers for convenience.
var (
	WithRouter              = core.WithRouter
	WithAddr                = core.WithAddr
	WithEnvPath             = core.WithEnvPath
	WithShutdownTimeout     = core.WithShutdownTimeout
	WithServerTimeouts      = core.WithServerTimeouts
	WithMaxHeaderBytes      = core.WithMaxHeaderBytes
	WithHTTP2               = core.WithHTTP2
	WithTLS                 = core.WithTLS
	WithTLSConfig           = core.WithTLSConfig
	WithDebug               = core.WithDebug
	WithDevTools            = core.WithDevTools
	WithLogger              = core.WithLogger
	WithComponent           = core.WithComponent
	WithComponents          = core.WithComponents
	WithMethodNotAllowed    = core.WithMethodNotAllowed
	WithShutdownHook        = core.WithShutdownHook
	WithShutdownHooks       = core.WithShutdownHooks
	WithRunner              = core.WithRunner
	WithRunners             = core.WithRunners
	WithHTTPMetrics         = core.WithHTTPMetrics
	WithPrometheusCollector = core.WithPrometheusCollector
	WithTracer              = core.WithTracer
)

// DefaultWebSocketConfig exposes the default WebSocket settings.
var DefaultWebSocketConfig = core.DefaultWebSocketConfig

// DefaultMetricsConfig exposes default metrics settings.
var DefaultMetricsConfig = compobs.DefaultMetricsConfig

// DefaultTracingConfig exposes default tracing settings.
var DefaultTracingConfig = compobs.DefaultTracingConfig

// DefaultObservabilityConfig exposes default observability settings.
var DefaultObservabilityConfig = compobs.DefaultObservabilityConfig

// Tenant types for multi-tenancy support.
type (
	TenantConfig              = tenant.Config
	TenantQuotaConfig         = tenant.QuotaConfig
	TenantPolicyConfig        = tenant.PolicyConfig
	TenantRateLimitConfig     = tenant.RateLimitConfig
	TenantQuotaLimit          = tenant.QuotaLimit
	TenantQuotaWindow         = tenant.QuotaWindow
	TenantRoutePolicy         = tenant.RoutePolicy
	TenantConfigManager       = tenant.ConfigManager
	TenantQuotaManager        = tenant.QuotaManager
	TenantPolicyEvaluator     = tenant.PolicyEvaluator
	TenantRateLimiter         = tenant.RateLimiter
	TenantRoutePolicyStore    = tenant.RoutePolicyStore
	TenantRoutePolicyProvider = tenant.RoutePolicyProvider
	TenantHooks               = tenant.Hooks
	TenantExtractor           = tenant.TenantExtractor
	TenantQuotaRequest        = tenant.QuotaRequest
	TenantQuotaResult         = tenant.QuotaResult
	TenantPolicyRequest       = tenant.PolicyRequest
	TenantPolicyResult        = tenant.PolicyResult
	TenantRateLimitRequest    = tenant.RateLimitRequest
	TenantRateLimitResult     = tenant.RateLimitResult
	TenantResolveInfo         = tenant.ResolveInfo
	TenantQuotaDecision       = tenant.QuotaDecision
	TenantPolicyDecision      = tenant.PolicyDecision
	TenantRateLimitDecision   = tenant.RateLimitDecision
	// TenantRateLimitConfigProviderFromConfig allows using any ConfigManager
	// as a RateLimitConfigProvider for TokenBucketRateLimiter, reading the
	// RateLimitConfig field from the unified tenant Config.
	TenantRateLimitConfigProviderFromConfig = tenant.RateLimitConfigProviderFromConfig
)

// Tenant functions for creating managers and extracting context.
var (
	NewInMemoryTenantConfigManager = tenant.NewInMemoryConfigManager
	NewDBTenantConfigManager       = db.NewDBTenantConfigManager
	NewInMemoryQuotaManager        = tenant.NewInMemoryQuotaManager
	NewSlidingWindowQuotaManager   = tenant.NewSlidingWindowQuotaManager
	NewWindowQuotaManager          = tenant.NewWindowQuotaManager
	NewInMemoryQuotaStore          = tenant.NewInMemoryQuotaStore
	NewConfigPolicyEvaluator       = tenant.NewConfigPolicyEvaluator
	NewInMemoryRateLimitProvider   = tenant.NewInMemoryRateLimitProvider
	NewTokenBucketRateLimiter      = tenant.NewTokenBucketRateLimiter
	NewInMemoryRoutePolicyStore    = tenant.NewInMemoryRoutePolicyStore
	NewInMemoryRoutePolicyCache    = tenant.NewInMemoryRoutePolicyCache
	NewCachedRoutePolicyProvider   = tenant.NewCachedRoutePolicyProvider
	TenantIDFromContext            = tenant.TenantIDFromContext
	ContextWithTenantID            = tenant.ContextWithTenantID
	RequestWithTenantID            = tenant.RequestWithTenantID
	ValidateTenantID               = tenant.ValidateTenantID
	TenantFromHeader               = tenant.FromHeader
	TenantFromQuery                = tenant.FromQuery
	TenantFromCookie               = tenant.FromCookie
	TenantFromSubdomain            = tenant.FromSubdomain
	TenantFromContextValue         = tenant.FromContextValue
	TenantChain                    = tenant.Chain
)

// Tenant database configuration options.
var (
	WithTenantCache  = db.WithTenantCache
	WithTenantColumn = db.WithTenantColumn
)

// Tenant database isolation functions.
var (
	NewTenantDB   = db.NewTenantDB
	ValidateQuery = db.ValidateQuery
)

// Tenant errors.
var (
	ErrTenantNotFound      = tenant.ErrTenantNotFound
	ErrInvalidTenantID     = tenant.ErrInvalidTenantID
	ErrQuotaExceeded       = tenant.ErrQuotaExceeded
	ErrRateLimitExceeded   = tenant.ErrRateLimitExceeded
	ErrRoutePolicyNotFound = tenant.ErrRoutePolicyNotFound
	ErrPolicyDenied        = tenant.ErrPolicyDenied
)
