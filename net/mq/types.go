package mq

import (
	"context"
	"errors"
	"time"

	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/pubsub"
)

// Configuration constants
const (
	// DefaultBufferSize is the default buffer size for new subscriptions.
	DefaultBufferSize = 16

	// MaxTopicLength is the maximum length of a topic name in characters.
	MaxTopicLength = 1024

	// DefaultHealthCheckInterval is the default interval for health checks.
	DefaultHealthCheckInterval = 30 * time.Second

	// DefaultAckTimeoutDuration is the default timeout for message acknowledgment.
	DefaultAckTimeoutDuration = 30 * time.Second

	// DefaultClusterSyncInterval is the default interval for cluster state synchronization.
	DefaultClusterSyncInterval = 5 * time.Second

	// DefaultTransactionTimeoutDuration is the default timeout for transactions.
	DefaultTransactionTimeoutDuration = 30 * time.Second

	// DefaultMQTTPort is the default port for MQTT protocol.
	DefaultMQTTPort = 1883

	// DefaultAMQPPort is the default port for AMQP protocol.
	DefaultAMQPPort = 5672
)

// Message re-exports pubsub.Message for broker compatibility.
type Message = pubsub.Message

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

// Errors
var (
	// ErrRecoveredPanic is returned when a broker operation recovers from panic.
	ErrRecoveredPanic = errors.New("mq: panic recovered")

	// ErrNotInitialized is returned when the broker is not properly initialized.
	ErrNotInitialized = errors.New("mq: broker not initialized")

	// ErrInvalidTopic is returned when a topic is invalid.
	ErrInvalidTopic = errors.New("mq: invalid topic")

	// ErrNilMessage is returned when attempting to publish a nil message.
	ErrNilMessage = errors.New("mq: message cannot be nil")

	// ErrInvalidConfig is returned when broker configuration is invalid.
	ErrInvalidConfig = errors.New("mq: invalid configuration")

	// ErrBrokerClosed is returned when attempting to use a closed broker.
	ErrBrokerClosed = errors.New("mq: broker is closed")

	// ErrMessageAcknowledged is returned when attempting to acknowledge an already acknowledged message.
	ErrMessageAcknowledged = errors.New("mq: message already acknowledged")

	// ErrMessageNotAcked is returned when a message requires acknowledgment but none was received.
	ErrMessageNotAcked = errors.New("mq: message requires acknowledgment")

	// ErrClusterDisabled is returned when cluster mode is disabled.
	ErrClusterDisabled = errors.New("mq: cluster mode is disabled")

	// ErrNodeNotFound is returned when a cluster node is not found.
	ErrNodeNotFound = errors.New("mq: cluster node not found")

	// ErrTransactionNotSupported is returned when transaction is not supported.
	ErrTransactionNotSupported = errors.New("mq: transaction not supported")

	// ErrDeadLetterNotSupported is returned when dead letter queue is not supported.
	ErrDeadLetterNotSupported = errors.New("mq: dead letter queue not supported")

	// ErrMemoryLimitExceeded is returned when memory usage exceeds the configured limit.
	ErrMemoryLimitExceeded = errors.New("mq: memory limit exceeded")

	// ErrMessageExpired is returned when a message has expired based on its TTL.
	ErrMessageExpired = errors.New("mq: message has expired")

	// ErrTransactionNotFound is returned when a transaction ID is not found.
	ErrTransactionNotFound = errors.New("mq: transaction not found")

	// ErrTransactionTimeout is returned when a transaction exceeds its timeout.
	ErrTransactionTimeout = errors.New("mq: transaction timeout")

	// ErrTransactionCommitted is returned when attempting to use a committed transaction.
	ErrTransactionCommitted = errors.New("mq: transaction already committed")

	// ErrTransactionRolledBack is returned when attempting to use a rolled back transaction.
	ErrTransactionRolledBack = errors.New("mq: transaction already rolled back")
)

// Broker defines the interface for message queue backends.
type Broker interface {
	Publish(ctx context.Context, topic string, msg Message) error
	Subscribe(ctx context.Context, topic string, opts SubOptions) (Subscription, error)
	Close() error
}

// DefaultSubOptions exposes the default subscription settings.
var DefaultSubOptions = pubsub.DefaultSubOptions
