package pubsub

import (
	"time"

	"github.com/spcent/plumego/metrics"
)

// Config holds the configuration for InProcPubSub.
type Config struct {
	// ShardCount is the number of shards for topic partitioning (default: 16)
	// Higher values reduce lock contention but increase memory usage
	ShardCount int

	// DefaultBufferSize is the default channel buffer size for subscriptions (default: 16)
	DefaultBufferSize int

	// DefaultPolicy is the default backpressure policy for subscriptions
	DefaultPolicy BackpressurePolicy

	// DefaultBlockTimeout is the default timeout for BlockWithTimeout policy (default: 50ms)
	DefaultBlockTimeout time.Duration

	// WorkerPoolSize limits concurrent async publishes (0 = unlimited, not recommended)
	WorkerPoolSize int

	// MetricsCollector is the unified metrics collector
	MetricsCollector metrics.MetricsCollector

	// Hooks contains lifecycle event callbacks
	Hooks Hooks

	// EnablePanicRecovery enables panic recovery in deliver goroutines (default: true)
	EnablePanicRecovery bool

	// OnPanic is called when a panic is recovered (if EnablePanicRecovery is true)
	OnPanic func(topic string, subID uint64, recovered any)

	// EnableScheduler enables the message scheduler for delayed message delivery.
	// When enabled, PublishDelayed and PublishAt methods become available.
	EnableScheduler bool

	// EnableTTL enables the TTL manager for message expiration tracking.
	// When enabled, PublishWithTTL method becomes available and expired messages
	// are skipped during delivery.
	EnableTTL bool

	// TTLCleanupInterval is how often the TTL manager checks for expired entries (default: 1m).
	// Only used when EnableTTL is true.
	TTLCleanupInterval time.Duration

	// EnableRingBuffer enables ring buffer for all DropOldest policy subscribers.
	// When enabled, DropOldest subscribers use an O(1) circular buffer instead
	// of the channel drain-and-retry loop. Individual subscribers can also opt in
	// via SubOptions.UseRingBuffer.
	EnableRingBuffer bool

	// EnableRequestReply enables the request-reply manager for efficient RPC-style
	// communication over pubsub. When enabled, a shared reply subscription is used
	// instead of creating a temporary subscription per request.
	EnableRequestReply bool

	// EnableHistory enables per-topic message history retention.
	// When enabled, published messages are stored in a circular buffer per topic,
	// allowing late subscribers to catch up on recent messages.
	EnableHistory bool

	// HistoryConfig configures topic history behavior.
	// Only used when EnableHistory is true.
	HistoryConfig HistoryConfig
}

// Hooks contains lifecycle event callbacks.
type Hooks struct {
	// OnSubscribe is called when a new subscription is created
	OnSubscribe func(topic string, subID uint64)

	// OnUnsubscribe is called when a subscription is cancelled
	OnUnsubscribe func(topic string, subID uint64)

	// OnPublish is called before a message is published
	OnPublish func(topic string, msg *Message)

	// OnDeliver is called when a message is delivered to a subscriber
	OnDeliver func(topic string, subID uint64, msg *Message)

	// OnDrop is called when a message is dropped due to backpressure
	OnDrop func(topic string, subID uint64, msg *Message, policy BackpressurePolicy)
}

// Option is a functional option for configuring InProcPubSub.
type Option func(*Config)

// DefaultConfig returns production-ready default configuration.
func DefaultConfig() Config {
	return Config{
		ShardCount:          16,
		DefaultBufferSize:   16,
		DefaultPolicy:       DropOldest,
		DefaultBlockTimeout: 50 * time.Millisecond,
		WorkerPoolSize:      1024,
		EnablePanicRecovery: true,
	}
}

// WithShardCount sets the number of shards for topic partitioning.
// More shards reduce lock contention but increase memory usage.
// Recommended values: 8, 16, 32, 64 (must be power of 2 for best performance).
func WithShardCount(count int) Option {
	return func(c *Config) {
		if count > 0 {
			c.ShardCount = count
		}
	}
}

// WithDefaultBufferSize sets the default channel buffer size for subscriptions.
func WithDefaultBufferSize(size int) Option {
	return func(c *Config) {
		if size > 0 {
			c.DefaultBufferSize = size
		}
	}
}

// WithDefaultPolicy sets the default backpressure policy for subscriptions.
func WithDefaultPolicy(policy BackpressurePolicy) Option {
	return func(c *Config) {
		c.DefaultPolicy = policy
	}
}

// WithDefaultBlockTimeout sets the default timeout for BlockWithTimeout policy.
func WithDefaultBlockTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		if timeout > 0 {
			c.DefaultBlockTimeout = timeout
		}
	}
}

// WithWorkerPoolSize sets the maximum number of concurrent async publish workers.
// Set to 0 for unlimited (not recommended in production).
func WithWorkerPoolSize(size int) Option {
	return func(c *Config) {
		c.WorkerPoolSize = size
	}
}

// WithMetricsCollector sets the unified metrics collector.
func WithMetricsCollector(collector metrics.MetricsCollector) Option {
	return func(c *Config) {
		c.MetricsCollector = collector
	}
}

// WithHooks sets lifecycle event callbacks.
func WithHooks(hooks Hooks) Option {
	return func(c *Config) {
		c.Hooks = hooks
	}
}

// WithPanicRecovery enables or disables panic recovery in deliver goroutines.
func WithPanicRecovery(enable bool) Option {
	return func(c *Config) {
		c.EnablePanicRecovery = enable
	}
}

// WithOnPanic sets the panic handler callback.
func WithOnPanic(handler func(topic string, subID uint64, recovered any)) Option {
	return func(c *Config) {
		c.OnPanic = handler
	}
}

// WithScheduler enables the message scheduler for delayed message delivery.
// This allows using PublishDelayed and PublishAt to schedule messages for future delivery.
func WithScheduler() Option {
	return func(c *Config) {
		c.EnableScheduler = true
	}
}

// WithTTL enables the TTL manager for message expiration tracking.
// Messages published via PublishWithTTL will be automatically skipped during delivery
// if they have expired. The optional cleanupInterval controls how often expired
// tracking entries are cleaned up (default: 1 minute).
func WithTTL(cleanupInterval ...time.Duration) Option {
	return func(c *Config) {
		c.EnableTTL = true
		if len(cleanupInterval) > 0 && cleanupInterval[0] > 0 {
			c.TTLCleanupInterval = cleanupInterval[0]
		}
	}
}

// WithRingBuffer enables ring buffer for all DropOldest policy subscribers.
// When enabled, DropOldest subscribers use an O(1) circular buffer with a pump
// goroutine instead of the channel drain-and-retry loop. This provides better
// performance under high contention and eliminates potential spin-loop issues.
func WithRingBuffer() Option {
	return func(c *Config) {
    c.EnableRingBuffer = true
  }
}

// WithRequestReply enables the request-reply manager for efficient RPC-style
// communication over pubsub. Instead of creating a temporary subscription per
// request, a shared reply subscription is maintained, reducing overhead for
// high-frequency request-reply workloads.
func WithRequestReply() Option {
	return func(c *Config) {
		c.EnableRequestReply = true
	}
}

// WithHistory enables per-topic message history retention.
// Published messages are stored in a circular buffer per topic, allowing
// late subscribers to catch up on recent messages via GetTopicHistory,
// GetTopicHistorySince, and GetRecentMessages methods.
//
// If no config is provided, DefaultHistoryConfig() is used.
func WithHistory(configs ...HistoryConfig) Option {
	return func(c *Config) {
		c.EnableHistory = true
		if len(configs) > 0 {
			c.HistoryConfig = configs[0]
		} else {
			c.HistoryConfig = DefaultHistoryConfig()
		}
	}
}
