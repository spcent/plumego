package plumego

import (
	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
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

// ComponentFunc is a helper implementation of Component.
type ComponentFunc = core.ComponentFunc

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

// New creates a new App instance.
func New(options ...Option) *App {
	return core.New(options...)
}

// Re-export common configuration helpers for convenience.
var (
	WithRouter            = core.WithRouter
	WithAddr              = core.WithAddr
	WithEnvPath           = core.WithEnvPath
	WithShutdownTimeout   = core.WithShutdownTimeout
	WithServerTimeouts    = core.WithServerTimeouts
	WithMaxHeaderBytes    = core.WithMaxHeaderBytes
	WithMaxBodyBytes      = core.WithMaxBodyBytes
	WithPubSub            = core.WithPubSub
	WithPubSubDebug       = core.WithPubSubDebug
	WithWebhookOut        = core.WithWebhookOut
	WithWebhookIn         = core.WithWebhookIn
	WithConcurrencyLimits = core.WithConcurrencyLimits
	WithHTTP2             = core.WithHTTP2
	WithTLS               = core.WithTLS
	WithTLSConfig         = core.WithTLSConfig
	WithDebug             = core.WithDebug
	WithLogger            = core.WithLogger
	WithComponent         = core.WithComponent
	WithComponents        = core.WithComponents
	WithMetricsCollector  = core.WithMetricsCollector
	WithTracer            = core.WithTracer
)

// DefaultWebSocketConfig exposes the default WebSocket settings.
var DefaultWebSocketConfig = core.DefaultWebSocketConfig
