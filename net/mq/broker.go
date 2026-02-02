// Package mq provides an in-process message broker with advanced features.
//
// Features:
//   - Basic pub/sub messaging with topic-based routing
//   - Priority queue support for message ordering
//   - Message TTL (time-to-live) with automatic expiration
//   - Message acknowledgment (ACK/NACK) with timeout and retry
//   - Transaction support for atomic message publishing
//   - Dead letter queue for failed message handling
//   - Persistence storage with automatic recovery
//   - Memory management with configurable limits
//   - Metrics collection and observability
//
// Status:
//   - Core features: Production-ready ✓
//   - Advanced features (transactions, persistence, DLQ): Stable ✓
//   - Protocol support (MQTT, AMQP): Not implemented
//   - Cluster mode: Interface defined, implementation pending
//
// Example usage:
//
//	cfg := mq.DefaultConfig()
//	cfg.EnableTransactions = true
//	cfg.EnablePersistence = true
//	cfg.PersistencePath = "/data/mq"
//	cfg.EnableDeadLetterQueue = true
//	cfg.DeadLetterTopic = "dlq"
//
//	broker := mq.NewInProcBroker(pubsub.New(), mq.WithConfig(cfg))
//	defer broker.Close()
//
//	// Publish with transaction
//	txID := "tx-1"
//	broker.PublishWithTransaction(ctx, "orders", msg1, txID)
//	broker.PublishWithTransaction(ctx, "orders", msg2, txID)
//	broker.CommitTransaction(ctx, txID)
//
//	// Subscribe and process
//	sub, _ := broker.Subscribe(ctx, "orders", mq.SubOptions{BufferSize: 100})
//	for msg := range sub.C() {
//	    // Process message
//	}
package mq

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spcent/plumego/pubsub"
)

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

	priorityMu     sync.Mutex
	priorityQueues map[string]*priorityDispatcher
	prioritySeq    uint64
	priorityClosed atomic.Bool

	ackTracker         *ackTracker
	ttlTracker         *ttlTracker
	txManager          *transactionManager
	deadLetterManager  *deadLetterManager
	persistenceManager *persistenceManager
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

// WithConfig sets the broker configuration.
func WithConfig(cfg Config) Option {
	return func(b *InProcBroker) {
		if err := cfg.Validate(); err != nil {
			panic(fmt.Sprintf("invalid broker config: %v", err))
		}
		b.config = cfg
	}
}

// validateTopic checks if a topic is valid.
func validateTopic(topic string) error {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return fmt.Errorf("%w: cannot be empty", ErrInvalidTopic)
	}
	if len(topic) > MaxTopicLength {
		return fmt.Errorf("%w: topic too long (max %d characters)", ErrInvalidTopic, MaxTopicLength)
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

// validatePublishOperation performs common validation for all publish operations.
// It checks context, broker initialization, topic, and optionally message.
func (b *InProcBroker) validatePublishOperation(ctx context.Context, topic string, msg *Message) error {
	// Validate context
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	// Validate broker initialization
	if b == nil || b.ps == nil {
		return ErrNotInitialized
	}

	// Validate topic
	if err := validateTopic(topic); err != nil {
		return err
	}

	// Validate message if provided
	if msg != nil {
		if err := validateMessage(*msg); err != nil {
			return err
		}
	}

	return nil
}

// validateSubscribeOperation performs common validation for all subscribe operations.
// It checks context, broker initialization, and topic.
func (b *InProcBroker) validateSubscribeOperation(ctx context.Context, topic string) error {
	// Validate context
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	// Validate broker initialization
	if b == nil || b.ps == nil {
		return ErrNotInitialized
	}

	// Validate topic
	if err := validateTopic(topic); err != nil {
		return err
	}

	return nil
}

// validateTTL checks if a TTL message is valid and not expired.
func (b *InProcBroker) validateTTL(expiresAt time.Time) error {
	if expiresAt.IsZero() {
		return nil // No TTL set, message doesn't expire
	}

	if time.Now().After(expiresAt) {
		return fmt.Errorf("%w: message expired at %v", ErrMessageExpired, expiresAt)
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

	// Initialize ackTracker if ACK support is enabled
	if broker.config.EnableAckSupport {
		broker.ackTracker = newAckTracker(broker)
	}

	// Initialize ttlTracker if TTL is enabled (MessageTTL > 0)
	if broker.config.MessageTTL > 0 {
		broker.ttlTracker = newTTLTracker(broker)
	}

	// Initialize txManager if transactions are enabled
	if broker.config.EnableTransactions {
		broker.txManager = newTransactionManager(broker)
	}

	// Initialize deadLetterManager if dead letter queue is enabled
	if broker.config.EnableDeadLetterQueue {
		broker.deadLetterManager = newDeadLetterManager(broker)
	}

	// Initialize persistenceManager if persistence is enabled
	if broker.config.EnablePersistence {
		if broker.config.PersistencePath == "" {
			panic(fmt.Sprintf("invalid broker config: PersistencePath is required when persistence is enabled"))
		}
		backend, err := NewKVPersistence(broker.config.PersistencePath)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize persistence: %v", err))
		}
		broker.persistenceManager = newPersistenceManager(broker, backend)
	}

	return broker
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
		// Validate operation
		if err := b.validatePublishOperation(ctx, topic, &msg); err != nil {
			return err
		}

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
			return err
		}

		// Persist message if persistence is enabled
		if b.config.EnablePersistence && b.persistenceManager != nil {
			if err := b.persistenceManager.saveMessage(ctx, topic, msg); err != nil {
				// Log error but don't fail the publish
				// This ensures availability over consistency
				b.lastError = fmt.Errorf("failed to persist message: %w", err)
			}
		}

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
			return ErrNotInitialized
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
			validMsgs = append(validMsgs, msg)
		}

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
			return err
		}

		// Persist messages if persistence is enabled
		if b.config.EnablePersistence && b.persistenceManager != nil {
			for _, msg := range validMsgs {
				if err := b.persistenceManager.saveMessage(ctx, topic, msg); err != nil {
					// Log error but don't fail the publish
					b.lastError = fmt.Errorf("failed to persist message: %w", err)
				}
			}
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
			return ErrNotInitialized
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
			return ErrNotInitialized
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

// checkMemoryLimit checks if memory usage exceeds the configured limit.
func (b *InProcBroker) checkMemoryLimit() error {
	if b.config.MaxMemoryUsage == 0 {
		return nil // No limit configured
	}

	// Get current memory usage from runtime
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Check if total allocated memory exceeds limit
	if memStats.Alloc > b.config.MaxMemoryUsage {
		return fmt.Errorf("%w: memory usage %d bytes exceeds limit %d bytes",
			ErrMemoryLimitExceeded, memStats.Alloc, b.config.MaxMemoryUsage)
	}

	return nil
}

// GetMemoryUsage returns current memory usage in bytes.
func (b *InProcBroker) GetMemoryUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.Alloc
}

// redeliverMessage re-publishes a message that failed acknowledgment.
func (b *InProcBroker) redeliverMessage(entry *ackEntry) error {
	if b == nil || b.ps == nil {
		return ErrNotInitialized
	}

	// Re-publish the message to the same topic
	ctx := context.Background()
	return b.Publish(ctx, entry.topic, entry.message)
}

// PublishToCluster publishes a message to the cluster (replicates to other nodes).
func (b *InProcBroker) PublishToCluster(ctx context.Context, topic string, msg Message) error {
	return b.executeWithObservability(ctx, OpPublish, topic, func() error {
		// Validate context
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		// Validate broker initialization
		if b == nil || b.ps == nil {
			return ErrNotInitialized
		}

		// Check if cluster mode is enabled
		if !b.config.EnableCluster {
			return fmt.Errorf("%w: cluster mode is disabled", ErrClusterDisabled)
		}

		// Validate topic
		if err := validateTopic(topic); err != nil {
			return err
		}

		// Validate message
		if err := validateMessage(msg); err != nil {
			return err
		}

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
			return err
		}

		// TODO: Implement cluster replication logic
		// This would:
		// 1. Replicate message to other cluster nodes
		// 2. Wait for acknowledgments from replicas
		// 3. Handle replication failures
		// 4. Maintain consistency across nodes

		// For now, just publish locally
		return b.ps.Publish(topic, msg)
	})
}

// SubscribeFromCluster subscribes to messages from the cluster.
func (b *InProcBroker) SubscribeFromCluster(ctx context.Context, topic string, opts SubOptions) (Subscription, error) {
	var sub Subscription
	err := b.executeWithObservability(ctx, OpSubscribe, topic, func() error {
		// Validate context
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		// Validate broker initialization
		if b == nil || b.ps == nil {
			return ErrNotInitialized
		}

		// Check if cluster mode is enabled
		if !b.config.EnableCluster {
			return fmt.Errorf("%w: cluster mode is disabled", ErrClusterDisabled)
		}

		// Validate topic
		if err := validateTopic(topic); err != nil {
			return err
		}

		// Subscribe locally
		subscription, err := b.ps.Subscribe(topic, opts)
		if err != nil {
			return err
		}

		sub = subscription
		return nil
	})

	return sub, err
}

// StartMQTTServer starts the MQTT protocol server.
func (b *InProcBroker) StartMQTTServer() error {
	if b == nil {
		return ErrNotInitialized
	}

	if !b.config.EnableMQTT {
		return fmt.Errorf("%w: MQTT support is disabled", ErrInvalidConfig)
	}

	// TODO: Implement MQTT server
	// This would:
	// 1. Start MQTT broker on configured port
	// 2. Handle MQTT protocol (CONNECT, PUBLISH, SUBSCRIBE, etc.)
	// 3. Bridge MQTT messages to internal pubsub

	return nil
}

// StartAMQPServer starts the AMQP protocol server.
func (b *InProcBroker) StartAMQPServer() error {
	if b == nil {
		return ErrNotInitialized
	}

	if !b.config.EnableAMQP {
		return fmt.Errorf("%w: AMQP support is disabled", ErrInvalidConfig)
	}

	// TODO: Implement AMQP server
	// This would:
	// 1. Start AMQP broker on configured port
	// 2. Handle AMQP protocol (channel, exchange, queue, etc.)
	// 3. Bridge AMQP messages to internal pubsub

	return nil
}

// RecoverMessages recovers persisted messages for a topic.
// This is useful for replaying messages after broker restart.
func (b *InProcBroker) RecoverMessages(ctx context.Context, topic string, limit int) ([]Message, error) {
	if b == nil || b.ps == nil {
		return nil, ErrNotInitialized
	}

	if !b.config.EnablePersistence {
		return nil, fmt.Errorf("persistence is not enabled")
	}

	if b.persistenceManager == nil {
		return nil, fmt.Errorf("%w: persistence manager not initialized", ErrNotInitialized)
	}

	return b.persistenceManager.loadMessages(ctx, topic, limit)
}

// ReplayMessages replays persisted messages to subscribers.
// This is useful for recovering messages after broker restart.
func (b *InProcBroker) ReplayMessages(ctx context.Context, topic string, limit int) error {
	messages, err := b.RecoverMessages(ctx, topic, limit)
	if err != nil {
		return err
	}

	// Republish each message
	for _, msg := range messages {
		if err := b.Publish(ctx, topic, msg); err != nil {
			return fmt.Errorf("failed to replay message %s: %w", msg.ID, err)
		}
	}

	return nil
}

// Close shuts down the broker.
func (b *InProcBroker) Close() error {
	return b.executeWithObservability(context.Background(), OpClose, "", func() error {
		// Validate broker initialization
		if b == nil || b.ps == nil {
			return nil // Close is idempotent
		}
		b.closePriorityDispatchers()

		// Close ack tracker if it exists
		if b.ackTracker != nil {
			b.ackTracker.close()
		}

		// Close TTL tracker if it exists
		if b.ttlTracker != nil {
			b.ttlTracker.close()
		}

		// Close transaction manager if it exists
		if b.txManager != nil {
			b.txManager.close()
		}

		// Close dead letter manager if it exists
		if b.deadLetterManager != nil {
			b.deadLetterManager.close()
		}

		// Close persistence manager if it exists
		if b.persistenceManager != nil {
			if err := b.persistenceManager.close(); err != nil {
				b.lastError = fmt.Errorf("failed to close persistence: %w", err)
			}
		}

		return b.ps.Close()
	})
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
