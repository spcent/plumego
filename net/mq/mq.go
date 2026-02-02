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

// ErrMemoryLimitExceeded is returned when memory usage exceeds the configured limit.
var ErrMemoryLimitExceeded = errors.New("mq: memory limit exceeded")

// ErrMessageExpired is returned when a message has expired based on its TTL.
var ErrMessageExpired = errors.New("mq: message has expired")

// ErrTransactionNotFound is returned when a transaction ID is not found.
var ErrTransactionNotFound = errors.New("mq: transaction not found")

// ErrTransactionTimeout is returned when a transaction exceeds its timeout.
var ErrTransactionTimeout = errors.New("mq: transaction timeout")

// ErrTransactionCommitted is returned when attempting to use a committed transaction.
var ErrTransactionCommitted = errors.New("mq: transaction already committed")

// ErrTransactionRolledBack is returned when attempting to use a rolled back transaction.
var ErrTransactionRolledBack = errors.New("mq: transaction already rolled back")

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

// PriorityTTLMessage combines priority and TTL features.
type PriorityTTLMessage struct {
	Message
	Priority  MessagePriority
	ExpiresAt time.Time
}

// AckTTLMessage combines acknowledgment and TTL features.
type AckTTLMessage struct {
	Message
	AckID      string
	AckPolicy  AckPolicy
	AckTimeout time.Duration
	ExpiresAt  time.Time
}

// ackEntry represents a pending acknowledgment for a message.
type ackEntry struct {
	messageID   string
	topic       string
	message     Message
	timestamp   time.Time
	timeout     time.Duration
	retryCount  int
	maxRetries  int
	timer       *time.Timer
	ackReceived bool
}

// ackTracker manages pending acknowledgments and timeouts.
type ackTracker struct {
	mu      sync.RWMutex
	pending map[string]*ackEntry
	broker  *InProcBroker
	closed  bool
}

// newAckTracker creates a new acknowledgment tracker.
func newAckTracker(broker *InProcBroker) *ackTracker {
	return &ackTracker{
		pending: make(map[string]*ackEntry),
		broker:  broker,
	}
}

// track starts tracking an acknowledgment for a message.
func (at *ackTracker) track(ackID, topic string, msg Message, timeout time.Duration, maxRetries int) error {
	at.mu.Lock()
	defer at.mu.Unlock()

	if at.closed {
		return ErrBrokerClosed
	}

	if _, exists := at.pending[ackID]; exists {
		return fmt.Errorf("%w: ack ID %s already exists", ErrMessageAcknowledged, ackID)
	}

	entry := &ackEntry{
		messageID:  msg.ID,
		topic:      topic,
		message:    msg,
		timestamp:  time.Now(),
		timeout:    timeout,
		retryCount: 0,
		maxRetries: maxRetries,
	}

	// Set up timeout timer
	entry.timer = time.AfterFunc(timeout, func() {
		at.handleTimeout(ackID)
	})

	at.pending[ackID] = entry
	return nil
}

// acknowledge marks a message as acknowledged.
func (at *ackTracker) acknowledge(ackID string) error {
	at.mu.Lock()
	defer at.mu.Unlock()

	if at.closed {
		return ErrBrokerClosed
	}

	entry, exists := at.pending[ackID]
	if !exists {
		return fmt.Errorf("%w: ack ID %s not found", ErrMessageNotAcked, ackID)
	}

	if entry.ackReceived {
		return fmt.Errorf("%w: ack ID %s", ErrMessageAcknowledged, ackID)
	}

	// Mark as acknowledged and stop timer
	entry.ackReceived = true
	if entry.timer != nil {
		entry.timer.Stop()
	}

	// Remove from pending
	delete(at.pending, ackID)
	return nil
}

// handleTimeout processes a timeout for an acknowledgment.
func (at *ackTracker) handleTimeout(ackID string) {
	at.mu.Lock()
	entry, exists := at.pending[ackID]
	if !exists || entry.ackReceived {
		at.mu.Unlock()
		return
	}

	entry.retryCount++
	shouldRetry := entry.retryCount <= entry.maxRetries
	at.mu.Unlock()

	if shouldRetry {
		// Retry: re-publish the message
		if err := at.broker.redeliverMessage(entry); err != nil {
			// If retry fails, send to dead letter queue
			if at.broker.config.EnableDeadLetterQueue {
				_ = at.broker.PublishToDeadLetter(
					context.Background(),
					entry.topic,
					entry.message,
					fmt.Sprintf("ack timeout after %d retries: %v", entry.retryCount, err),
				)
			}
		}
	} else {
		// Max retries exceeded, send to dead letter queue
		if at.broker.config.EnableDeadLetterQueue {
			_ = at.broker.PublishToDeadLetter(
				context.Background(),
				entry.topic,
				entry.message,
				fmt.Sprintf("ack timeout after %d retries", entry.retryCount),
			)
		}

		// Remove from pending
		at.mu.Lock()
		delete(at.pending, ackID)
		at.mu.Unlock()
	}
}

// close shuts down the tracker and cancels all pending timeouts.
func (at *ackTracker) close() {
	at.mu.Lock()
	defer at.mu.Unlock()

	if at.closed {
		return
	}

	at.closed = true

	// Stop all timers
	for _, entry := range at.pending {
		if entry.timer != nil {
			entry.timer.Stop()
		}
	}

	// Clear pending
	at.pending = nil
}

// stats returns statistics about pending acknowledgments.
func (at *ackTracker) stats() (pending int, oldestAge time.Duration) {
	at.mu.RLock()
	defer at.mu.RUnlock()

	pending = len(at.pending)
	if pending == 0 {
		return
	}

	now := time.Now()
	oldest := now
	for _, entry := range at.pending {
		if entry.timestamp.Before(oldest) {
			oldest = entry.timestamp
		}
	}
	oldestAge = now.Sub(oldest)

	return
}

// ttlEntry tracks a message with TTL.
type ttlEntry struct {
	messageID string
	topic     string
	expiresAt time.Time
}

// ttlTracker manages messages with TTL and handles expiration cleanup.
type ttlTracker struct {
	mu       sync.RWMutex
	messages map[string]*ttlEntry // messageID -> entry
	broker   *InProcBroker
	stopCh   chan struct{}
	closed   bool
}

// newTTLTracker creates a new TTL tracker and starts the cleanup goroutine.
func newTTLTracker(broker *InProcBroker) *ttlTracker {
	tt := &ttlTracker{
		messages: make(map[string]*ttlEntry),
		broker:   broker,
		stopCh:   make(chan struct{}),
	}

	// Start background cleanup goroutine
	go tt.cleanupLoop()

	return tt
}

// track registers a message with TTL for tracking.
func (tt *ttlTracker) track(messageID, topic string, expiresAt time.Time) error {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	if tt.closed {
		return ErrBrokerClosed
	}

	// Don't track if already expired
	if time.Now().After(expiresAt) {
		return ErrMessageExpired
	}

	tt.messages[messageID] = &ttlEntry{
		messageID: messageID,
		topic:     topic,
		expiresAt: expiresAt,
	}

	return nil
}

// isExpired checks if a message has expired.
func (tt *ttlTracker) isExpired(messageID string) bool {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	entry, exists := tt.messages[messageID]
	if !exists {
		return false
	}

	return time.Now().After(entry.expiresAt)
}

// remove removes a message from tracking (e.g., when consumed).
func (tt *ttlTracker) remove(messageID string) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	delete(tt.messages, messageID)
}

// cleanupLoop runs in the background and removes expired messages.
func (tt *ttlTracker) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tt.cleanupExpiredMessages()
		case <-tt.stopCh:
			return
		}
	}
}

// cleanupExpiredMessages removes all expired messages from tracking.
func (tt *ttlTracker) cleanupExpiredMessages() {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	now := time.Now()
	expiredIDs := make([]string, 0)

	for id, entry := range tt.messages {
		if now.After(entry.expiresAt) {
			expiredIDs = append(expiredIDs, id)
		}
	}

	// Remove expired messages
	for _, id := range expiredIDs {
		delete(tt.messages, id)
	}

	// Note: The actual message removal from pubsub would require
	// additional support in the pubsub package to remove specific messages
	// from subscriber buffers. For now, we just track expiration.
}

// close stops the cleanup goroutine and clears all tracked messages.
func (tt *ttlTracker) close() {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	if tt.closed {
		return
	}

	tt.closed = true
	close(tt.stopCh)

	// Clear all tracked messages
	tt.messages = make(map[string]*ttlEntry)
}

// stats returns the number of tracked messages and count of expired messages.
func (tt *ttlTracker) stats() (tracked int, expired int) {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	tracked = len(tt.messages)
	now := time.Now()

	for _, entry := range tt.messages {
		if now.After(entry.expiresAt) {
			expired++
		}
	}

	return
}

// TransactionState represents the state of a transaction.
type TransactionState int

const (
	// TxStateActive - Transaction is active and accepting messages
	TxStateActive TransactionState = iota
	// TxStateCommitted - Transaction has been committed
	TxStateCommitted
	// TxStateRolledBack - Transaction has been rolled back
	TxStateRolledBack
	// TxStateTimeout - Transaction has timed out
	TxStateTimeout
)

// transaction represents a message queue transaction.
type transaction struct {
	id        string
	state     TransactionState
	messages  []messageEnvelope // Buffered messages
	createdAt time.Time
	timeout   time.Duration
	timer     *time.Timer
	mu        sync.RWMutex
}

// messageEnvelope wraps a message with its topic for transaction buffering.
type messageEnvelope struct {
	topic string
	msg   Message
}

// transactionManager manages all active transactions.
type transactionManager struct {
	mu           sync.RWMutex
	transactions map[string]*transaction
	broker       *InProcBroker
	closed       bool
}

// newTransactionManager creates a new transaction manager.
func newTransactionManager(broker *InProcBroker) *transactionManager {
	return &transactionManager{
		transactions: make(map[string]*transaction),
		broker:       broker,
	}
}

// begin starts a new transaction with the given ID and timeout.
func (tm *transactionManager) begin(txID string, timeout time.Duration) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.closed {
		return ErrBrokerClosed
	}

	if _, exists := tm.transactions[txID]; exists {
		return fmt.Errorf("transaction %s already exists", txID)
	}

	tx := &transaction{
		id:        txID,
		state:     TxStateActive,
		messages:  make([]messageEnvelope, 0),
		createdAt: time.Now(),
		timeout:   timeout,
	}

	// Set up timeout timer
	if timeout > 0 {
		tx.timer = time.AfterFunc(timeout, func() {
			tm.handleTimeout(txID)
		})
	}

	tm.transactions[txID] = tx
	return nil
}

// addMessage adds a message to the transaction buffer.
func (tm *transactionManager) addMessage(txID string, topic string, msg Message) error {
	tm.mu.RLock()
	tx, exists := tm.transactions[txID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: %s", ErrTransactionNotFound, txID)
	}

	tx.mu.Lock()
	defer tx.mu.Unlock()

	switch tx.state {
	case TxStateCommitted:
		return ErrTransactionCommitted
	case TxStateRolledBack:
		return ErrTransactionRolledBack
	case TxStateTimeout:
		return ErrTransactionTimeout
	}

	tx.messages = append(tx.messages, messageEnvelope{
		topic: topic,
		msg:   msg,
	})

	return nil
}

// commit commits the transaction, publishing all buffered messages.
func (tm *transactionManager) commit(txID string) error {
	tm.mu.Lock()
	tx, exists := tm.transactions[txID]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrTransactionNotFound, txID)
	}
	tm.mu.Unlock()

	tx.mu.Lock()
	defer tx.mu.Unlock()

	switch tx.state {
	case TxStateCommitted:
		return ErrTransactionCommitted
	case TxStateRolledBack:
		return ErrTransactionRolledBack
	case TxStateTimeout:
		return ErrTransactionTimeout
	}

	// Stop the timeout timer
	if tx.timer != nil {
		tx.timer.Stop()
	}

	// Publish all buffered messages
	for _, env := range tx.messages {
		if err := tm.broker.ps.Publish(env.topic, env.msg); err != nil {
			// If publish fails, mark as rolled back
			tx.state = TxStateRolledBack
			return fmt.Errorf("transaction commit failed: %w", err)
		}
	}

	// Mark as committed
	tx.state = TxStateCommitted

	// Remove from active transactions
	tm.mu.Lock()
	delete(tm.transactions, txID)
	tm.mu.Unlock()

	return nil
}

// rollback rolls back the transaction, discarding all buffered messages.
func (tm *transactionManager) rollback(txID string) error {
	tm.mu.Lock()
	tx, exists := tm.transactions[txID]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrTransactionNotFound, txID)
	}
	tm.mu.Unlock()

	tx.mu.Lock()
	defer tx.mu.Unlock()

	switch tx.state {
	case TxStateCommitted:
		return ErrTransactionCommitted
	case TxStateRolledBack:
		return ErrTransactionRolledBack
	case TxStateTimeout:
		return ErrTransactionTimeout
	}

	// Stop the timeout timer
	if tx.timer != nil {
		tx.timer.Stop()
	}

	// Mark as rolled back and discard messages
	tx.state = TxStateRolledBack
	tx.messages = nil

	// Remove from active transactions
	tm.mu.Lock()
	delete(tm.transactions, txID)
	tm.mu.Unlock()

	return nil
}

// handleTimeout processes a transaction timeout.
func (tm *transactionManager) handleTimeout(txID string) {
	tm.mu.Lock()
	tx, exists := tm.transactions[txID]
	if !exists {
		tm.mu.Unlock()
		return
	}
	tm.mu.Unlock()

	tx.mu.Lock()
	if tx.state != TxStateActive {
		tx.mu.Unlock()
		return
	}

	// Mark as timed out and discard messages
	tx.state = TxStateTimeout
	tx.messages = nil
	tx.mu.Unlock()

	// Remove from active transactions
	tm.mu.Lock()
	delete(tm.transactions, txID)
	tm.mu.Unlock()
}

// close shuts down the transaction manager and aborts all active transactions.
func (tm *transactionManager) close() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.closed {
		return
	}

	tm.closed = true

	// Abort all active transactions
	for _, tx := range tm.transactions {
		tx.mu.Lock()
		if tx.timer != nil {
			tx.timer.Stop()
		}
		tx.state = TxStateRolledBack
		tx.messages = nil
		tx.mu.Unlock()
	}

	// Clear all transactions
	tm.transactions = nil
}

// stats returns statistics about active transactions.
func (tm *transactionManager) stats() (active int, totalMessages int) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	active = len(tm.transactions)
	for _, tx := range tm.transactions {
		tx.mu.RLock()
		totalMessages += len(tx.messages)
		tx.mu.RUnlock()
	}

	return
}

// deadLetterEntry represents a message in the dead letter queue.
type deadLetterEntry struct {
	originalTopic string
	message       Message
	reason        string
	timestamp     time.Time
}

// deadLetterManager manages the dead letter queue.
type deadLetterManager struct {
	mu              sync.RWMutex
	entries         []deadLetterEntry
	totalCount      uint64
	lastMessageTime time.Time
	broker          *InProcBroker
	closed          bool
}

// newDeadLetterManager creates a new dead letter manager.
func newDeadLetterManager(broker *InProcBroker) *deadLetterManager {
	return &deadLetterManager{
		entries: make([]deadLetterEntry, 0),
		broker:  broker,
	}
}

// add adds a message to the dead letter queue.
func (dlm *deadLetterManager) add(originalTopic string, msg Message, reason string) error {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()

	if dlm.closed {
		return ErrBrokerClosed
	}

	entry := deadLetterEntry{
		originalTopic: originalTopic,
		message:       msg,
		reason:        reason,
		timestamp:     time.Now(),
	}

	dlm.entries = append(dlm.entries, entry)
	dlm.totalCount++
	dlm.lastMessageTime = time.Now()

	return nil
}

// stats returns statistics about the dead letter queue.
func (dlm *deadLetterManager) stats() DeadLetterStats {
	dlm.mu.RLock()
	defer dlm.mu.RUnlock()

	return DeadLetterStats{
		Enabled:         true,
		Topic:           dlm.broker.config.DeadLetterTopic,
		TotalMessages:   dlm.totalCount,
		CurrentCount:    len(dlm.entries),
		LastMessageTime: dlm.lastMessageTime,
	}
}

// close shuts down the dead letter manager.
func (dlm *deadLetterManager) close() {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()

	if dlm.closed {
		return
	}

	dlm.closed = true
	dlm.entries = nil
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
		DefaultBufferSize:        DefaultBufferSize,
		EnableMetrics:            true,
		HealthCheckInterval:      DefaultHealthCheckInterval,
		MessageTTL:               0, // No TTL by default
		EnablePriorityQueue:      true,
		EnableAckSupport:         false, // Disabled by default for backward compatibility
		DefaultAckTimeout:        DefaultAckTimeoutDuration,
		MaxMemoryUsage:           0,     // No limit by default
		EnableTriePattern:        false, // Disabled by default for backward compatibility
		EnableCluster:            false, // Disabled by default
		ClusterReplicationFactor: 1,     // No replication by default
		ClusterSyncInterval:      DefaultClusterSyncInterval,
		EnablePersistence:        false, // Disabled by default
		EnableDeadLetterQueue:    false, // Disabled by default
		EnableTransactions:       false, // Disabled by default
		TransactionTimeout:       DefaultTransactionTimeoutDuration,
		EnableMQTT:               false, // Disabled by default
		MQTTPort:                 DefaultMQTTPort,
		EnableAMQP:               false, // Disabled by default
		AMQPPort:                 DefaultAMQPPort,
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	// Basic validation
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

	// Cluster configuration validation
	if c.EnableCluster {
		if strings.TrimSpace(c.ClusterNodeID) == "" {
			return fmt.Errorf("%w: ClusterNodeID is required when cluster mode is enabled", ErrInvalidConfig)
		}
		if c.ClusterReplicationFactor < 1 {
			return fmt.Errorf("%w: ClusterReplicationFactor must be at least 1", ErrInvalidConfig)
		}
		if c.ClusterSyncInterval < 0 {
			return fmt.Errorf("%w: ClusterSyncInterval cannot be negative", ErrInvalidConfig)
		}
	}

	// Persistence configuration validation
	if c.EnablePersistence {
		if strings.TrimSpace(c.PersistencePath) == "" {
			return fmt.Errorf("%w: PersistencePath is required when persistence is enabled", ErrInvalidConfig)
		}
	}

	// Dead letter queue configuration validation
	if c.EnableDeadLetterQueue {
		if strings.TrimSpace(c.DeadLetterTopic) == "" {
			return fmt.Errorf("%w: DeadLetterTopic is required when dead letter queue is enabled", ErrInvalidConfig)
		}
	}

	// Acknowledgment configuration validation
	if c.EnableAckSupport {
		if c.DefaultAckTimeout < 0 {
			return fmt.Errorf("%w: DefaultAckTimeout cannot be negative", ErrInvalidConfig)
		}
	}

	// Transaction configuration validation
	if c.EnableTransactions {
		if c.TransactionTimeout < 0 {
			return fmt.Errorf("%w: TransactionTimeout cannot be negative", ErrInvalidConfig)
		}
	}

	// Protocol configuration validation
	if c.EnableMQTT {
		if c.MQTTPort <= 0 || c.MQTTPort > 65535 {
			return fmt.Errorf("%w: MQTTPort must be between 1 and 65535", ErrInvalidConfig)
		}
	}
	if c.EnableAMQP {
		if c.AMQPPort <= 0 || c.AMQPPort > 65535 {
			return fmt.Errorf("%w: AMQPPort must be between 1 and 65535", ErrInvalidConfig)
		}
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
		return fmt.Errorf("%w", ErrNotInitialized)
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
		return fmt.Errorf("%w", ErrNotInitialized)
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
		// Validate operation
		if err := b.validatePublishOperation(ctx, topic, &msg); err != nil {
			return err
		}

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
			return err
		}

		// Check TTL if message has expiration
		// Note: TTLMessage is a wrapper, so we need to check the underlying type
		// For now, we'll skip TTL check for regular Message types

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

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
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

// PublishTTL publishes a message with time-to-live expiration.
func (b *InProcBroker) PublishTTL(ctx context.Context, topic string, msg TTLMessage) error {
	return b.executeWithObservability(ctx, OpPublish, topic, func() error {
		// Validate operation
		if err := b.validatePublishOperation(ctx, topic, &msg.Message); err != nil {
			return err
		}

		// Validate TTL
		if err := b.validateTTL(msg.ExpiresAt); err != nil {
			return err
		}

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
			return err
		}

		// Track message with TTL if tracker is enabled
		if b.ttlTracker != nil && !msg.ExpiresAt.IsZero() {
			if err := b.ttlTracker.track(msg.Message.ID, topic, msg.ExpiresAt); err != nil {
				return err
			}
		}

		return b.ps.Publish(topic, msg.Message)
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

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
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

		// Generate ACK ID if not provided
		if msg.AckID == "" {
			msg.AckID = fmt.Sprintf("%s-%d", msg.Message.ID, time.Now().UnixNano())
		}

		// Track acknowledgment if required
		if msg.AckPolicy == AckRequired || msg.AckPolicy == AckTimeout {
			maxRetries := 3 // Default max retries
			if err := b.ackTracker.track(msg.AckID, topic, msg.Message, msg.AckTimeout, maxRetries); err != nil {
				return err
			}
		}

		// Publish the message
		return b.ps.Publish(topic, msg.Message)
	})
}

// PublishPriorityTTL publishes a message with both priority and TTL.
func (b *InProcBroker) PublishPriorityTTL(ctx context.Context, topic string, msg PriorityTTLMessage) error {
	return b.executeWithObservability(ctx, OpPublish, topic, func() error {
		// Validate operation
		if err := b.validatePublishOperation(ctx, topic, &msg.Message); err != nil {
			return err
		}

		// Validate TTL
		if err := b.validateTTL(msg.ExpiresAt); err != nil {
			return err
		}

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
			return err
		}

		// Track message with TTL if tracker is enabled
		if b.ttlTracker != nil && !msg.ExpiresAt.IsZero() {
			if err := b.ttlTracker.track(msg.Message.ID, topic, msg.ExpiresAt); err != nil {
				return err
			}
		}

		// Create priority envelope and dispatch
		env := &priorityEnvelope{
			msg:      msg.Message,
			priority: msg.Priority,
			seq:      atomic.AddUint64(&b.prioritySeq, 1),
			ctx:      ctx,
			done:     make(chan error, 1),
		}

		dispatcher, err := b.ensurePriorityDispatcher(topic)
		if err != nil {
			return err
		}

		// Enqueue message with priority
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

// PublishWithAckTTL publishes a message with both acknowledgment and TTL.
func (b *InProcBroker) PublishWithAckTTL(ctx context.Context, topic string, msg AckTTLMessage) error {
	return b.executeWithObservability(ctx, OpPublish, topic, func() error {
		// Validate operation
		if err := b.validatePublishOperation(ctx, topic, &msg.Message); err != nil {
			return err
		}

		// Validate TTL
		if err := b.validateTTL(msg.ExpiresAt); err != nil {
			return err
		}

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
			return err
		}

		// Check if acknowledgment support is enabled
		if !b.config.EnableAckSupport {
			return fmt.Errorf("%w: acknowledgment support is disabled", ErrInvalidConfig)
		}

		// Set default timeout if not specified
		ackTimeout := msg.AckTimeout
		if ackTimeout == 0 {
			ackTimeout = b.config.DefaultAckTimeout
		}

		// Generate ACK ID if not provided
		ackID := msg.AckID
		if ackID == "" {
			ackID = fmt.Sprintf("%s-%d", msg.Message.ID, time.Now().UnixNano())
		}

		// Track acknowledgment if required
		if msg.AckPolicy == AckRequired || msg.AckPolicy == AckTimeout {
			maxRetries := 3 // Default retry count
			if err := b.ackTracker.track(ackID, topic, msg.Message, ackTimeout, maxRetries); err != nil {
				return err
			}
		}

		// Track message with TTL if tracker is enabled
		if b.ttlTracker != nil && !msg.ExpiresAt.IsZero() {
			if err := b.ttlTracker.track(msg.Message.ID, topic, msg.ExpiresAt); err != nil {
				return err
			}
		}

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

		// Validate ackTracker
		if b.ackTracker == nil {
			return fmt.Errorf("%w: ack tracker not initialized", ErrNotInitialized)
		}

		// Acknowledge the message
		return b.ackTracker.acknowledge(messageID)
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

		// Validate ackTracker
		if b.ackTracker == nil {
			return fmt.Errorf("%w: ack tracker not initialized", ErrNotInitialized)
		}

		// Get the entry from tracker
		b.ackTracker.mu.RLock()
		entry, exists := b.ackTracker.pending[messageID]
		b.ackTracker.mu.RUnlock()

		if !exists {
			return fmt.Errorf("%w: ack ID %s not found", ErrMessageNotAcked, messageID)
		}

		// Immediately redeliver the message
		if err := b.redeliverMessage(entry); err != nil {
			return fmt.Errorf("failed to redeliver message: %w", err)
		}

		// Stop the timer and remove from pending
		b.ackTracker.mu.Lock()
		if entry.timer != nil {
			entry.timer.Stop()
		}
		delete(b.ackTracker.pending, messageID)
		b.ackTracker.mu.Unlock()

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

		// Validate txManager
		if b.txManager == nil {
			return fmt.Errorf("%w: transaction manager not initialized", ErrNotInitialized)
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

		// Begin transaction if it doesn't exist yet
		// Note: begin will return error if transaction already exists
		timeout := b.config.TransactionTimeout
		if timeout == 0 {
			timeout = DefaultTransactionTimeoutDuration
		}
		_ = b.txManager.begin(txID, timeout) // Ignore "already exists" error

		// Add message to transaction buffer
		return b.txManager.addMessage(txID, topic, msg)
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

		// Validate txManager
		if b.txManager == nil {
			return fmt.Errorf("%w: transaction manager not initialized", ErrNotInitialized)
		}

		// Commit the transaction
		return b.txManager.commit(txID)
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

		// Validate txManager
		if b.txManager == nil {
			return fmt.Errorf("%w: transaction manager not initialized", ErrNotInitialized)
		}

		// Rollback the transaction
		return b.txManager.rollback(txID)
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

		// Validate deadLetterManager
		if b.deadLetterManager == nil {
			return fmt.Errorf("%w: dead letter manager not initialized", ErrNotInitialized)
		}

		// Validate message
		if err := validateMessage(msg); err != nil {
			return err
		}

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
			return err
		}

		// Add to dead letter manager
		if err := b.deadLetterManager.add(originalTopic, msg, reason); err != nil {
			return err
		}

		// Determine dead letter topic
		topic := b.config.DeadLetterTopic
		if topic == "" {
			topic = "dead-letter"
		}

		// Publish to dead letter topic
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

	if b.deadLetterManager == nil {
		return DeadLetterStats{
			Enabled: true,
			Topic:   b.config.DeadLetterTopic,
		}
	}

	return b.deadLetterManager.stats()
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

// RecoverMessages recovers persisted messages for a topic.
// This is useful for replaying messages after broker restart.
func (b *InProcBroker) RecoverMessages(ctx context.Context, topic string, limit int) ([]Message, error) {
	if b == nil || b.ps == nil {
		return nil, fmt.Errorf("%w", ErrNotInitialized)
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
