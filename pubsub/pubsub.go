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

	// back-reference to parent
	ps *InProcPubSub
}

type patternSnapshot struct {
	pattern string
	subs    []*subscriber
}

// C returns the receive-only message channel.
func (s *subscriber) C() <-chan Message { return s.ch }

// Cancel unsubscribes and closes the channel.
func (s *subscriber) Cancel() {
	s.once.Do(func() {
		s.mu.Lock()
		if s.closed.Swap(true) {
			s.mu.Unlock()
			return
		}
		close(s.ch)
		s.mu.Unlock()
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
//	// Create a pubsub instance
//	ps := pubsub.New()
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
	mu       sync.RWMutex
	topics   map[string]map[uint64]*subscriber
	patterns map[string]map[uint64]*subscriber

	closed atomic.Bool
	nextID atomic.Uint64

	metrics   metricsPubSub
	collector metrics.MetricsCollector // Unified metrics collector
}

// New creates a new InProcPubSub instance.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
func New() *InProcPubSub {
	ps := &InProcPubSub{
		topics:   make(map[string]map[uint64]*subscriber),
		patterns: make(map[string]map[uint64]*subscriber),
	}
	ps.nextID.Store(0)
	return ps
}

// Close shuts down the pubsub system.
//
// This method:
//   - Marks the pubsub as closed
//   - Cancels all active subscriptions
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

	// Collect all subscribers under lock
	ps.mu.Lock()
	all := make([]*subscriber, 0, 64)
	for _, subs := range ps.topics {
		for _, s := range subs {
			all = append(all, s)
		}
	}
	for _, subs := range ps.patterns {
		for _, s := range subs {
			all = append(all, s)
		}
	}
	// Clear map to reject further deliveries
	ps.topics = make(map[string]map[uint64]*subscriber)
	ps.patterns = make(map[string]map[uint64]*subscriber)
	ps.mu.Unlock()

	// Cancel subscribers without holding lock
	for _, s := range all {
		s.Cancel()
	}

	return nil
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
		err = ErrInvalidTopic
		return nil, err
	}

	opts = normalizeSubOptions(opts)
	if opts.BufferSize < 1 {
		err = ErrBufferTooSmall
		return nil, err
	}

	id := ps.nextID.Add(1)
	sub := &subscriber{
		id:    id,
		topic: topic,
		ch:    make(chan Message, opts.BufferSize),
		opts:  opts,
		ps:    ps,
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed.Load() {
		err = ErrSubscribeToClosed
		return nil, err
	}

	if ps.topics[topic] == nil {
		ps.topics[topic] = make(map[uint64]*subscriber)
	}
	ps.topics[topic][id] = sub
	ps.metrics.addSubs(topic, 1)

	return sub, nil
}

// SubscribePattern creates a new subscription to a topic pattern.
//
// Patterns use filepath.Match syntax:
//   - "*" matches any sequence of non-separator characters
//   - "?" matches any single non-separator character
//   - "[...]" matches any character in the bracket expression
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
//	if err != nil {
//		// Handle error
//	}
//	defer sub.Cancel()
//
//	// Will receive messages from:
//	// - user.created
//	// - user.updated
//	// - user.deleted
func (ps *InProcPubSub) SubscribePattern(pattern string, opts SubOptions) (Subscription, error) {
	start := time.Now()
	var err error
	defer func() {
		ps.recordMetrics("subscribe", pattern, time.Since(start), err)
	}()

	if ps.closed.Load() {
		err = ErrSubscribeToClosed
		return nil, err
	}

	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		err = ErrInvalidPattern
		return nil, err
	}
	if _, err := path.Match(pattern, "probe"); err != nil {
		err = ErrInvalidPattern
		return nil, err
	}

	opts = normalizeSubOptions(opts)
	if opts.BufferSize < 1 {
		err = ErrBufferTooSmall
		return nil, err
	}

	id := ps.nextID.Add(1)
	sub := &subscriber{
		id:      id,
		topic:   pattern,
		ch:      make(chan Message, opts.BufferSize),
		opts:    opts,
		ps:      ps,
		pattern: true,
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed.Load() {
		err = ErrSubscribeToClosed
		return nil, err
	}

	if ps.patterns[pattern] == nil {
		ps.patterns[pattern] = make(map[uint64]*subscriber)
	}
	ps.patterns[pattern][id] = sub
	ps.metrics.addSubs(pattern, 1)

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
//	if err != nil {
//		// Handle error
//	}
func (ps *InProcPubSub) Publish(topic string, msg Message) error {
	start := time.Now()
	var err error
	defer func() {
		ps.recordMetrics("publish", topic, time.Since(start), err)
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

	// Fill message metadata (immutable copy)
	msg.Topic = topic
	if msg.Time.IsZero() {
		msg.Time = time.Now().UTC()
	}
	msg = cloneMessage(msg)

	ps.metrics.incPublish(topic)

	// Snapshot subscribers under RLock
	ps.mu.RLock()
	subsMap := ps.topics[topic]
	var subs []*subscriber
	if len(subsMap) > 0 {
		// Create a snapshot to avoid holding lock during delivery
		subs = make([]*subscriber, 0, len(subsMap))
		for _, s := range subsMap {
			subs = append(subs, s)
		}
	}

	var patterns []patternSnapshot
	if len(ps.patterns) > 0 {
		patterns = make([]patternSnapshot, 0, len(ps.patterns))
		for pattern, subMap := range ps.patterns {
			if len(subMap) == 0 {
				continue
			}
			snapshot := make([]*subscriber, 0, len(subMap))
			for _, s := range subMap {
				snapshot = append(snapshot, s)
			}
			patterns = append(patterns, patternSnapshot{
				pattern: pattern,
				subs:    snapshot,
			})
		}
	}
	ps.mu.RUnlock()

	// Deliver to each subscriber
	for _, s := range subs {
		ps.deliver(s, msg)
	}

	if len(patterns) > 0 {
		for _, entry := range patterns {
			// Fast path: check if pattern contains wildcards
			// If not, we can use simple string comparison
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

	return nil
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
	start := time.Now()
	var err error
	defer func() {
		ps.recordMetrics("publish_async", topic, time.Since(start), err)
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

	// Fill message metadata (immutable copy)
	msg.Topic = topic
	if msg.Time.IsZero() {
		msg.Time = time.Now().UTC()
	}
	msg = cloneMessage(msg)

	ps.metrics.incPublish(topic)

	// Use goroutine to deliver asynchronously
	go func() {
		ps.mu.RLock()
		subsMap := ps.topics[topic]
		var subs []*subscriber
		if len(subsMap) > 0 {
			// Create a snapshot to avoid holding lock during delivery
			subs = make([]*subscriber, 0, len(subsMap))
			for _, s := range subsMap {
				subs = append(subs, s)
			}
		}

		var patterns []patternSnapshot
		if len(ps.patterns) > 0 {
			patterns = make([]patternSnapshot, 0, len(ps.patterns))
			for pattern, subMap := range ps.patterns {
				if len(subMap) == 0 {
					continue
				}
				snapshot := make([]*subscriber, 0, len(subMap))
				for _, s := range subMap {
					snapshot = append(snapshot, s)
				}
				patterns = append(patterns, patternSnapshot{
					pattern: pattern,
					subs:    snapshot,
				})
			}
		}
		ps.mu.RUnlock()

		// Deliver to each subscriber
		for _, s := range subs {
			ps.deliver(s, msg)
		}

		if len(patterns) > 0 {
			for _, entry := range patterns {
				if matched, err := path.Match(entry.pattern, topic); err == nil && matched {
					for _, s := range entry.subs {
						ps.deliver(s, msg)
					}
				}
			}
		}
	}()

	return nil
}

// GetSubscriberCount returns the number of subscribers for a topic.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	count := ps.GetSubscriberCount("user.created")
//	fmt.Printf("Topic has %d subscribers\n", count)
func (ps *InProcPubSub) GetSubscriberCount(topic string) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	subs := ps.topics[topic]
	if subs == nil {
		return 0
	}
	return len(subs)
}

// GetPatternSubscriberCount returns the number of subscribers for a pattern.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	count := ps.GetPatternSubscriberCount("user.*")
//	fmt.Printf("Pattern has %d subscribers\n", count)
func (ps *InProcPubSub) GetPatternSubscriberCount(pattern string) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	subs := ps.patterns[pattern]
	if subs == nil {
		return 0
	}
	return len(subs)
}

// ListTopics returns all active topics.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	topics := ps.ListTopics()
//	for _, topic := range topics {
//		fmt.Printf("Topic: %s, Subscribers: %d\n", topic, ps.GetSubscriberCount(topic))
//	}
func (ps *InProcPubSub) ListTopics() []string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	topics := make([]string, 0, len(ps.topics))
	for topic := range ps.topics {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	return topics
}

// ListPatterns returns all active subscription patterns.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	patterns := ps.ListPatterns()
//	for _, pattern := range patterns {
//		fmt.Printf("Pattern: %s, Subscribers: %d\n", pattern, ps.GetPatternSubscriberCount(pattern))
//	}
func (ps *InProcPubSub) ListPatterns() []string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	patterns := make([]string, 0, len(ps.patterns))
	for pattern := range ps.patterns {
		patterns = append(patterns, pattern)
	}
	sort.Strings(patterns)
	return patterns
}

// removeSubscriber removes subscription from topic map; safe for concurrent Cancel/Close.
func (ps *InProcPubSub) removeSubscriber(topic string, id uint64, pattern bool) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	var subs map[uint64]*subscriber
	if pattern {
		subs = ps.patterns[topic]
	} else {
		subs = ps.topics[topic]
	}
	if subs == nil {
		return
	}

	if _, ok := subs[id]; ok {
		delete(subs, id)
		ps.metrics.addSubs(topic, -1)
	}

	if len(subs) == 0 {
		if pattern {
			delete(ps.patterns, topic)
		} else {
			delete(ps.topics, topic)
		}
	}
}

// deliver handles message delivery to a subscriber with backpressure policy.
func (ps *InProcPubSub) deliver(s *subscriber, msg Message) {
	metricTopic := s.topic
	if s.pattern {
		metricTopic = msg.Topic
	}

	msg = cloneMessage(msg)

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
// This implementation uses a buffered channel with a fixed size and implements
// a proper ring buffer semantics by dropping the oldest message when full.
func (ps *InProcPubSub) deliverDropOldest(s *subscriber, msg Message, metricTopic string) bool {
	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(metricTopic)
		return false
	default:
		// Buffer is full, we need to drop the oldest message and add the new one
		// Since Go channels don't guarantee FIFO order for buffered channels,
		// we implement a simple approach: drain one message and try again
		select {
		case <-s.ch:
			// Successfully drained one message (conceptually the oldest)
			ps.metrics.incDropped(metricTopic, DropOldest)
		default:
			// This shouldn't happen since we know the channel is full
			ps.metrics.incDropped(metricTopic, DropOldest)
		}

		// Now try to send the new message
		select {
		case s.ch <- msg:
			ps.metrics.incDelivered(metricTopic)
		default:
			// Still full, drop this message too
			ps.metrics.incDropped(metricTopic, DropOldest)
		}
		return false
	}
}

// deliverDropNewest drops the newest message if buffer is full.
func (ps *InProcPubSub) deliverDropNewest(s *subscriber, msg Message, metricTopic string) bool {
	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(metricTopic)
		return false
	default:
		ps.metrics.incDropped(metricTopic, DropNewest)
		return false
	}
}

// deliverBlockWithTimeout blocks until timeout or success.
func (ps *InProcPubSub) deliverBlockWithTimeout(s *subscriber, msg Message, metricTopic string) bool {
	timeout := s.opts.BlockTimeout
	if timeout <= 0 {
		timeout = 50 * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(metricTopic)
		return false
	case <-ctx.Done():
		ps.metrics.incDropped(metricTopic, BlockWithTimeout)
		return false
	}
}

// deliverCloseSubscriber closes the subscription if buffer is full.
func (ps *InProcPubSub) deliverCloseSubscriber(s *subscriber, msg Message, metricTopic string) bool {
	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(metricTopic)
		return false
	default:
		ps.metrics.incDropped(metricTopic, CloseSubscriber)
		return true // signal to cancel subscription
	}
}

// normalizeSubOptions ensures valid subscription options.
func normalizeSubOptions(opts SubOptions) SubOptions {
	if opts.BufferSize <= 0 {
		opts.BufferSize = 16
	}

	// Validate policy
	switch opts.Policy {
	case DropOldest, DropNewest, BlockWithTimeout, CloseSubscriber:
	default:
		opts.Policy = DropOldest
	}

	// Validate timeout
	if opts.Policy == BlockWithTimeout && opts.BlockTimeout <= 0 {
		opts.BlockTimeout = 50 * time.Millisecond
	}

	return opts
}

// Snapshot exposes observability metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	snapshot := ps.Snapshot()
//	fmt.Printf("Total messages: %d\n", snapshot.TotalPublished)
func (ps *InProcPubSub) Snapshot() MetricsSnapshot {
	return ps.metrics.Snapshot()
}

// SetMetricsCollector sets the unified metrics collector.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//	import "github.com/spcent/plumego/metrics"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	collector := metrics.NewPrometheusCollector()
//	ps.SetMetricsCollector(collector)
func (ps *InProcPubSub) SetMetricsCollector(collector metrics.MetricsCollector) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.collector = collector
}

// GetMetricsCollector returns the current metrics collector.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	collector := ps.GetMetricsCollector()
func (ps *InProcPubSub) GetMetricsCollector() metrics.MetricsCollector {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.collector
}

// recordMetrics records metrics using the unified collector.
func (ps *InProcPubSub) recordMetrics(operation, topic string, duration time.Duration, err error) {
	ps.mu.RLock()
	collector := ps.collector
	ps.mu.RUnlock()
	if collector == nil {
		return
	}
	ctx := context.Background()
	collector.ObservePubSub(ctx, operation, topic, duration, err)
}
