// Package mq provides an in-process message broker.
//
// Experimental: this module includes incomplete features (see TODOs) and may change
// without notice. Avoid production use until the TODOs are fully implemented.
package mq

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
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

// ErrMessageAcknowledged is returned when attempting to acknowledge an already acknowledged message.
var ErrMessageAcknowledged = errors.New("mq: message already acknowledged")

// ErrMessageNotAcked is returned when a message requires acknowledgment but none was received.
var ErrMessageNotAcked = errors.New("mq: message requires acknowledgment")

// ErrClusterDisabled is returned when cluster mode is disabled.
var ErrClusterDisabled = errors.New("mq: cluster mode is disabled")

// ErrNodeNotFound is returned when a cluster node is not found.
var ErrNodeNotFound = errors.New("mq: cluster node not found")

// ErrTransactionNotSupported is returned when transaction is not supported.
var ErrTransactionNotSupported = errors.New("mq: transaction not supported")

// ErrDeadLetterNotSupported is returned when dead letter queue is not supported.
var ErrDeadLetterNotSupported = errors.New("mq: dead letter queue not supported")

// MessagePriority represents the priority of a message.
type MessagePriority int

const (
	PriorityLowest  MessagePriority = 0
	PriorityLow     MessagePriority = 10
	PriorityNormal  MessagePriority = 20
	PriorityHigh    MessagePriority = 30
	PriorityHighest MessagePriority = 40
)

// AckPolicy defines the acknowledgment policy for messages.
type AckPolicy int

const (
	// AckNone - No acknowledgment required
	AckNone AckPolicy = iota
	// AckRequired - Message requires explicit acknowledgment
	AckRequired
	// AckTimeout - Message requires acknowledgment within timeout
	AckTimeout
)

// PriorityMessage extends Message with priority support.
type PriorityMessage struct {
	Message
	Priority MessagePriority
}

type priorityEnvelope struct {
	msg      Message
	priority MessagePriority
	seq      uint64
	ctx      context.Context
	done     chan error
}

type priorityQueue []*priorityEnvelope

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].priority == pq[j].priority {
		return pq[i].seq < pq[j].seq
	}
	return pq[i].priority > pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *priorityQueue) Push(x any) {
	*pq = append(*pq, x.(*priorityEnvelope))
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

type priorityDispatcher struct {
	broker *InProcBroker
	topic  string

	mu     sync.Mutex
	cond   *sync.Cond
	queue  priorityQueue
	closed bool
}

func newPriorityDispatcher(b *InProcBroker, topic string) *priorityDispatcher {
	d := &priorityDispatcher{
		broker: b,
		topic:  topic,
	}
	d.cond = sync.NewCond(&d.mu)
	heap.Init(&d.queue)
	go d.run()
	return d
}

func (d *priorityDispatcher) enqueue(env *priorityEnvelope) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return ErrBrokerClosed
	}
	heap.Push(&d.queue, env)
	d.cond.Signal()
	return nil
}

func (d *priorityDispatcher) close() {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return
	}
	d.closed = true
	pending := make([]*priorityEnvelope, len(d.queue))
	copy(pending, d.queue)
	d.queue = nil
	d.cond.Broadcast()
	d.mu.Unlock()

	for _, env := range pending {
		if env.done != nil {
			env.done <- ErrBrokerClosed
			close(env.done)
		}
	}
}

func (d *priorityDispatcher) run() {
	for {
		d.mu.Lock()
		for !d.closed && len(d.queue) == 0 {
			d.cond.Wait()
		}
		if d.closed {
			d.mu.Unlock()
			return
		}
		env := heap.Pop(&d.queue).(*priorityEnvelope)
		d.mu.Unlock()

		err := d.publish(env)
		if env.done != nil {
			env.done <- err
			close(env.done)
		}
	}
}

func (d *priorityDispatcher) publish(env *priorityEnvelope) error {
	if env.ctx != nil {
		if err := env.ctx.Err(); err != nil {
			return err
		}
	}
	if d.broker == nil || d.broker.ps == nil {
		return ErrNotInitialized
	}
	return d.broker.ps.Publish(d.topic, env.msg)
}

// AckMessage extends Message with acknowledgment support.
type AckMessage struct {
	Message
	AckID      string
	AckPolicy  AckPolicy
	AckTimeout time.Duration
}

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

	// EnablePriorityQueue enables priority queue support
	EnablePriorityQueue bool

	// EnableAckSupport enables message acknowledgment support
	EnableAckSupport bool

	// DefaultAckTimeout is the default timeout for acknowledgment (default: 30s)
	DefaultAckTimeout time.Duration

	// MaxMemoryUsage limits memory usage in bytes (0 = no limit)
	MaxMemoryUsage uint64

	// EnableTriePattern enables Trie-based pattern matching
	EnableTriePattern bool

	// EnableCluster enables distributed cluster mode
	EnableCluster bool

	// ClusterNodeID is the unique identifier for this node in the cluster
	ClusterNodeID string

	// ClusterNodes is the list of peer nodes in the cluster (format: "node-id@host:port")
	ClusterNodes []string

	// ClusterReplicationFactor is the number of replicas for each message (default: 1)
	ClusterReplicationFactor int

	// ClusterSyncInterval is the interval for cluster state synchronization (default: 5s)
	ClusterSyncInterval time.Duration

	// EnablePersistence enables persistent storage backend
	EnablePersistence bool

	// PersistencePath is the directory path for persistent storage
	PersistencePath string

	// EnableDeadLetterQueue enables dead letter queue support
	EnableDeadLetterQueue bool

	// DeadLetterTopic is the topic for dead letter messages
	DeadLetterTopic string

	// EnableTransactions enables transaction support
	EnableTransactions bool

	// TransactionTimeout is the timeout for transactions (default: 30s)
	TransactionTimeout time.Duration

	// EnableMQTT enables MQTT protocol support
	EnableMQTT bool

	// MQTTPort is the port for MQTT protocol (default: 1883)
	MQTTPort int

	// EnableAMQP enables AMQP protocol support
	EnableAMQP bool

	// AMQPPort is the port for AMQP protocol (default: 5672)
	AMQPPort int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		EnableHealthCheck:        true,
		MaxTopics:                0, // No limit
		MaxSubscribers:           0, // No limit
		DefaultBufferSize:        16,
		EnableMetrics:            true,
		HealthCheckInterval:      30 * time.Second,
		MessageTTL:               0, // No TTL by default
		EnablePriorityQueue:      true,
		EnableAckSupport:         false, // Disabled by default for backward compatibility
		DefaultAckTimeout:        30 * time.Second,
		MaxMemoryUsage:           0,     // No limit by default
		EnableTriePattern:        false, // Disabled by default for backward compatibility
		EnableCluster:            false, // Disabled by default
		ClusterReplicationFactor: 1,     // No replication by default
		ClusterSyncInterval:      5 * time.Second,
		EnablePersistence:        false, // Disabled by default
		EnableDeadLetterQueue:    false, // Disabled by default
		EnableTransactions:       false, // Disabled by default
		TransactionTimeout:       30 * time.Second,
		EnableMQTT:               false, // Disabled by default
		MQTTPort:                 1883,
		EnableAMQP:               false, // Disabled by default
		AMQPPort:                 5672,
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
	MemoryLimit uint64          `json:"memory_limit,omitempty"`
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

// ClusterStatus represents the status of the cluster.
type ClusterStatus struct {
	Status            string        `json:"status"`
	NodeID            string        `json:"node_id,omitempty"`
	Peers             []string      `json:"peers,omitempty"`
	ReplicationFactor int           `json:"replication_factor,omitempty"`
	SyncInterval      time.Duration `json:"sync_interval,omitempty"`
	TotalNodes        int           `json:"total_nodes,omitempty"`
	HealthyNodes      int           `json:"healthy_nodes,omitempty"`
	LastSyncTime      time.Time     `json:"last_sync_time,omitempty"`
}

// DeadLetterStats represents statistics about the dead letter queue.
type DeadLetterStats struct {
	Enabled         bool      `json:"enabled"`
	Topic           string    `json:"topic,omitempty"`
	TotalMessages   uint64    `json:"total_messages,omitempty"`
	CurrentCount    int       `json:"current_count,omitempty"`
	LastMessageTime time.Time `json:"last_message_time,omitempty"`
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

	priorityMu     sync.Mutex
	priorityQueues map[string]*priorityDispatcher
	prioritySeq    uint64
	priorityClosed atomic.Bool
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

func (b *InProcBroker) ensurePriorityDispatcher(topic string) (*priorityDispatcher, error) {
	if b == nil || b.ps == nil {
		return nil, ErrNotInitialized
	}
	if b.priorityClosed.Load() {
		return nil, ErrBrokerClosed
	}

	b.priorityMu.Lock()
	defer b.priorityMu.Unlock()

	if b.priorityClosed.Load() {
		return nil, ErrBrokerClosed
	}

	if b.priorityQueues == nil {
		b.priorityQueues = make(map[string]*priorityDispatcher)
	}

	dispatcher := b.priorityQueues[topic]
	if dispatcher == nil {
		dispatcher = newPriorityDispatcher(b, topic)
		b.priorityQueues[topic] = dispatcher
	}

	return dispatcher, nil
}

func (b *InProcBroker) closePriorityDispatchers() {
	if b == nil {
		return
	}
	if !b.priorityClosed.CompareAndSwap(false, true) {
		return
	}

	b.priorityMu.Lock()
	dispatchers := make([]*priorityDispatcher, 0, len(b.priorityQueues))
	for _, dispatcher := range b.priorityQueues {
		dispatchers = append(dispatchers, dispatcher)
	}
	b.priorityQueues = nil
	b.priorityMu.Unlock()

	for _, dispatcher := range dispatchers {
		dispatcher.close()
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

// PublishPriority publishes a message with priority to a topic.
func (b *InProcBroker) PublishPriority(ctx context.Context, topic string, msg PriorityMessage) error {
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
		if err := validateMessage(msg.Message); err != nil {
			return err
		}

		// Check if priority queue is enabled
		if !b.config.EnablePriorityQueue {
			// Fall back to regular publish
			return b.ps.Publish(topic, msg.Message)
		}

		dispatcher, err := b.ensurePriorityDispatcher(topic)
		if err != nil {
			return err
		}

		env := &priorityEnvelope{
			msg:      msg.Message,
			priority: msg.Priority,
			seq:      atomic.AddUint64(&b.prioritySeq, 1),
			ctx:      ctx,
			done:     make(chan error, 1),
		}

		if err := dispatcher.enqueue(env); err != nil {
			return err
		}

		if ctx == nil {
			return <-env.done
		}

		select {
		case err := <-env.done:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}

// PublishWithAck publishes a message that requires acknowledgment.
func (b *InProcBroker) PublishWithAck(ctx context.Context, topic string, msg AckMessage) error {
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
		if err := validateMessage(msg.Message); err != nil {
			return err
		}

		// Check if acknowledgment support is enabled
		if !b.config.EnableAckSupport {
			return fmt.Errorf("%w: acknowledgment support is disabled", ErrInvalidConfig)
		}

		// Set default timeout if not specified
		if msg.AckTimeout == 0 {
			msg.AckTimeout = b.config.DefaultAckTimeout
		}

		// TODO: Implement acknowledgment logic
		// For now, just publish the message
		return b.ps.Publish(topic, msg.Message)
	})
}

// SubscribeWithAck subscribes to a topic with acknowledgment support.
func (b *InProcBroker) SubscribeWithAck(ctx context.Context, topic string, opts SubOptions) (Subscription, error) {
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
			return fmt.Errorf("%w", ErrNotInitialized)
		}

		// Validate topic
		if err := validateTopic(topic); err != nil {
			return err
		}

		// Check if acknowledgment support is enabled
		if !b.config.EnableAckSupport {
			return fmt.Errorf("%w: acknowledgment support is disabled", ErrInvalidConfig)
		}

		// Subscribe
		subscription, err := b.ps.Subscribe(topic, opts)
		if err != nil {
			return err
		}

		sub = subscription
		return nil
	})

	return sub, err
}

// Ack acknowledges a message by ID.
func (b *InProcBroker) Ack(ctx context.Context, topic string, messageID string) error {
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

		// Check if acknowledgment support is enabled
		if !b.config.EnableAckSupport {
			return fmt.Errorf("%w: acknowledgment support is disabled", ErrInvalidConfig)
		}

		// TODO: Implement acknowledgment tracking
		// This would track which messages have been acknowledged
		return nil
	})
}

// Nack negatively acknowledges a message by ID (request re-delivery).
func (b *InProcBroker) Nack(ctx context.Context, topic string, messageID string) error {
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

		// Check if acknowledgment support is enabled
		if !b.config.EnableAckSupport {
			return fmt.Errorf("%w: acknowledgment support is disabled", ErrInvalidConfig)
		}

		// TODO: Implement negative acknowledgment logic
		// This would request re-delivery of the message
		return nil
	})
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
			ErrInvalidConfig, memStats.Alloc, b.config.MaxMemoryUsage)
	}

	return nil
}

// GetMemoryUsage returns current memory usage in bytes.
func (b *InProcBroker) GetMemoryUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.Alloc
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
			return fmt.Errorf("%w", ErrNotInitialized)
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
			return fmt.Errorf("%w", ErrNotInitialized)
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

// GetClusterStatus returns the current cluster status.
func (b *InProcBroker) GetClusterStatus() ClusterStatus {
	if b == nil || !b.config.EnableCluster {
		return ClusterStatus{
			Status: "disabled",
		}
	}

	return ClusterStatus{
		Status:            "active",
		NodeID:            b.config.ClusterNodeID,
		Peers:             b.config.ClusterNodes,
		ReplicationFactor: b.config.ClusterReplicationFactor,
		SyncInterval:      b.config.ClusterSyncInterval,
	}
}

// PublishWithTransaction publishes a message within a transaction.
func (b *InProcBroker) PublishWithTransaction(ctx context.Context, topic string, msg Message, txID string) error {
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

		// Check if transaction support is enabled
		if !b.config.EnableTransactions {
			return fmt.Errorf("%w: transaction support is disabled", ErrTransactionNotSupported)
		}

		// Validate topic
		if err := validateTopic(topic); err != nil {
			return err
		}

		// Validate message
		if err := validateMessage(msg); err != nil {
			return err
		}

		// TODO: Implement transaction logic
		// This would:
		// 1. Start a transaction with txID
		// 2. Buffer the message until commit
		// 3. Support rollback on failure
		// 4. Handle transaction timeout

		// For now, just publish the message
		return b.ps.Publish(topic, msg)
	})
}

// CommitTransaction commits a transaction.
func (b *InProcBroker) CommitTransaction(ctx context.Context, txID string) error {
	return b.executeWithObservability(ctx, OpPublish, "", func() error {
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

		// Check if transaction support is enabled
		if !b.config.EnableTransactions {
			return fmt.Errorf("%w: transaction support is disabled", ErrTransactionNotSupported)
		}

		// TODO: Implement transaction commit logic
		// This would:
		// 1. Validate transaction exists and is active
		// 2. Flush buffered messages
		// 3. Clean up transaction state

		return nil
	})
}

// RollbackTransaction rolls back a transaction.
func (b *InProcBroker) RollbackTransaction(ctx context.Context, txID string) error {
	return b.executeWithObservability(ctx, OpPublish, "", func() error {
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

		// Check if transaction support is enabled
		if !b.config.EnableTransactions {
			return fmt.Errorf("%w: transaction support is disabled", ErrTransactionNotSupported)
		}

		// TODO: Implement transaction rollback logic
		// This would:
		// 1. Validate transaction exists and is active
		// 2. Discard buffered messages
		// 3. Clean up transaction state

		return nil
	})
}

// PublishToDeadLetter publishes a message to the dead letter queue.
func (b *InProcBroker) PublishToDeadLetter(ctx context.Context, originalTopic string, msg Message, reason string) error {
	return b.executeWithObservability(ctx, OpPublish, originalTopic, func() error {
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

		// Check if dead letter queue is enabled
		if !b.config.EnableDeadLetterQueue {
			return fmt.Errorf("%w: dead letter queue is disabled", ErrDeadLetterNotSupported)
		}

		// Validate message
		if err := validateMessage(msg); err != nil {
			return err
		}

		// Determine dead letter topic
		topic := b.config.DeadLetterTopic
		if topic == "" {
			topic = "dead-letter"
		}

		// TODO: Implement dead letter queue logic
		// This would:
		// 1. Add metadata about original topic and reason
		// 2. Publish to dead letter topic
		// 3. Track dead letter metrics

		return b.ps.Publish(topic, msg)
	})
}

// GetDeadLetterStats returns statistics about the dead letter queue.
func (b *InProcBroker) GetDeadLetterStats() DeadLetterStats {
	if b == nil || !b.config.EnableDeadLetterQueue {
		return DeadLetterStats{
			Enabled: false,
		}
	}

	// TODO: Implement dead letter stats collection
	// This would track:
	// - Total dead letter messages
	// - Dead letter rate
	// - Recent dead letter reasons

	return DeadLetterStats{
		Enabled: true,
		Topic:   b.config.DeadLetterTopic,
	}
}

// StartMQTTServer starts the MQTT protocol server.
func (b *InProcBroker) StartMQTTServer() error {
	if b == nil {
		return fmt.Errorf("%w", ErrNotInitialized)
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
		return fmt.Errorf("%w", ErrNotInitialized)
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

// Close shuts down the broker.
func (b *InProcBroker) Close() error {
	return b.executeWithObservability(context.Background(), OpClose, "", func() error {
		// Validate broker initialization
		if b == nil || b.ps == nil {
			return nil // Close is idempotent
		}
		b.closePriorityDispatchers()
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

	// Get memory usage
	status.MemoryUsage = b.GetMemoryUsage()
	status.MemoryLimit = b.config.MaxMemoryUsage

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
