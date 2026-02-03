package plumego

import (
	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/middleware"
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

// Component hooks into routing, middleware, and lifecycle events.
type Component = core.Component

// Runner defines a background task with a lifecycle.
type Runner = core.Runner

// ShutdownHook is invoked during application shutdown.
type ShutdownHook = core.ShutdownHook

// CORSOptions configures the CORS middleware.
type CORSOptions = middleware.CORSOptions

// TLSConfig re-exports the TLS settings for App configuration.
type TLSConfig = core.TLSConfig

// PubSubConfig re-exports the debug Pub/Sub snapshot configuration.
type PubSubConfig = core.PubSubConfig

// WebhookOutConfig configures outbound webhook management endpoints.
type WebhookOutConfig = core.WebhookOutConfig

// WebhookInConfig configures inbound webhook receivers.
type WebhookInConfig = core.WebhookInConfig

// WebSocketConfig defines the WebSocket hub configuration.
type WebSocketConfig = core.WebSocketConfig

// MetricsConfig configures built-in metrics wiring.
type MetricsConfig = core.MetricsConfig

// TracingConfig configures built-in tracing wiring.
type TracingConfig = core.TracingConfig

// ObservabilityConfig bundles metrics + tracing settings.
type ObservabilityConfig = core.ObservabilityConfig

// New creates a new App instance.
func New(options ...Option) *App {
	return core.New(options...)
}

// Re-export common configuration helpers for convenience.
var (
	WithRouter                 = core.WithRouter
	WithAddr                   = core.WithAddr
	WithEnvPath                = core.WithEnvPath
	WithShutdownTimeout        = core.WithShutdownTimeout
	WithServerTimeouts         = core.WithServerTimeouts
	WithMaxHeaderBytes         = core.WithMaxHeaderBytes
	WithMaxBodyBytes           = core.WithMaxBodyBytes
	WithRecovery               = core.WithRecovery
	WithLogging                = core.WithLogging
	WithRequestID              = core.WithRequestID
	WithRecommendedMiddleware  = core.WithRecommendedMiddleware
	WithCORS                   = core.WithCORS
	WithCORSOptions            = core.WithCORSOptions
	WithSecurityHeadersEnabled = core.WithSecurityHeadersEnabled
	WithSecurityHeadersPolicy  = core.WithSecurityHeadersPolicy
	WithAbuseGuardEnabled      = core.WithAbuseGuardEnabled
	WithAbuseGuardConfig       = core.WithAbuseGuardConfig
	WithPubSub                 = core.WithPubSub
	WithPubSubDebug            = core.WithPubSubDebug
	WithWebhookOut             = core.WithWebhookOut
	WithWebhookIn              = core.WithWebhookIn
	WithConcurrencyLimits      = core.WithConcurrencyLimits
	WithHTTP2                  = core.WithHTTP2
	WithTLS                    = core.WithTLS
	WithTLSConfig              = core.WithTLSConfig
	WithDebug                  = core.WithDebug
	WithLogger                 = core.WithLogger
	WithComponent              = core.WithComponent
	WithComponents             = core.WithComponents
	WithMethodNotAllowed       = core.WithMethodNotAllowed
	WithShutdownHook           = core.WithShutdownHook
	WithShutdownHooks          = core.WithShutdownHooks
	WithRunner                 = core.WithRunner
	WithRunners                = core.WithRunners
	WithMetricsCollector       = core.WithMetricsCollector
	WithTracer                 = core.WithTracer
)

// DefaultWebSocketConfig exposes the default WebSocket settings.
var DefaultWebSocketConfig = core.DefaultWebSocketConfig

// DefaultMetricsConfig exposes default metrics settings.
var DefaultMetricsConfig = core.DefaultMetricsConfig

// DefaultTracingConfig exposes default tracing settings.
var DefaultTracingConfig = core.DefaultTracingConfig

// DefaultObservabilityConfig exposes default observability settings.
var DefaultObservabilityConfig = core.DefaultObservabilityConfig

// Tenant types for multi-tenancy support.
type (
	TenantConfig          = tenant.Config
	TenantQuotaConfig     = tenant.QuotaConfig
	TenantPolicyConfig    = tenant.PolicyConfig
	TenantConfigManager   = tenant.ConfigManager
	TenantQuotaManager    = tenant.QuotaManager
	TenantPolicyEvaluator = tenant.PolicyEvaluator
	TenantHooks           = tenant.Hooks
	TenantQuotaRequest    = tenant.QuotaRequest
	TenantQuotaResult     = tenant.QuotaResult
	TenantPolicyRequest   = tenant.PolicyRequest
	TenantPolicyResult    = tenant.PolicyResult
	TenantResolveInfo     = tenant.ResolveInfo
	TenantQuotaDecision   = tenant.QuotaDecision
	TenantPolicyDecision  = tenant.PolicyDecision
)

// Tenant configuration and middleware options.
type TenantMiddlewareOptions = core.TenantMiddlewareOptions

// Tenant functions for creating managers and extracting context.
var (
	NewInMemoryTenantConfigManager = tenant.NewInMemoryConfigManager
	NewDBTenantConfigManager       = db.NewDBTenantConfigManager
	NewInMemoryQuotaManager        = tenant.NewInMemoryQuotaManager
	NewSlidingWindowQuotaManager   = tenant.NewSlidingWindowQuotaManager
	NewConfigPolicyEvaluator       = tenant.NewConfigPolicyEvaluator
	TenantIDFromContext            = tenant.TenantIDFromContext
	ContextWithTenantID            = tenant.ContextWithTenantID
	RequestWithTenantID            = tenant.RequestWithTenantID
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

// Tenant configuration options.
var (
	WithTenantConfigManager = core.WithTenantConfigManager
	WithTenantMiddleware    = core.WithTenantMiddleware
)

// Tenant errors.
var (
	ErrTenantNotFound = tenant.ErrTenantNotFound
	ErrQuotaExceeded  = tenant.ErrQuotaExceeded
	ErrPolicyDenied   = tenant.ErrPolicyDenied
)
