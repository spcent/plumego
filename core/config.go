package core

import (
	"time"

	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/net/webhookin"
	webhook "github.com/spcent/plumego/net/webhookout"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/security/headers"
)

// TLSConfig defines TLS configuration.
type TLSConfig struct {
	Enabled  bool   // Whether to enable TLS
	CertFile string // Path to TLS certificate file
	KeyFile  string // Path to TLS private key file
}

// PubSubConfig configures the optional pubsub snapshot endpoint.
type PubSubConfig struct {
	Enabled bool          // Whether to enable the debug endpoint
	Path    string        // Path for the debug endpoint
	Pub     pubsub.PubSub // PubSub instance for snapshotting
}

// WebhookOutConfig configures the outbound webhook management endpoints.
type WebhookOutConfig struct {
	Enabled bool             // Whether to enable the outbound webhook endpoints
	Service *webhook.Service // Webhook service instance

	// TriggerToken protects event triggering. When empty and AllowEmptyToken is false, triggers are forbidden.
	TriggerToken     string // Token for event triggering
	AllowEmptyToken  bool   // Whether to allow empty trigger token
	BasePath         string // Base path for webhook endpoints
	IncludeStats     bool   // Whether to include stats in responses
	DefaultPageLimit int    // Default page limit for paginated responses
}

// WebhookInConfig configures inbound webhook receivers.
type WebhookInConfig struct {
	Enabled bool          // Whether to enable the inbound webhook endpoints
	Pub     pubsub.PubSub // PubSub instance for event publishing

	GitHubSecret      string             // Secret for GitHub webhook events
	StripeSecret      string             // Secret for Stripe webhook events
	MaxBodyBytes      int64              // Maximum allowed request body size
	StripeTolerance   time.Duration      // Tolerance for Stripe webhook event timestamps
	DedupTTL          time.Duration      // Time-to-live for deduplication cache
	Deduper           *webhookin.Deduper // Deduper instance for event deduplication
	GitHubPath        string             // Path for GitHub webhook events
	StripePath        string             // Path for Stripe webhook events
	TopicPrefixGitHub string             // Prefix for GitHub event topics
	TopicPrefixStripe string             // Prefix for Stripe event topics
}

// AppConfig defines application configuration.
type AppConfig struct {
	Addr            string        // Server address
	EnvFile         string        // Path to .env file
	TLS             TLSConfig     // TLS configuration
	Debug           bool          // Debug mode
	ShutdownTimeout time.Duration // Graceful shutdown timeout
	// HTTP server hardening
	ReadTimeout       time.Duration // Maximum duration for reading the entire request, including the body
	ReadHeaderTimeout time.Duration // Maximum duration for reading the request headers (slowloris protection)
	WriteTimeout      time.Duration // Maximum duration before timing out writes of the response
	IdleTimeout       time.Duration // Maximum time to wait for the next request when keep-alives are enabled
	MaxHeaderBytes    int           // Maximum size of request headers
	EnableHTTP2       bool          // Whether to keep HTTP/2 support enabled
	DrainInterval     time.Duration // How often to log in-flight connection counts while draining
	MaxBodyBytes      int64         // Per-request body limit
	MaxConcurrency    int           // Maximum number of concurrent requests being served
	QueueDepth        int           // Maximum number of requests allowed to queue while waiting for a worker
	QueueTimeout      time.Duration // Maximum time a request can wait in the queue
	// Security guardrails
	EnableSecurityHeaders bool
	SecurityHeadersPolicy *headers.Policy
	EnableAbuseGuard      bool
	AbuseGuardConfig      *middleware.AbuseGuardConfig

	PubSub     PubSubConfig     // PubSub configuration
	WebhookOut WebhookOutConfig // Webhook outbound configuration
	WebhookIn  WebhookInConfig  // Webhook inbound configuration
}
