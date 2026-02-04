package core

import (
	"time"

	"github.com/spcent/plumego/middleware/ratelimit"
	"github.com/spcent/plumego/security/headers"
)

// TLSConfig defines TLS configuration.
type TLSConfig struct {
	Enabled  bool   // Whether to enable TLS
	CertFile string // Path to TLS certificate file
	KeyFile  string // Path to TLS private key file
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
	AbuseGuardConfig      *ratelimit.AbuseGuardConfig

	PubSub     PubSubConfig     // PubSub configuration
	WebhookOut WebhookOutConfig // Webhook outbound configuration
	WebhookIn  WebhookInConfig  // Webhook inbound configuration
}
