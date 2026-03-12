package pubsub

import (
	"time"

	"github.com/spcent/plumego/metrics"
)

// Config holds the runtime configuration for InProcBroker.
// It contains only operational parameters — capability features are provided
// by wrapping the broker with the corresponding sub-package constructors.
type Config struct {
	// ShardCount is the number of topic shards (default: 16).
	// Higher values reduce lock contention at the cost of memory.
	// Recommended: power of 2 (8, 16, 32, 64).
	ShardCount int

	// DefaultBufferSize is the per-subscription channel buffer size (default: 16).
	DefaultBufferSize int

	// DefaultPolicy is the backpressure policy applied when SubOptions.Policy
	// is the zero value (default: DropOldest).
	DefaultPolicy BackpressurePolicy

	// DefaultBlockTimeout is the default timeout for the BlockWithTimeout policy
	// (default: 50ms).
	DefaultBlockTimeout time.Duration

	// WorkerPoolSize limits concurrent async publishes (default: 1024).
	// Set to 0 for unlimited (not recommended in production).
	WorkerPoolSize int

	// MetricsCollector is the external metrics sink.
	MetricsCollector metrics.PubSubObserver

	// EnablePanicRecovery wraps the delivery goroutine in a recover() guard
	// (default: true).
	EnablePanicRecovery bool

	// OnPanic is called when a delivery panic is recovered.
	// Called only when EnablePanicRecovery is true.
	OnPanic func(topic string, subID uint64, recovered any)

	// observers holds registered BrokerObserver instances (populated by WithObserver).
	observers []BrokerObserver
}

// Option is a functional option for configuring InProcBroker.
type Option func(*Config)

// DefaultConfig returns production-ready defaults.
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
func WithShardCount(count int) Option {
	return func(c *Config) {
		if count > 0 {
			c.ShardCount = count
		}
	}
}

// WithDefaultBufferSize sets the default subscription channel buffer size.
func WithDefaultBufferSize(size int) Option {
	return func(c *Config) {
		if size > 0 {
			c.DefaultBufferSize = size
		}
	}
}

// WithDefaultPolicy sets the default backpressure policy.
func WithDefaultPolicy(policy BackpressurePolicy) Option {
	return func(c *Config) {
		c.DefaultPolicy = policy
	}
}

// WithDefaultBlockTimeout sets the default timeout for the BlockWithTimeout policy.
func WithDefaultBlockTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		if timeout > 0 {
			c.DefaultBlockTimeout = timeout
		}
	}
}

// WithWorkerPoolSize sets the maximum number of concurrent async-publish workers.
func WithWorkerPoolSize(size int) Option {
	return func(c *Config) {
		c.WorkerPoolSize = size
	}
}

// WithMetricsCollector sets the external metrics sink.
func WithMetricsCollector(collector metrics.PubSubObserver) Option {
	return func(c *Config) {
		c.MetricsCollector = collector
	}
}

// WithPanicRecovery enables or disables panic recovery in delivery goroutines.
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

// WithObserver registers a BrokerObserver.
// Multiple observers may be registered; they are called in registration order.
func WithObserver(o BrokerObserver) Option {
	return func(c *Config) {
		c.observers = append(c.observers, o)
	}
}
