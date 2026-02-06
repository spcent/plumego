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
		close(s.ch)
		close(s.done)
		if s.cancel != nil {
			s.cancel()
		}
		s.mu.Unlock()

		// Call hook before removing
		if s.ps.config.Hooks.OnUnsubscribe != nil {
			s.ps.config.Hooks.OnUnsubscribe(s.topic, s.id)
		}

		s.ps.removeSubscriber(s.topic, s.id, s.pattern)
	})
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
		ps.recordMetrics("subscribe", topic, time.Since(start), err)
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
		ch:      make(chan Message, opts.BufferSize),
		opts:    opts,
		ps:      ps,
		pattern: isPattern,
		done:    make(chan struct{}),
		cancel:  cancel,
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

	// Get subscribers once
	subs := ps.shards.getTopicSubscribers(topic)
	patterns := ps.shards.getAllPatterns()

	ps.pendingOps.Add(1)
	defer ps.pendingOps.Add(-1)

	for _, msg := range msgs {
		clonedMsg := cloneMessage(msg)
		ps.metrics.incPublish(topic)

		// Call hook
		if ps.config.Hooks.OnPublish != nil {
			ps.config.Hooks.OnPublish(topic, &clonedMsg)
		}

		// Deliver to exact topic subscribers
		for _, s := range subs {
			ps.deliver(s, clonedMsg)
		}

		// Deliver to pattern subscribers
		ps.deliverToPatterns(patterns, topic, clonedMsg)
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
		ps.recordMetrics(opName, topic, time.Since(start), err)
	}()

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
		// Use worker pool for async publish
		ps.pendingOps.Add(1)
		submitted := ps.workerPool.Submit(func() {
			defer ps.pendingOps.Add(-1)
			ps.doPublish(topic, msg)
		})
		if !submitted {
			ps.pendingOps.Add(-1)
			// Worker pool is closed, do sync publish
			ps.doPublish(topic, msg)
		}
	} else {
		ps.pendingOps.Add(1)
		defer ps.pendingOps.Add(-1)
		ps.doPublish(topic, msg)
	}

	return nil
}

// doPublish performs the actual message delivery.
func (ps *InProcPubSub) doPublish(topic string, msg Message) {
	// Get topic subscribers
	subs := ps.shards.getTopicSubscribers(topic)

	// Deliver to exact topic subscribers
	for _, s := range subs {
		ps.deliver(s, msg)
	}

	// Get and deliver to pattern subscribers
	patterns := ps.shards.getAllPatterns()
	ps.deliverToPatterns(patterns, topic, msg)
}

// deliverToPatterns delivers a message to all matching pattern subscribers.
func (ps *InProcPubSub) deliverToPatterns(patterns []patternSnapshot, topic string, msg Message) {
	for _, entry := range patterns {
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
				ps.deliver(s, msg)
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
func (ps *InProcPubSub) deliver(s *subscriber, msg Message) {
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
		cancelSub = ps.deliverBlockWithTimeout(s, msg, metricTopic)
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
// This implementation is thread-safe and handles the race condition properly.
func (ps *InProcPubSub) deliverDropOldest(s *subscriber, msg Message, metricTopic string) bool {
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

// deliverBlockWithTimeout blocks until timeout or success.
func (ps *InProcPubSub) deliverBlockWithTimeout(s *subscriber, msg Message, metricTopic string) bool {
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

// recordMetrics records metrics using the unified collector.
func (ps *InProcPubSub) recordMetrics(operation, topic string, duration time.Duration, err error) {
	collector := ps.config.MetricsCollector
	if collector == nil {
		return
	}
	ctx := context.Background()
	collector.ObservePubSub(ctx, operation, topic, duration, err)
}

// Ensure InProcPubSub implements all interfaces
var (
	_ PubSub          = (*InProcPubSub)(nil)
	_ PatternPubSub   = (*InProcPubSub)(nil)
	_ ContextPubSub   = (*InProcPubSub)(nil)
	_ BatchPubSub     = (*InProcPubSub)(nil)
	_ DrainablePubSub = (*InProcPubSub)(nil)
)
