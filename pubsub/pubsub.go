package pubsub

import (
	"context"
	"path"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spcent/plumego/metrics"
)

// subscriber represents a subscription to a topic.
type subscriber struct {
	id      uint64
	topic   string
	ch      chan Message
	opts    SubOptions
	closed  atomic.Bool
	once    sync.Once
	mu      sync.Mutex
	pattern bool
	done    chan struct{}

	// Ring buffer for O(1) DropOldest (nil when not using ring buffer)
	ringBuf  *ringBuffer
	pumpDone chan struct{}

	// Statistics
	received atomic.Uint64
	dropped  atomic.Uint64

	// back-reference to parent
	ps *InProcPubSub

	// context cancellation
	cancel context.CancelFunc
}

type patternSnapshot struct {
	pattern string
	subs    []*subscriber
}

// C returns the receive-only message channel.
func (s *subscriber) C() <-chan Message { return s.ch }

// ID returns the unique subscription ID.
func (s *subscriber) ID() uint64 { return s.id }

// Topic returns the topic or pattern this subscription is for.
func (s *subscriber) Topic() string { return s.topic }

// Done returns a channel that is closed when the subscription is cancelled.
func (s *subscriber) Done() <-chan struct{} { return s.done }

// Stats returns subscription statistics.
func (s *subscriber) Stats() SubscriptionStats {
	if s.ringBuf != nil {
		return SubscriptionStats{
			Received: s.received.Load(),
			Dropped:  s.dropped.Load(),
			QueueLen: s.ringBuf.Len(),
			QueueCap: s.ringBuf.Cap(),
		}
	}
	return SubscriptionStats{
		Received: s.received.Load(),
		Dropped:  s.dropped.Load(),
		QueueLen: len(s.ch),
		QueueCap: cap(s.ch),
	}
}

// Cancel unsubscribes and closes the channel.
func (s *subscriber) Cancel() {
	s.once.Do(func() {
		s.mu.Lock()
		if s.closed.Swap(true) {
			s.mu.Unlock()
			return
		}
		if s.ringBuf != nil {
			// For ring buffer subscribers, close the ring buffer.
			// The pump goroutine will close s.ch when it exits.
			s.ringBuf.Close()
		} else {
			close(s.ch)
		}
		close(s.done)
		if s.cancel != nil {
			s.cancel()
		}
		s.mu.Unlock()

		// Wait for pump goroutine to finish if using ring buffer
		if s.pumpDone != nil {
			<-s.pumpDone
		}

		// Call hook before removing
		if s.ps.config.Hooks.OnUnsubscribe != nil {
			s.ps.config.Hooks.OnUnsubscribe(s.topic, s.id)
		}

		s.ps.removeSubscriber(s.topic, s.id, s.pattern)
	})
}

// pumpRingBuffer moves messages from the ring buffer to the output channel.
// It runs as a background goroutine for ring buffer–based subscribers.
func (s *subscriber) pumpRingBuffer() {
	defer close(s.ch)
	defer close(s.pumpDone)

	for {
		// Drain all available messages from ring buffer
		for {
			if s.closed.Load() {
				return
			}
			msg, ok := s.ringBuf.Pop()
			if !ok {
				break
			}
			select {
			case s.ch <- msg:
			case <-s.done:
				return
			}
		}

		// Wait for new messages or closure
		select {
		case _, ok := <-s.ringBuf.Notify():
			if !ok {
				return // ring buffer closed
			}
		case <-s.done:
			return
		}
	}
}

// InProcPubSub is an in-memory pubsub implementation.
//
// This is a lightweight, in-process publish/subscribe system designed for:
//   - Event-driven architectures
//   - WebSocket fan-out scenarios
//   - Internal service communication
//   - Webhook event distribution
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	// Create a pubsub instance with options
//	ps := pubsub.New(
//		pubsub.WithShardCount(32),
//		pubsub.WithDefaultBufferSize(64),
//		pubsub.WithWorkerPoolSize(512),
//	)
//	defer ps.Close()
//
//	// Subscribe to a topic
//	sub, err := ps.Subscribe("user.created", pubsub.SubOptions{
//		BufferSize: 16,
//		Policy:     pubsub.DropOldest,
//	})
//	if err != nil {
//		// Handle error
//	}
//	defer sub.Cancel()
//
//	// Publish a message
//	msg := pubsub.Message{
//		Data: map[string]any{"user_id": "123", "email": "user@example.com"},
//	}
//	err = ps.Publish("user.created", msg)
type InProcPubSub struct {
	shards *shardedMap
	config Config

	closed atomic.Bool
	nextID atomic.Uint64

	metrics    metricsPubSub
	workerPool *workerPool

	// draining state
	draining   atomic.Bool
	drainWg    sync.WaitGroup
	drainMu    sync.RWMutex
	drainCond  *sync.Cond
	pendingOps atomic.Int64

	// message scheduler for delayed delivery
	scheduler *messageScheduler

	// TTL manager for message expiration
	ttlMgr *ttlManager

	// topic history for message retention
	history *topicHistory

	// request-reply manager for efficient RPC-style communication
	requestMgr *requestManager
}

// New creates a new InProcPubSub instance.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	// Simple creation
//	ps := pubsub.New()
//	defer ps.Close()
//
//	// With options
//	ps := pubsub.New(
//		pubsub.WithShardCount(32),
//		pubsub.WithDefaultBufferSize(64),
//		pubsub.WithWorkerPoolSize(512),
//		pubsub.WithMetricsCollector(collector),
//	)
func New(opts ...Option) *InProcPubSub {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(&config)
	}

	ps := &InProcPubSub{
		shards:     newShardedMap(config.ShardCount),
		config:     config,
		workerPool: newWorkerPool(config.WorkerPoolSize),
	}
	ps.drainCond = sync.NewCond(&ps.drainMu)
	ps.nextID.Store(0)

	if config.EnableScheduler {
		ps.scheduler = newMessageScheduler(ps)
	}

	if config.EnableTTL {
		ps.ttlMgr = newTTLManager(ps, config.TTLCleanupInterval)
	}

	if config.EnableHistory {
		ps.history = newTopicHistory(config.HistoryConfig)
	}

	if config.EnableRequestReply {
		ps.requestMgr = newRequestManager(ps)
	}

	return ps
}

// Close shuts down the pubsub system.
//
// This method:
//   - Marks the pubsub as closed
//   - Cancels all active subscriptions
//   - Waits for all async operations to complete
//   - Cleans up all resources
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	// ... use pubsub ...
//	ps.Close()
func (ps *InProcPubSub) Close() error {
	if ps.closed.Swap(true) {
		return nil
	}

	// Stop request manager, scheduler and TTL manager before closing subscriptions
	if ps.requestMgr != nil {
		ps.requestMgr.Close()
	}
	if ps.scheduler != nil {
		ps.scheduler.Close()
	}
	if ps.ttlMgr != nil {
		ps.ttlMgr.Close()
	}

	// Wait for pending operations
	ps.workerPool.Close()

	// Collect and cancel all subscribers
	all := ps.shards.collectAllSubscribers()
	for _, s := range all {
		s.Cancel()
	}

	return nil
}

// Drain waits for all pending messages to be delivered or until the context is cancelled.
func (ps *InProcPubSub) Drain(ctx context.Context) error {
	if ps.closed.Load() {
		return ErrClosed
	}

	ps.draining.Store(true)
	defer ps.draining.Store(false)

	// Wait for pending operations
	done := make(chan struct{})
	go func() {
		for ps.pendingOps.Load() > 0 {
			time.Sleep(10 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Subscribe creates a new subscription to a topic.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	sub, err := ps.Subscribe("user.created", pubsub.SubOptions{
//		BufferSize: 16,
//		Policy:     pubsub.DropOldest,
//	})
//	if err != nil {
//		// Handle error
//	}
//	defer sub.Cancel()
//
//	// Receive messages
//	for msg := range sub.C() {
//		// Process message
//		fmt.Printf("Received: %v\n", msg.Data)
//	}
func (ps *InProcPubSub) Subscribe(topic string, opts SubOptions) (Subscription, error) {
	return ps.subscribeInternal(context.Background(), topic, opts, false)
}

// SubscribeWithContext creates a subscription that is cancelled when the context is done.
func (ps *InProcPubSub) SubscribeWithContext(ctx context.Context, topic string, opts SubOptions) (Subscription, error) {
	return ps.subscribeInternal(ctx, topic, opts, false)
}

// SubscribePattern creates a new subscription to a topic pattern.
//
// Pattern matching uses filepath.Match syntax with topic names as literal strings.
// The pattern is matched against the entire topic name character-by-character.
//
// Supported wildcards:
//   - "*" matches zero or more characters (including '.')
//   - "?" matches exactly one character (including '.')
//   - "[...]" matches any single character in the bracket expression
//   - "[^...]" matches any single character NOT in the bracket expression
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	// Subscribe to all user events
//	sub, err := ps.SubscribePattern("user.*", pubsub.SubOptions{
//		BufferSize: 16,
//		Policy:     pubsub.DropOldest,
//	})
func (ps *InProcPubSub) SubscribePattern(pattern string, opts SubOptions) (Subscription, error) {
	return ps.subscribeInternal(context.Background(), pattern, opts, true)
}

// SubscribePatternWithContext creates a pattern subscription that is cancelled when the context is done.
func (ps *InProcPubSub) SubscribePatternWithContext(ctx context.Context, pattern string, opts SubOptions) (Subscription, error) {
	return ps.subscribeInternal(ctx, pattern, opts, true)
}

// subscribeInternal is the internal implementation for all subscribe methods.
func (ps *InProcPubSub) subscribeInternal(ctx context.Context, topic string, opts SubOptions, isPattern bool) (Subscription, error) {
	start := time.Now()
	var err error
	defer func() {
		ps.recordMetrics(ctx, "subscribe", topic, time.Since(start), err)
	}()

	if ps.closed.Load() {
		err = ErrSubscribeToClosed
		return nil, err
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		if isPattern {
			err = ErrInvalidPattern
		} else {
			err = ErrInvalidTopic
		}
		return nil, err
	}

	// Validate pattern syntax
	if isPattern {
		if _, matchErr := path.Match(topic, "probe"); matchErr != nil {
			err = ErrInvalidPattern
			return nil, err
		}
	}

	opts = ps.normalizeSubOptions(opts)
	if opts.BufferSize < 1 {
		err = ErrBufferTooSmall
		return nil, err
	}

	id := ps.nextID.Add(1)

	var cancel context.CancelFunc
	if ctx != nil {
		ctx, cancel = context.WithCancel(ctx)
	}

	sub := &subscriber{
		id:      id,
		topic:   topic,
		opts:    opts,
		ps:      ps,
		pattern: isPattern,
		done:    make(chan struct{}),
		cancel:  cancel,
	}

	// Use ring buffer for DropOldest when enabled
	if opts.UseRingBuffer && opts.Policy == DropOldest {
		sub.ringBuf = newRingBuffer(opts.BufferSize)
		sub.ch = make(chan Message) // unbuffered: pump goroutine handles buffering
		sub.pumpDone = make(chan struct{})
		go sub.pumpRingBuffer()
	} else {
		sub.ch = make(chan Message, opts.BufferSize)
	}

	// Add to sharded map
	if isPattern {
		ps.shards.addPattern(topic, id, sub)
	} else {
		ps.shards.addTopic(topic, id, sub)
	}
	ps.metrics.addSubs(topic, 1)

	// Call hook
	if ps.config.Hooks.OnSubscribe != nil {
		ps.config.Hooks.OnSubscribe(topic, id)
	}

	// Handle context cancellation
	if ctx != nil {
		go func() {
			select {
			case <-ctx.Done():
				sub.Cancel()
			case <-sub.done:
			}
		}()
	}

	return sub, nil
}

// Publish sends a message to a topic.
//
// This method:
//   - Validates the topic
//   - Fills message metadata (topic, timestamp)
//   - Delivers to all subscribers (including pattern matches)
//   - Handles backpressure according to subscriber policy
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	msg := pubsub.Message{
//		Data: map[string]any{
//			"user_id": "123",
//			"email":   "user@example.com",
//		},
//	}
//
//	err := ps.Publish("user.created", msg)
func (ps *InProcPubSub) Publish(topic string, msg Message) error {
	return ps.publishInternal(context.Background(), topic, msg, false)
}

// PublishWithContext publishes a message with context support.
func (ps *InProcPubSub) PublishWithContext(ctx context.Context, topic string, msg Message) error {
	return ps.publishInternal(ctx, topic, msg, false)
}

// PublishAsync publishes without waiting for delivery results (fire-and-forget).
//
// This method returns immediately after queuing the message for delivery,
// without waiting for the actual delivery to subscribers. The message will
// be delivered asynchronously to all matching subscribers.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	// Fire-and-forget publish
//	err := ps.PublishAsync("user.created", msg)
func (ps *InProcPubSub) PublishAsync(topic string, msg Message) error {
	return ps.publishInternal(context.Background(), topic, msg, true)
}

// PublishBatch publishes multiple messages to a topic atomically.
func (ps *InProcPubSub) PublishBatch(topic string, msgs []Message) error {
	if ps.closed.Load() {
		return ErrPublishToClosed
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return ErrInvalidTopic
	}

	for i := range msgs {
		msgs[i].Topic = topic
		if msgs[i].Time.IsZero() {
			msgs[i].Time = time.Now().UTC()
		}
	}

	// Get subscribers once; skip expensive lookups when no subscribers exist
	var subs []*subscriber
	if ps.shards.topicExists(topic) {
		subs = ps.shards.getTopicSubscribers(topic)
	}
	var patterns []patternSnapshot
	if ps.shards.hasAnyPatterns() {
		patterns = ps.shards.getAllPatterns()
	}

	ps.pendingOps.Add(1)
	defer ps.pendingOps.Add(-1)

	bgCtx := context.Background()
	for _, msg := range msgs {
		clonedMsg := cloneMessage(msg)
		ps.metrics.incPublish(topic)

		// Call hook
		if ps.config.Hooks.OnPublish != nil {
			ps.config.Hooks.OnPublish(topic, &clonedMsg)
		}

		// Record message in history
		if ps.history != nil {
			h := ps.history.GetOrCreate(topic, 0)
			h.Add(clonedMsg)
		}

		// Deliver to exact topic subscribers
		for _, s := range subs {
			ps.deliver(bgCtx, s, clonedMsg)
		}

		// Deliver to pattern subscribers
		if len(patterns) > 0 {
			ps.deliverToPatterns(bgCtx, patterns, topic, clonedMsg)
		}
	}

	return nil
}

// PublishMulti publishes messages to multiple topics.
func (ps *InProcPubSub) PublishMulti(msgs map[string][]Message) error {
	if ps.closed.Load() {
		return ErrPublishToClosed
	}

	for topic, topicMsgs := range msgs {
		if err := ps.PublishBatch(topic, topicMsgs); err != nil {
			return err
		}
	}

	return nil
}

// publishInternal is the internal implementation for publish methods.
func (ps *InProcPubSub) publishInternal(ctx context.Context, topic string, msg Message, async bool) error {
	start := time.Now()
	var err error
	defer func() {
		opName := "publish"
		if async {
			opName = "publish_async"
		}
		ps.recordMetrics(ctx, opName, topic, time.Since(start), err)
	}()

	// Check context cancellation early
	if ctx.Err() != nil {
		err = ctx.Err()
		return err
	}

	if ps.closed.Load() {
		err = ErrPublishToClosed
		return err
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		err = ErrInvalidTopic
		return err
	}

	// Fill message metadata
	msg.Topic = topic
	if msg.Time.IsZero() {
		msg.Time = time.Now().UTC()
	}
	msg = cloneMessage(msg)

	ps.metrics.incPublish(topic)

	// Call hook
	if ps.config.Hooks.OnPublish != nil {
		ps.config.Hooks.OnPublish(topic, &msg)
	}

	if async {
		// Use worker pool for async publish with context cancellation support
		ps.pendingOps.Add(1)
		submitted := ps.workerPool.SubmitWithContext(ctx, func() {
			defer ps.pendingOps.Add(-1)
			ps.doPublish(ctx, topic, msg)
		})
		if !submitted {
			ps.pendingOps.Add(-1)
			// Context cancelled or worker pool closed
			if ctx.Err() != nil {
				err = ctx.Err()
				return err
			}
			// Worker pool is closed, do sync publish
			ps.doPublish(ctx, topic, msg)
		}
	} else {
		ps.pendingOps.Add(1)
		defer ps.pendingOps.Add(-1)
		ps.doPublish(ctx, topic, msg)
	}

	return nil
}

// doPublish performs the actual message delivery.
func (ps *InProcPubSub) doPublish(ctx context.Context, topic string, msg Message) {
	// Check TTL expiration from message metadata
	if ps.ttlMgr != nil && msg.Meta != nil {
		if expiresAt, ok := msg.Meta["X-Message-ExpiresAt"]; ok {
			if t, err := time.Parse(time.RFC3339Nano, expiresAt); err == nil && time.Now().After(t) {
				return
			}
		}
	}

	// Check context cancellation before delivery
	if ctx.Err() != nil {
		return
	}

	// Record message in history
	if ps.history != nil {
		h := ps.history.GetOrCreate(topic, 0)
		h.Add(msg)
	}

	// Fast path: skip subscriber lookup when no subscribers exist
	hasTopicSubs := ps.shards.topicExists(topic)
	hasPatternSubs := ps.shards.hasAnyPatterns()

	if !hasTopicSubs && !hasPatternSubs {
		return
	}

	// Deliver to exact topic subscribers
	if hasTopicSubs {
		subs := ps.shards.getTopicSubscribers(topic)
		for _, s := range subs {
			if ctx.Err() != nil {
				return
			}
			ps.deliver(ctx, s, msg)
		}
	}

	// Get and deliver to pattern subscribers
	if hasPatternSubs {
		patterns := ps.shards.getAllPatterns()
		ps.deliverToPatterns(ctx, patterns, topic, msg)
	}
}

// deliverToPatterns delivers a message to all matching pattern subscribers.
func (ps *InProcPubSub) deliverToPatterns(ctx context.Context, patterns []patternSnapshot, topic string, msg Message) {
	for _, entry := range patterns {
		if ctx.Err() != nil {
			return
		}

		// Fast path: check if pattern contains wildcards
		var matched bool
		if strings.ContainsAny(entry.pattern, "*?[]\\") {
			var err error
			matched, err = path.Match(entry.pattern, topic)
			if err != nil {
				matched = false
			}
		} else {
			// No wildcards - just compare strings
			matched = entry.pattern == topic
		}

		if matched {
			for _, s := range entry.subs {
				if ctx.Err() != nil {
					return
				}
				ps.deliver(ctx, s, msg)
			}
		}
	}
}

// GetSubscriberCount returns the number of subscribers for a topic.
func (ps *InProcPubSub) GetSubscriberCount(topic string) int {
	return ps.shards.getTopicSubscriberCount(topic)
}

// GetPatternSubscriberCount returns the number of subscribers for a pattern.
func (ps *InProcPubSub) GetPatternSubscriberCount(pattern string) int {
	return ps.shards.getPatternSubscriberCount(pattern)
}

// TopicExists checks if a topic has any active subscribers.
// This is a lightweight check that only acquires a read lock on one shard.
func (ps *InProcPubSub) TopicExists(topic string) bool {
	return ps.shards.topicExists(topic)
}

// PatternExists checks if a pattern has any active subscribers.
// This is a lightweight check that only acquires a read lock on one shard.
func (ps *InProcPubSub) PatternExists(pattern string) bool {
	return ps.shards.patternExists(pattern)
}

// HasSubscribers checks if publishing to the given topic would reach any subscriber,
// either through an exact topic match or a matching pattern subscription.
func (ps *InProcPubSub) HasSubscribers(topic string) bool {
	// Fast path: check exact topic subscribers
	if ps.shards.topicExists(topic) {
		return true
	}

	// Slow path: check pattern subscribers
	patterns := ps.shards.getAllPatterns()
	for _, entry := range patterns {
		if strings.ContainsAny(entry.pattern, "*?[]\\") {
			if matched, err := path.Match(entry.pattern, topic); err == nil && matched {
				return true
			}
		} else if entry.pattern == topic {
			return true
		}
	}

	return false
}

// ListTopics returns all active topics.
func (ps *InProcPubSub) ListTopics() []string {
	topics := ps.shards.listTopics()
	sort.Strings(topics)
	return topics
}

// ListPatterns returns all active subscription patterns.
func (ps *InProcPubSub) ListPatterns() []string {
	patterns := ps.shards.listPatterns()
	sort.Strings(patterns)
	return patterns
}

// removeSubscriber removes subscription from the sharded map.
func (ps *InProcPubSub) removeSubscriber(topic string, id uint64, pattern bool) {
	if pattern {
		ps.shards.removePattern(topic, id)
	} else {
		ps.shards.removeTopic(topic, id)
	}
	ps.metrics.addSubs(topic, -1)
}

// deliver handles message delivery to a subscriber with backpressure policy.
func (ps *InProcPubSub) deliver(ctx context.Context, s *subscriber, msg Message) {
	// Panic recovery
	if ps.config.EnablePanicRecovery {
		defer func() {
			if r := recover(); r != nil {
				if ps.config.OnPanic != nil {
					ps.config.OnPanic(s.topic, s.id, r)
				}
				s.Cancel()
			}
		}()
	}

	metricTopic := s.topic
	if s.pattern {
		metricTopic = msg.Topic
	}

	// Apply filter if present
	if s.opts.Filter != nil && !s.opts.Filter(msg) {
		return
	}

	// Clone message unless zero-copy is enabled
	if !s.opts.ZeroCopy {
		msg = cloneMessage(msg)
	}

	// Fast path: check if closed without lock
	if s.closed.Load() {
		return
	}

	s.mu.Lock()
	if s.closed.Load() {
		s.mu.Unlock()
		return
	}

	var cancelSub bool

	switch s.opts.Policy {
	case DropOldest:
		cancelSub = ps.deliverDropOldest(s, msg, metricTopic)
	case DropNewest:
		cancelSub = ps.deliverDropNewest(s, msg, metricTopic)
	case BlockWithTimeout:
		cancelSub = ps.deliverBlockWithTimeout(ctx, s, msg, metricTopic)
	case CloseSubscriber:
		cancelSub = ps.deliverCloseSubscriber(s, msg, metricTopic)
	default:
		// Safe default: DropOldest
		cancelSub = ps.deliverDropOldest(s, msg, metricTopic)
	}

	s.mu.Unlock()

	// Cancel outside the lock to avoid deadlock
	if cancelSub {
		s.Cancel()
	}
}

// deliverDropOldest drops the oldest message if buffer is full.
// When the subscriber has a ring buffer, it uses O(1) push with automatic eviction.
// Otherwise, it falls back to the channel drain-and-retry loop.
func (ps *InProcPubSub) deliverDropOldest(s *subscriber, msg Message, metricTopic string) bool {
	// Fast path: ring buffer provides O(1) DropOldest
	if s.ringBuf != nil {
		dropped, wasDropped := s.ringBuf.Push(msg)
		if wasDropped {
			ps.metrics.incDropped(metricTopic, DropOldest)
			s.dropped.Add(1)
			if ps.config.Hooks.OnDrop != nil {
				ps.config.Hooks.OnDrop(metricTopic, s.id, &dropped, DropOldest)
			}
		}
		ps.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		if ps.config.Hooks.OnDeliver != nil {
			ps.config.Hooks.OnDeliver(metricTopic, s.id, &msg)
		}
		return false
	}

	// Fallback: channel-based drain-and-retry
	for {
		select {
		case s.ch <- msg:
			ps.metrics.incDelivered(metricTopic)
			s.received.Add(1)
			// Call deliver hook
			if ps.config.Hooks.OnDeliver != nil {
				ps.config.Hooks.OnDeliver(metricTopic, s.id, &msg)
			}
			return false
		default:
			// Buffer is full, try to drop the oldest
			select {
			case dropped := <-s.ch:
				ps.metrics.incDropped(metricTopic, DropOldest)
				s.dropped.Add(1)
				// Call drop hook
				if ps.config.Hooks.OnDrop != nil {
					ps.config.Hooks.OnDrop(metricTopic, s.id, &dropped, DropOldest)
				}
				// Continue loop to try sending again
			default:
				// Channel was emptied by consumer, try sending again
			}
		}
	}
}

// deliverDropNewest drops the newest message if buffer is full.
func (ps *InProcPubSub) deliverDropNewest(s *subscriber, msg Message, metricTopic string) bool {
	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		if ps.config.Hooks.OnDeliver != nil {
			ps.config.Hooks.OnDeliver(metricTopic, s.id, &msg)
		}
		return false
	default:
		ps.metrics.incDropped(metricTopic, DropNewest)
		s.dropped.Add(1)
		if ps.config.Hooks.OnDrop != nil {
			ps.config.Hooks.OnDrop(metricTopic, s.id, &msg, DropNewest)
		}
		return false
	}
}

// deliverBlockWithTimeout blocks until timeout, context cancellation, or success.
func (ps *InProcPubSub) deliverBlockWithTimeout(ctx context.Context, s *subscriber, msg Message, metricTopic string) bool {
	timeout := s.opts.BlockTimeout
	if timeout <= 0 {
		timeout = ps.config.DefaultBlockTimeout
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		if ps.config.Hooks.OnDeliver != nil {
			ps.config.Hooks.OnDeliver(metricTopic, s.id, &msg)
		}
		return false
	case <-timer.C:
		ps.metrics.incDropped(metricTopic, BlockWithTimeout)
		s.dropped.Add(1)
		if ps.config.Hooks.OnDrop != nil {
			ps.config.Hooks.OnDrop(metricTopic, s.id, &msg, BlockWithTimeout)
		}
		return false
	case <-ctx.Done():
		ps.metrics.incDropped(metricTopic, BlockWithTimeout)
		s.dropped.Add(1)
		if ps.config.Hooks.OnDrop != nil {
			ps.config.Hooks.OnDrop(metricTopic, s.id, &msg, BlockWithTimeout)
		}
		return false
	}
}

// deliverCloseSubscriber closes the subscription if buffer is full.
func (ps *InProcPubSub) deliverCloseSubscriber(s *subscriber, msg Message, metricTopic string) bool {
	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		if ps.config.Hooks.OnDeliver != nil {
			ps.config.Hooks.OnDeliver(metricTopic, s.id, &msg)
		}
		return false
	default:
		ps.metrics.incDropped(metricTopic, CloseSubscriber)
		s.dropped.Add(1)
		if ps.config.Hooks.OnDrop != nil {
			ps.config.Hooks.OnDrop(metricTopic, s.id, &msg, CloseSubscriber)
		}
		return true // signal to cancel subscription
	}
}

// normalizeSubOptions ensures valid subscription options with config defaults.
func (ps *InProcPubSub) normalizeSubOptions(opts SubOptions) SubOptions {
	if opts.BufferSize <= 0 {
		opts.BufferSize = ps.config.DefaultBufferSize
	}

	// Validate policy
	switch opts.Policy {
	case DropOldest, DropNewest, BlockWithTimeout, CloseSubscriber:
	default:
		opts.Policy = ps.config.DefaultPolicy
	}

	// Validate timeout
	if opts.Policy == BlockWithTimeout && opts.BlockTimeout <= 0 {
		opts.BlockTimeout = ps.config.DefaultBlockTimeout
	}

	// Propagate config-level ring buffer setting
	if ps.config.EnableRingBuffer && opts.Policy == DropOldest {
		opts.UseRingBuffer = true
	}

	return opts
}

// Snapshot exposes observability metrics.
func (ps *InProcPubSub) Snapshot() MetricsSnapshot {
	return ps.metrics.Snapshot()
}

// SetMetricsCollector sets the unified metrics collector.
func (ps *InProcPubSub) SetMetricsCollector(collector metrics.MetricsCollector) {
	ps.config.MetricsCollector = collector
}

// GetMetricsCollector returns the current metrics collector.
func (ps *InProcPubSub) GetMetricsCollector() metrics.MetricsCollector {
	return ps.config.MetricsCollector
}

// Config returns a copy of the current configuration.
func (ps *InProcPubSub) Config() Config {
	return ps.config
}

// TopicShard returns the shard index that a topic or pattern is assigned to.
// This is useful for diagnostics and understanding topic distribution across shards.
func (ps *InProcPubSub) TopicShard(topic string) int {
	return ps.shards.getShardIndex(topic)
}

// ShardStats returns per-shard statistics for monitoring shard balance and distribution.
func (ps *InProcPubSub) ShardStats() []ShardStat {
	return ps.shards.shardStats()
}

// TopicShardMapping returns a map of all active topics and patterns to their shard indices.
// This is useful for understanding how topics are distributed across shards.
func (ps *InProcPubSub) TopicShardMapping() map[string]int {
	return ps.shards.topicShardMapping()
}

// recordMetrics records metrics using the unified collector.
func (ps *InProcPubSub) recordMetrics(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	collector := ps.config.MetricsCollector
	if collector == nil {
		return
	}
	collector.ObservePubSub(ctx, operation, topic, duration, err)
}

// PublishDelayed schedules a message for delivery after the specified delay.
// Requires the scheduler to be enabled via WithScheduler() option.
// Returns a schedule ID that can be used to cancel the scheduled message.
func (ps *InProcPubSub) PublishDelayed(topic string, msg Message, delay time.Duration) (uint64, error) {
	if ps.closed.Load() {
		return 0, ErrPublishToClosed
	}
	if ps.scheduler == nil {
		return 0, ErrSchedulerDisabled
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return 0, ErrInvalidTopic
	}

	msg.Topic = topic
	if msg.Time.IsZero() {
		msg.Time = time.Now().UTC()
	}

	id := ps.scheduler.Schedule(topic, msg, delay)
	return id, nil
}

// PublishAt schedules a message for delivery at a specific time.
// Requires the scheduler to be enabled via WithScheduler() option.
// Returns a schedule ID that can be used to cancel the scheduled message.
func (ps *InProcPubSub) PublishAt(topic string, msg Message, at time.Time) (uint64, error) {
	if ps.closed.Load() {
		return 0, ErrPublishToClosed
	}
	if ps.scheduler == nil {
		return 0, ErrSchedulerDisabled
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return 0, ErrInvalidTopic
	}

	msg.Topic = topic
	if msg.Time.IsZero() {
		msg.Time = time.Now().UTC()
	}

	id := ps.scheduler.ScheduleAt(topic, msg, at)
	return id, nil
}

// CancelScheduled cancels a previously scheduled message by its schedule ID.
// Returns true if the message was found and cancelled.
func (ps *InProcPubSub) CancelScheduled(id uint64) bool {
	if ps.scheduler == nil {
		return false
	}
	return ps.scheduler.Cancel(id)
}

// ScheduledCount returns the number of messages pending scheduled delivery.
func (ps *InProcPubSub) ScheduledCount() int {
	if ps.scheduler == nil {
		return 0
	}
	return ps.scheduler.PendingCount()
}

// PublishWithTTL publishes a message with a time-to-live. The message will be
// skipped during delivery if it has expired (useful with delayed delivery).
// Requires the TTL manager to be enabled via WithTTL() option.
func (ps *InProcPubSub) PublishWithTTL(topic string, msg Message, ttl time.Duration) error {
	if ps.closed.Load() {
		return ErrPublishToClosed
	}
	if ps.ttlMgr == nil {
		return ErrTTLDisabled
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return ErrInvalidTopic
	}

	if msg.ID == "" {
		msg.ID = generateCorrelationID()
	}

	// Add TTL metadata
	if msg.Meta == nil {
		msg.Meta = make(map[string]string)
	}
	msg.Meta["X-Message-TTL"] = ttl.String()
	msg.Meta["X-Message-ExpiresAt"] = time.Now().Add(ttl).UTC().Format(time.RFC3339Nano)

	// Track in TTL manager
	ps.ttlMgr.Track(topic, msg.ID, ttl)

	return ps.Publish(topic, msg)
}

// GetTTLStats returns statistics from the TTL manager.
func (ps *InProcPubSub) GetTTLStats() TTLStats {
	if ps.ttlMgr == nil {
		return TTLStats{}
	}
	return ps.ttlMgr.Stats()
}

// GetTopicHistory returns all retained messages for a topic (oldest first).
// Requires history to be enabled via WithHistory() option.
func (ps *InProcPubSub) GetTopicHistory(topic string) ([]Message, error) {
	if ps.history == nil {
		return nil, ErrHistoryDisabled
	}

	h := ps.history.Get(topic)
	if h == nil {
		return nil, nil
	}
	return h.GetAll(), nil
}

// GetTopicHistorySince returns messages added after the given sequence number.
// Requires history to be enabled via WithHistory() option.
func (ps *InProcPubSub) GetTopicHistorySince(topic string, sequence uint64) ([]Message, error) {
	if ps.history == nil {
		return nil, ErrHistoryDisabled
	}

	h := ps.history.Get(topic)
	if h == nil {
		return nil, nil
	}
	return h.GetSince(sequence), nil
}

// GetRecentMessages returns the last N messages for a topic (oldest first).
// Requires history to be enabled via WithHistory() option.
func (ps *InProcPubSub) GetRecentMessages(topic string, count int) ([]Message, error) {
	if ps.history == nil {
		return nil, ErrHistoryDisabled
	}

	h := ps.history.Get(topic)
	if h == nil {
		return nil, nil
	}
	return h.GetLast(count), nil
}

// GetTopicHistoryByTTL returns messages not older than the given TTL duration.
// Requires history to be enabled via WithHistory() option.
func (ps *InProcPubSub) GetTopicHistoryByTTL(topic string, ttl time.Duration) ([]Message, error) {
	if ps.history == nil {
		return nil, ErrHistoryDisabled
	}

	h := ps.history.Get(topic)
	if h == nil {
		return nil, nil
	}
	return h.GetWithTTL(ttl), nil
}

// ClearTopicHistory removes all retained messages for a topic.
// Requires history to be enabled via WithHistory() option.
func (ps *InProcPubSub) ClearTopicHistory(topic string) error {
	if ps.history == nil {
		return ErrHistoryDisabled
	}

	ps.history.Delete(topic)
	return nil
}

// TopicHistoryStats returns history statistics for all topics with retained messages.
// Requires history to be enabled via WithHistory() option.
func (ps *InProcPubSub) TopicHistoryStats() (map[string]HistoryStats, error) {
	if ps.history == nil {
		return nil, ErrHistoryDisabled
	}

	return ps.history.Stats(), nil
}

// TopicHistorySequence returns the current sequence number for a topic's history.
// Returns 0 if no history exists for the topic.
// Requires history to be enabled via WithHistory() option.
func (ps *InProcPubSub) TopicHistorySequence(topic string) (uint64, error) {
	if ps.history == nil {
		return 0, ErrHistoryDisabled
	}

	h := ps.history.Get(topic)
	if h == nil {
		return 0, nil
	}
	return h.CurrentSequence(), nil
}

// Ensure InProcPubSub implements all interfaces
var (
	_ PubSub             = (*InProcPubSub)(nil)
	_ PatternPubSub      = (*InProcPubSub)(nil)
	_ ContextPubSub      = (*InProcPubSub)(nil)
	_ BatchPubSub        = (*InProcPubSub)(nil)
	_ DrainablePubSub    = (*InProcPubSub)(nil)
	_ HistoryPubSub      = (*InProcPubSub)(nil)
	_ RequestReplyPubSub = (*InProcPubSub)(nil)
)
