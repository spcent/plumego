package pubsub

import (
	"context"
	"errors"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"
)

// Ordering errors
var (
	ErrOrderingClosed     = errors.New("ordering system is closed")
	ErrInvalidOrderLevel  = errors.New("invalid order level")
	ErrSequenceGap        = errors.New("sequence number gap detected")
	ErrOutOfOrderMessage  = errors.New("out of order message")
)

// OrderLevel defines the level of ordering guarantee
type OrderLevel int

const (
	// OrderNone - no ordering guarantee (default, highest throughput)
	OrderNone OrderLevel = iota

	// OrderPerTopic - messages within the same topic are ordered
	OrderPerTopic

	// OrderPerKey - messages with the same partition key are ordered
	OrderPerKey

	// OrderGlobal - all messages are globally ordered (lowest throughput)
	OrderGlobal
)

// OrderingConfig configures the ordering system
type OrderingConfig struct {
	// DefaultLevel is the default ordering level
	DefaultLevel OrderLevel

	// QueueSize for ordered message queues (default: 1000)
	QueueSize int

	// WorkerCount for processing ordered messages (default: 4)
	WorkerCount int

	// MaxBatchSize for batching ordered messages (default: 10)
	MaxBatchSize int

	// BatchTimeout for flushing batched messages (default: 10ms)
	BatchTimeout time.Duration

	// SequenceCheckEnabled enables sequence number verification
	SequenceCheckEnabled bool
}

// DefaultOrderingConfig returns default ordering configuration
func DefaultOrderingConfig() OrderingConfig {
	return OrderingConfig{
		DefaultLevel:         OrderNone,
		QueueSize:            1000,
		WorkerCount:          4,
		MaxBatchSize:         10,
		BatchTimeout:         10 * time.Millisecond,
		SequenceCheckEnabled: false,
	}
}

// OrderedPubSub wraps InProcPubSub with ordering guarantees
type OrderedPubSub struct {
	*InProcPubSub

	config OrderingConfig

	// Per-topic queues
	topicQueues   map[string]*orderedQueue
	topicQueuesMu sync.RWMutex

	// Per-key queues (for OrderPerKey)
	keyQueues   map[string]*orderedQueue
	keyQueuesMu sync.RWMutex

	// Global queue (for OrderGlobal)
	globalQueue *orderedQueue

	// Sequence tracking
	sequences   map[string]*sequenceTracker
	sequencesMu sync.RWMutex

	// Background workers
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed atomic.Bool

	// Metrics
	orderedPublishes atomic.Uint64
	queuedMessages   atomic.Uint64
	sequenceErrors   atomic.Uint64
}

// orderedQueue represents a FIFO queue for ordered messages
type orderedQueue struct {
	ch        chan orderedMessage
	mu        sync.Mutex
	sequence  atomic.Uint64
	batch     []orderedMessage
	batchSize int
	timer     *time.Timer
	closed    atomic.Bool
}

// orderedMessage wraps a message with ordering metadata
type orderedMessage struct {
	topic    string
	key      string
	message  Message
	sequence uint64
	enqueued time.Time
}

// sequenceTracker tracks message sequences for verification
type sequenceTracker struct {
	mu           sync.Mutex
	lastSequence uint64
	missing      map[uint64]bool
}

// NewOrdered creates a new ordered pubsub instance
func NewOrdered(config OrderingConfig, opts ...Option) *OrderedPubSub {
	// Apply defaults
	if config.QueueSize == 0 {
		config.QueueSize = 1000
	}
	if config.WorkerCount == 0 {
		config.WorkerCount = 4
	}
	if config.MaxBatchSize == 0 {
		config.MaxBatchSize = 10
	}
	if config.BatchTimeout == 0 {
		config.BatchTimeout = 10 * time.Millisecond
	}

	// Create base pubsub
	ps := New(opts...)

	ctx, cancel := context.WithCancel(context.Background())

	ops := &OrderedPubSub{
		InProcPubSub: ps,
		config:       config,
		topicQueues:  make(map[string]*orderedQueue),
		keyQueues:    make(map[string]*orderedQueue),
		sequences:    make(map[string]*sequenceTracker),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Create global queue if needed
	if config.DefaultLevel == OrderGlobal {
		ops.globalQueue = ops.createQueue(config.MaxBatchSize)
		ops.startQueueWorker(ops.globalQueue, "")
	}

	return ops
}

// createQueue creates a new ordered queue
func (ops *OrderedPubSub) createQueue(batchSize int) *orderedQueue {
	return &orderedQueue{
		ch:        make(chan orderedMessage, ops.config.QueueSize),
		batch:     make([]orderedMessage, 0, batchSize),
		batchSize: batchSize,
	}
}

// PublishOrdered publishes a message with ordering guarantee
func (ops *OrderedPubSub) PublishOrdered(topic string, msg Message, level OrderLevel) error {
	if ops.closed.Load() {
		return ErrOrderingClosed
	}

	return ops.PublishWithKey(topic, "", msg, level)
}

// PublishWithKey publishes a message with a partition key
func (ops *OrderedPubSub) PublishWithKey(topic, key string, msg Message, level OrderLevel) error {
	if ops.closed.Load() {
		return ErrOrderingClosed
	}

	// Generate sequence number
	var queue *orderedQueue

	switch level {
	case OrderNone:
		// Fast path: no ordering
		return ops.InProcPubSub.Publish(topic, msg)

	case OrderPerTopic:
		queue = ops.getOrCreateTopicQueue(topic)

	case OrderPerKey:
		if key == "" {
			key = topic // Default to topic if no key
		}
		queue = ops.getOrCreateKeyQueue(key)

	case OrderGlobal:
		if ops.globalQueue == nil {
			ops.globalQueue = ops.createQueue(ops.config.MaxBatchSize)
			ops.startQueueWorker(ops.globalQueue, "")
		}
		queue = ops.globalQueue

	default:
		return ErrInvalidOrderLevel
	}

	// Create ordered message
	om := orderedMessage{
		topic:    topic,
		key:      key,
		message:  msg,
		sequence: queue.sequence.Add(1),
		enqueued: time.Now(),
	}

	// Enqueue
	select {
	case queue.ch <- om:
		ops.queuedMessages.Add(1)
		return nil

	case <-ops.ctx.Done():
		return ErrOrderingClosed

	default:
		// Queue full - either block or drop
		return errors.New("ordered queue full")
	}
}

// getOrCreateTopicQueue gets or creates a queue for a topic
func (ops *OrderedPubSub) getOrCreateTopicQueue(topic string) *orderedQueue {
	ops.topicQueuesMu.RLock()
	queue, exists := ops.topicQueues[topic]
	ops.topicQueuesMu.RUnlock()

	if exists {
		return queue
	}

	ops.topicQueuesMu.Lock()
	defer ops.topicQueuesMu.Unlock()

	// Double-check
	if queue, exists = ops.topicQueues[topic]; exists {
		return queue
	}

	// Create new queue
	queue = ops.createQueue(ops.config.MaxBatchSize)
	ops.topicQueues[topic] = queue

	// Start worker
	ops.startQueueWorker(queue, topic)

	return queue
}

// getOrCreateKeyQueue gets or creates a queue for a key
func (ops *OrderedPubSub) getOrCreateKeyQueue(key string) *orderedQueue {
	ops.keyQueuesMu.RLock()
	queue, exists := ops.keyQueues[key]
	ops.keyQueuesMu.RUnlock()

	if exists {
		return queue
	}

	ops.keyQueuesMu.Lock()
	defer ops.keyQueuesMu.Unlock()

	// Double-check
	if queue, exists = ops.keyQueues[key]; exists {
		return queue
	}

	// Create new queue
	queue = ops.createQueue(ops.config.MaxBatchSize)
	ops.keyQueues[key] = queue

	// Start worker
	ops.startQueueWorker(queue, key)

	return queue
}

// startQueueWorker starts a worker to process an ordered queue
func (ops *OrderedPubSub) startQueueWorker(queue *orderedQueue, identifier string) {
	ops.wg.Add(1)
	go func() {
		defer ops.wg.Done()
		ops.processQueue(queue, identifier)
	}()
}

// processQueue processes messages from an ordered queue
func (ops *OrderedPubSub) processQueue(queue *orderedQueue, identifier string) {
	ticker := time.NewTicker(ops.config.BatchTimeout)
	defer ticker.Stop()

	for {
		select {
		case om, ok := <-queue.ch:
			if !ok {
				// Queue closed, flush remaining
				ops.flushBatch(queue)
				return
			}

			// Add to batch
			queue.mu.Lock()
			queue.batch = append(queue.batch, om)
			shouldFlush := len(queue.batch) >= queue.batchSize
			queue.mu.Unlock()

			if shouldFlush {
				ops.flushBatch(queue)
			}

		case <-ticker.C:
			// Periodic flush
			ops.flushBatch(queue)

		case <-ops.ctx.Done():
			ops.flushBatch(queue)
			return
		}
	}
}

// flushBatch flushes a batch of messages
func (ops *OrderedPubSub) flushBatch(queue *orderedQueue) {
	queue.mu.Lock()
	if len(queue.batch) == 0 {
		queue.mu.Unlock()
		return
	}

	batch := queue.batch
	queue.batch = make([]orderedMessage, 0, queue.batchSize)
	queue.mu.Unlock()

	// Publish in order
	for _, om := range batch {
		// Verify sequence if enabled
		if ops.config.SequenceCheckEnabled {
			if err := ops.verifySequence(om); err != nil {
				ops.sequenceErrors.Add(1)
				// Log but continue
				continue
			}
		}

		// Publish
		_ = ops.InProcPubSub.Publish(om.topic, om.message)
		ops.orderedPublishes.Add(1)
	}
}

// verifySequence verifies message sequence
func (ops *OrderedPubSub) verifySequence(om orderedMessage) error {
	key := om.topic
	if om.key != "" {
		key = om.key
	}

	ops.sequencesMu.Lock()
	tracker, exists := ops.sequences[key]
	if !exists {
		tracker = &sequenceTracker{
			missing: make(map[uint64]bool),
		}
		ops.sequences[key] = tracker
	}
	ops.sequencesMu.Unlock()

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	expected := tracker.lastSequence + 1
	if om.sequence < expected {
		return ErrOutOfOrderMessage
	}

	if om.sequence > expected {
		// Gap detected - mark as missing
		for seq := expected; seq < om.sequence; seq++ {
			tracker.missing[seq] = true
		}
	}

	// Check if this fills a gap
	delete(tracker.missing, om.sequence)
	tracker.lastSequence = om.sequence

	return nil
}

// PartitionKey generates a partition key from a value
func PartitionKey(value string, partitions int) string {
	h := fnv.New32a()
	h.Write([]byte(value))
	partition := int(h.Sum32()) % partitions
	return string(rune('0' + partition))
}

// Close closes the ordered pubsub
func (ops *OrderedPubSub) Close() error {
	if ops.closed.Swap(true) {
		return nil
	}

	// Stop workers
	ops.cancel()

	// Close all queues
	ops.topicQueuesMu.Lock()
	for _, queue := range ops.topicQueues {
		queue.closed.Store(true)
		close(queue.ch)
	}
	ops.topicQueuesMu.Unlock()

	ops.keyQueuesMu.Lock()
	for _, queue := range ops.keyQueues {
		queue.closed.Store(true)
		close(queue.ch)
	}
	ops.keyQueuesMu.Unlock()

	if ops.globalQueue != nil {
		ops.globalQueue.closed.Store(true)
		close(ops.globalQueue.ch)
	}

	// Wait for workers
	ops.wg.Wait()

	// Close base pubsub
	return ops.InProcPubSub.Close()
}

// OrderingStats returns ordering statistics
func (ops *OrderedPubSub) OrderingStats() OrderingStats {
	ops.topicQueuesMu.RLock()
	topicQueues := len(ops.topicQueues)
	ops.topicQueuesMu.RUnlock()

	ops.keyQueuesMu.RLock()
	keyQueues := len(ops.keyQueues)
	ops.keyQueuesMu.RUnlock()

	return OrderingStats{
		OrderedPublishes: ops.orderedPublishes.Load(),
		QueuedMessages:   ops.queuedMessages.Load(),
		SequenceErrors:   ops.sequenceErrors.Load(),
		TopicQueues:      topicQueues,
		KeyQueues:        keyQueues,
	}
}

// OrderingStats holds ordering metrics
type OrderingStats struct {
	OrderedPublishes uint64
	QueuedMessages   uint64
	SequenceErrors   uint64
	TopicQueues      int
	KeyQueues        int
}

// OrderedSubscription wraps a subscription with sequence verification
type OrderedSubscription struct {
	Subscription
	tracker *sequenceTracker
	topic   string
}

// SubscribeOrdered creates a subscription with sequence verification
func (ops *OrderedPubSub) SubscribeOrdered(topic string, opts SubOptions) (*OrderedSubscription, error) {
	sub, err := ops.Subscribe(topic, opts)
	if err != nil {
		return nil, err
	}

	ops.sequencesMu.Lock()
	tracker, exists := ops.sequences[topic]
	if !exists {
		tracker = &sequenceTracker{
			missing: make(map[uint64]bool),
		}
		ops.sequences[topic] = tracker
	}
	ops.sequencesMu.Unlock()

	return &OrderedSubscription{
		Subscription: sub,
		tracker:      tracker,
		topic:        topic,
	}, nil
}

// Receive receives a message and verifies sequence
func (os *OrderedSubscription) Receive(ctx context.Context) (Message, error) {
	select {
	case msg, ok := <-os.C():
		if !ok {
			return Message{}, ErrClosed
		}

		// TODO: Verify sequence from message metadata
		return msg, nil

	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}

// MissingSequences returns missing sequence numbers
func (os *OrderedSubscription) MissingSequences() []uint64 {
	os.tracker.mu.Lock()
	defer os.tracker.mu.Unlock()

	result := make([]uint64, 0, len(os.tracker.missing))
	for seq := range os.tracker.missing {
		result = append(result, seq)
	}

	return result
}
