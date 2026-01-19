package mq

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/pubsub"
)

// Message re-exports pubsub.Message for broker compatibility.
type Message = pubsub.Message

// TTLMessage extends Message with TTL support.
type TTLMessage struct {
	Message
	// ExpiresAt is the timestamp when the message expires
	ExpiresAt time.Time
}

// SubOptions re-exports pubsub.SubOptions for broker compatibility.
type SubOptions = pubsub.SubOptions

// Subscription re-exports pubsub.Subscription for broker compatibility.
type Subscription = pubsub.Subscription

// Operation labels broker actions for observability.
type Operation string

const (
	OpPublish   Operation = "publish"
	OpSubscribe Operation = "subscribe"
	OpClose     Operation = "close"
	OpMetrics   Operation = "metrics"
)

// Metrics captures timing and error information for broker actions.
type Metrics struct {
	Operation Operation
	Topic     string
	Duration  time.Duration
	Err       error
	Panic     bool
}

// MetricsCollector can be plugged into the broker to observe activity.
// This is now an alias for the unified metrics collector
type MetricsCollector = metrics.MetricsCollector

// PanicHandler is invoked when a broker operation panics.
type PanicHandler func(ctx context.Context, op Operation, recovered any)

// ErrRecoveredPanic is returned when a broker operation recovers from panic.
var ErrRecoveredPanic = errors.New("mq: panic recovered")

// ErrNotInitialized is returned when the broker is not properly initialized.
var ErrNotInitialized = errors.New("mq: broker not initialized")

// ErrInvalidTopic is returned when a topic is invalid.
var ErrInvalidTopic = errors.New("mq: invalid topic")

// ErrNilMessage is returned when attempting to publish a nil message.
var ErrNilMessage = errors.New("mq: message cannot be nil")

// ErrInvalidConfig is returned when broker configuration is invalid.
var ErrInvalidConfig = errors.New("mq: invalid configuration")

// ErrBrokerClosed is returned when attempting to use a closed broker.
var ErrBrokerClosed = errors.New("mq: broker is closed")

// Broker defines the interface for message queue backends.
type Broker interface {
	Publish(ctx context.Context, topic string, msg Message) error
	Subscribe(ctx context.Context, topic string, opts SubOptions) (Subscription, error)
	Close() error
}

// Config holds broker configuration.
type Config struct {
	// EnableHealthCheck enables health check endpoint
	EnableHealthCheck bool

	// MaxTopics limits the number of topics (0 = no limit)
	MaxTopics int

	// MaxSubscribers limits the number of subscribers per topic (0 = no limit)
	MaxSubscribers int

	// DefaultBufferSize is the default buffer size for new subscriptions
	DefaultBufferSize int

	// EnableMetrics enables metrics collection
	EnableMetrics bool

	// HealthCheckInterval is the interval for health checks (default: 30s)
	HealthCheckInterval time.Duration

	// MessageTTL is the default time-to-live for messages (0 = no TTL)
	MessageTTL time.Duration
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		EnableHealthCheck:   true,
		MaxTopics:           0, // No limit
		MaxSubscribers:      0, // No limit
		DefaultBufferSize:   16,
		EnableMetrics:       true,
		HealthCheckInterval: 30 * time.Second,
		MessageTTL:          0, // No TTL by default
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.DefaultBufferSize <= 0 {
		return fmt.Errorf("%w: DefaultBufferSize must be positive", ErrInvalidConfig)
	}
	if c.HealthCheckInterval < 0 {
		return fmt.Errorf("%w: HealthCheckInterval cannot be negative", ErrInvalidConfig)
	}
	if c.MaxTopics < 0 {
		return fmt.Errorf("%w: MaxTopics cannot be negative", ErrInvalidConfig)
	}
	if c.MaxSubscribers < 0 {
		return fmt.Errorf("%w: MaxSubscribers cannot be negative", ErrInvalidConfig)
	}
	return nil
}

// HealthStatus represents the health status of the broker.
type HealthStatus struct {
	Status      string          `json:"status"`
	Timestamp   time.Time       `json:"timestamp"`
	Uptime      string          `json:"uptime"`
	TotalTopics int             `json:"total_topics"`
	TotalSubs   int             `json:"total_subscribers"`
	MemoryUsage uint64          `json:"memory_usage,omitempty"`
	Metrics     MetricsSnapshot `json:"metrics,omitempty"`
}

// MetricsSnapshot represents a snapshot of broker metrics.
type MetricsSnapshot struct {
	TotalPublished uint64        `json:"total_published"`
	TotalDelivered uint64        `json:"total_delivered"`
	TotalDropped   uint64        `json:"total_dropped"`
	ActiveTopics   int           `json:"active_topics"`
	ActiveSubs     int           `json:"active_subscribers"`
	AverageLatency time.Duration `json:"average_latency"`
	LastError      string        `json:"last_error,omitempty"`
	LastPanic      string        `json:"last_panic,omitempty"`
	LastPanicTime  time.Time     `json:"last_panic_time,omitempty"`
}

// InProcBroker adapts pubsub.PubSub to the Broker interface.
type InProcBroker struct {
	ps            pubsub.PubSub
	metrics       MetricsCollector
	panicHandler  PanicHandler
	config        Config
	startTime     time.Time
	lastError     error
	lastPanic     error
	lastPanicTime time.Time
}

// Option configures the broker.
type Option func(*InProcBroker)

// WithMetricsCollector registers a metrics collector.
func WithMetricsCollector(collector MetricsCollector) Option {
	return func(b *InProcBroker) {
		b.metrics = collector
	}
}

// WithPanicHandler registers a panic handler.
func WithPanicHandler(handler PanicHandler) Option {
	return func(b *InProcBroker) {
		b.panicHandler = handler
	}
}

// validateTopic checks if a topic is valid.
func validateTopic(topic string) error {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return fmt.Errorf("%w: cannot be empty", ErrInvalidTopic)
	}
	if len(topic) > 1024 {
		return fmt.Errorf("%w: topic too long (max 1024 characters)", ErrInvalidTopic)
	}
	return nil
}

// validateMessage checks if a message is valid.
func validateMessage(msg Message) error {
	if msg.ID == "" {
		return fmt.Errorf("%w: ID is required", ErrNilMessage)
	}
	return nil
}

// NewInProcBroker wraps the in-process pubsub implementation.
func NewInProcBroker(ps pubsub.PubSub, opts ...Option) *InProcBroker {
	if ps == nil {
		ps = pubsub.New()
	}
	broker := &InProcBroker{
		ps:        ps,
		config:    DefaultConfig(),
		startTime: time.Now(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(broker)
		}
	}
	return broker
}

// WithConfig sets the broker configuration.
func WithConfig(cfg Config) Option {
	return func(b *InProcBroker) {
		if err := cfg.Validate(); err != nil {
			panic(fmt.Sprintf("invalid broker config: %v", err))
		}
		b.config = cfg
	}
}

// executeWithObservability wraps an operation with observability logic.
func (b *InProcBroker) executeWithObservability(
	ctx context.Context,
	op Operation,
	topic string,
	fn func() error,
) (err error) {
	start := time.Now()
	panicked := false
	defer func() {
		if recovered := recover(); recovered != nil {
			panicked = true
			err = b.handlePanic(ctx, op, recovered)
		}
		b.observe(ctx, op, topic, start, err, panicked)
	}()
	return fn()
}

// Publish sends a message to a topic.
func (b *InProcBroker) Publish(ctx context.Context, topic string, msg Message) error {
	return b.executeWithObservability(ctx, OpPublish, topic, func() error {
		// Validate context
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		// Validate broker initialization
		if b == nil || b.ps == nil {
			return fmt.Errorf("%w", ErrNotInitialized)
		}

		// Validate topic
		if err := validateTopic(topic); err != nil {
			return err
		}

		// Validate message
		if err := validateMessage(msg); err != nil {
			return err
		}

		// Check TTL if message has expiration
		// Note: TTLMessage is a wrapper, so we need to check the underlying type
		// For now, we'll skip TTL check for regular Message types

		return b.ps.Publish(topic, msg)
	})
}

// PublishBatch sends multiple messages to a topic in a single operation.
func (b *InProcBroker) PublishBatch(ctx context.Context, topic string, msgs []Message) error {
	return b.executeWithObservability(ctx, OpPublish, topic, func() error {
		// Validate context
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		// Validate broker initialization
		if b == nil || b.ps == nil {
			return fmt.Errorf("%w", ErrNotInitialized)
		}

		// Validate topic
		if err := validateTopic(topic); err != nil {
			return err
		}

		// Validate and filter messages
		validMsgs := make([]Message, 0, len(msgs))
		for _, msg := range msgs {
			if err := validateMessage(msg); err != nil {
				return err
			}

			// Note: TTL check would require type assertion on TTLMessage
			// For regular Message types, we skip TTL validation
			validMsgs = append(validMsgs, msg)
		}

		// Publish all valid messages
		for _, msg := range validMsgs {
			if err := b.ps.Publish(topic, msg); err != nil {
				return err
			}
		}

		return nil
	})
}

// SubscribeBatch subscribes to multiple topics at once.
func (b *InProcBroker) SubscribeBatch(ctx context.Context, topics []string, opts SubOptions) ([]Subscription, error) {
	var subs []Subscription

	err := b.executeWithObservability(ctx, OpSubscribe, "", func() error {
		// Validate context
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		// Validate broker initialization
		if b == nil || b.ps == nil {
			return fmt.Errorf("%w", ErrNotInitialized)
		}

		// Subscribe to each topic
		for _, topic := range topics {
			if err := validateTopic(topic); err != nil {
				return err
			}

			sub, err := b.ps.Subscribe(topic, opts)
			if err != nil {
				return err
			}

			subs = append(subs, sub)
		}

		return nil
	})

	if err != nil {
		// Clean up any subscriptions that were created
		for _, sub := range subs {
			sub.Cancel()
		}
		return nil, err
	}

	return subs, nil
}

// Subscribe registers a subscription for a topic.
func (b *InProcBroker) Subscribe(ctx context.Context, topic string, opts SubOptions) (sub Subscription, err error) {
	err = b.executeWithObservability(ctx, OpSubscribe, topic, func() error {
		// Validate context
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		// Validate broker initialization
		if b == nil || b.ps == nil {
			return fmt.Errorf("%w", ErrNotInitialized)
		}

		// Validate topic
		if err := validateTopic(topic); err != nil {
			return err
		}

		// Subscribe
		var subscription Subscription
		subscription, err = b.ps.Subscribe(topic, opts)
		if err != nil {
			return err
		}

		// Store subscription for return
		sub = subscription
		return nil
	})

	return sub, err
}

// Close shuts down the broker.
func (b *InProcBroker) Close() error {
	return b.executeWithObservability(context.Background(), OpClose, "", func() error {
		// Validate broker initialization
		if b == nil || b.ps == nil {
			return nil // Close is idempotent
		}
		return b.ps.Close()
	})
}

// DefaultSubOptions exposes the default subscription settings.
var DefaultSubOptions = pubsub.DefaultSubOptions

// HealthCheck returns the current health status of the broker.
func (b *InProcBroker) HealthCheck() HealthStatus {
	if b == nil || b.ps == nil {
		return HealthStatus{
			Status:    "unhealthy",
			Timestamp: time.Now(),
		}
	}

	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(b.startTime).String(),
	}

	// Get topic and subscriber counts
	if snapper, ok := b.ps.(interface {
		ListTopics() []string
		GetSubscriberCount(topic string) int
	}); ok {
		topics := snapper.ListTopics()
		status.TotalTopics = len(topics)
		for _, topic := range topics {
			status.TotalSubs += snapper.GetSubscriberCount(topic)
		}
	}

	// Get metrics snapshot
	if b.config.EnableMetrics {
		status.Metrics = b.getMetricsSnapshot()
	}

	return status
}

// UpdateConfig dynamically updates the broker configuration.
func (b *InProcBroker) UpdateConfig(cfg Config) error {
	if b == nil {
		return fmt.Errorf("%w", ErrNotInitialized)
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	b.config = cfg
	return nil
}

// GetConfig returns the current broker configuration.
func (b *InProcBroker) GetConfig() Config {
	if b == nil {
		return Config{}
	}
	return b.config
}

// getMetricsSnapshot creates a snapshot of current metrics.
func (b *InProcBroker) getMetricsSnapshot() MetricsSnapshot {
	snapshot := MetricsSnapshot{}

	// Get pubsub metrics if available
	if snapper, ok := b.ps.(interface{ Snapshot() pubsub.MetricsSnapshot }); ok {
		pubsubSnap := snapper.Snapshot()

		// Aggregate metrics from all topics
		var totalPublished, totalDelivered, totalDropped uint64
		var activeSubs int

		for _, topicMetrics := range pubsubSnap.Topics {
			totalPublished += topicMetrics.PublishTotal
			totalDelivered += topicMetrics.DeliveredTotal

			// Sum all dropped counts
			for _, dropped := range topicMetrics.DroppedByPolicy {
				totalDropped += dropped
			}

			activeSubs += topicMetrics.SubscribersGauge
		}

		snapshot.TotalPublished = totalPublished
		snapshot.TotalDelivered = totalDelivered
		snapshot.TotalDropped = totalDropped
		snapshot.ActiveTopics = len(pubsubSnap.Topics)
		snapshot.ActiveSubs = activeSubs
	}

	// Add error and panic information
	if b.lastError != nil {
		snapshot.LastError = b.lastError.Error()
	}
	if b.lastPanic != nil {
		snapshot.LastPanic = b.lastPanic.Error()
		snapshot.LastPanicTime = b.lastPanicTime
	}

	return snapshot
}

// Snapshot exposes in-process pubsub metrics when supported.
func (b *InProcBroker) Snapshot() pubsub.MetricsSnapshot {
	if b == nil || b.ps == nil {
		return pubsub.MetricsSnapshot{}
	}
	if snapper, ok := b.ps.(interface{ Snapshot() pubsub.MetricsSnapshot }); ok {
		return snapper.Snapshot()
	}
	return pubsub.MetricsSnapshot{}
}

func (b *InProcBroker) handlePanic(ctx context.Context, op Operation, recovered any) error {
	if b != nil && b.panicHandler != nil {
		func() {
			defer func() {
				_ = recover()
			}()
			b.panicHandler(ctx, op, recovered)
		}()
	}
	return fmt.Errorf("%w: %s: %v", ErrRecoveredPanic, op, recovered)
}

func (b *InProcBroker) observe(ctx context.Context, op Operation, topic string, start time.Time, err error, panicked bool) {
	if b == nil || b.metrics == nil {
		return
	}

	duration := time.Since(start)

	func() {
		defer func() {
			if recovered := recover(); recovered != nil && b.panicHandler != nil {
				b.panicHandler(ctx, OpMetrics, recovered)
			}
		}()
		// Use the unified interface
		b.metrics.ObserveMQ(ctx, string(op), topic, duration, err, panicked)
	}()
}
