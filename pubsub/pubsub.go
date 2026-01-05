package pubsub

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// subscriber represents a subscription to a topic.
type subscriber struct {
	id     uint64
	topic  string
	ch     chan Message
	opts   SubOptions
	closed atomic.Bool
	once   sync.Once
	mu     sync.Mutex

	// back-reference to parent
	ps *InProcPubSub
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
		s.ps.removeSubscriber(s.topic, s.id)
	})
}

// InProcPubSub is an in-memory pubsub implementation.
type InProcPubSub struct {
	mu     sync.RWMutex
	topics map[string]map[uint64]*subscriber

	closed atomic.Bool
	nextID atomic.Uint64

	metrics metrics
}

// New creates a new InProcPubSub instance.
func New() *InProcPubSub {
	ps := &InProcPubSub{
		topics: make(map[string]map[uint64]*subscriber),
	}
	ps.nextID.Store(0)
	return ps
}

// Close shuts down the pubsub system.
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
	// Clear map to reject further deliveries
	ps.topics = make(map[string]map[uint64]*subscriber)
	ps.mu.Unlock()

	// Cancel subscribers without holding lock
	for _, s := range all {
		s.Cancel()
	}

	return nil
}

// Subscribe creates a new subscription to a topic.
func (ps *InProcPubSub) Subscribe(topic string, opts SubOptions) (Subscription, error) {
	if ps.closed.Load() {
		return nil, ErrSubscribeToClosed
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil, ErrInvalidTopic
	}

	opts = normalizeSubOptions(opts)
	if opts.BufferSize < 1 {
		return nil, ErrBufferTooSmall
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
		return nil, ErrSubscribeToClosed
	}

	if ps.topics[topic] == nil {
		ps.topics[topic] = make(map[uint64]*subscriber)
	}
	ps.topics[topic][id] = sub
	ps.metrics.addSubs(topic, 1)

	return sub, nil
}

// Publish sends a message to a topic.
func (ps *InProcPubSub) Publish(topic string, msg Message) error {
	if ps.closed.Load() {
		return ErrPublishToClosed
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return ErrInvalidTopic
	}

	// Fill message metadata (immutable copy)
	msg.Topic = topic
	if msg.Time.IsZero() {
		msg.Time = time.Now().UTC()
	}

	ps.metrics.incPublish(topic)

	// Snapshot subscribers under RLock
	ps.mu.RLock()
	subsMap := ps.topics[topic]
	if len(subsMap) == 0 {
		ps.mu.RUnlock()
		return nil
	}

	// Create a snapshot to avoid holding lock during delivery
	subs := make([]*subscriber, 0, len(subsMap))
	for _, s := range subsMap {
		subs = append(subs, s)
	}
	ps.mu.RUnlock()

	// Deliver to each subscriber
	for _, s := range subs {
		ps.deliver(s, msg)
	}

	return nil
}

// PublishAsync publishes without waiting for delivery results (fire-and-forget).
func (ps *InProcPubSub) PublishAsync(topic string, msg Message) error {
	return ps.Publish(topic, msg)
}

// GetSubscriberCount returns the number of subscribers for a topic.
func (ps *InProcPubSub) GetSubscriberCount(topic string) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	subs := ps.topics[topic]
	if subs == nil {
		return 0
	}
	return len(subs)
}

// ListTopics returns all active topics.
func (ps *InProcPubSub) ListTopics() []string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	topics := make([]string, 0, len(ps.topics))
	for topic := range ps.topics {
		topics = append(topics, topic)
	}
	return topics
}

// removeSubscriber removes subscription from topic map; safe for concurrent Cancel/Close.
func (ps *InProcPubSub) removeSubscriber(topic string, id uint64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	subs := ps.topics[topic]
	if subs == nil {
		return
	}

	if _, ok := subs[id]; ok {
		delete(subs, id)
		ps.metrics.addSubs(topic, -1)
	}

	if len(subs) == 0 {
		delete(ps.topics, topic)
	}
}

// deliver handles message delivery to a subscriber with backpressure policy.
func (ps *InProcPubSub) deliver(s *subscriber, msg Message) {
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
		cancelSub = ps.deliverDropOldest(s, msg)
	case DropNewest:
		cancelSub = ps.deliverDropNewest(s, msg)
	case BlockWithTimeout:
		cancelSub = ps.deliverBlockWithTimeout(s, msg)
	case CloseSubscriber:
		cancelSub = ps.deliverCloseSubscriber(s, msg)
	default:
		// Safe default: DropOldest
		cancelSub = ps.deliverDropOldest(s, msg)
	}

	s.mu.Unlock()

	// Cancel outside the lock to avoid deadlock
	if cancelSub {
		s.Cancel()
	}
}

// deliverDropOldest drops the oldest message if buffer is full.
func (ps *InProcPubSub) deliverDropOldest(s *subscriber, msg Message) bool {
	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(s.topic)
		return false
	default:
		// Buffer full, drop oldest
		select {
		case <-s.ch:
			ps.metrics.incDropped(s.topic, DropOldest)
		default:
			// Shouldn't happen, but safe
		}

		// Try again
		select {
		case s.ch <- msg:
			ps.metrics.incDelivered(s.topic)
		default:
			// Still full after dropping one, drop this message
			ps.metrics.incDropped(s.topic, DropOldest)
		}
		return false
	}
}

// deliverDropNewest drops the newest message if buffer is full.
func (ps *InProcPubSub) deliverDropNewest(s *subscriber, msg Message) bool {
	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(s.topic)
		return false
	default:
		ps.metrics.incDropped(s.topic, DropNewest)
		return false
	}
}

// deliverBlockWithTimeout blocks until timeout or success.
func (ps *InProcPubSub) deliverBlockWithTimeout(s *subscriber, msg Message) bool {
	timeout := s.opts.BlockTimeout
	if timeout <= 0 {
		timeout = 50 * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(s.topic)
		return false
	case <-ctx.Done():
		ps.metrics.incDropped(s.topic, BlockWithTimeout)
		return false
	}
}

// deliverCloseSubscriber closes the subscription if buffer is full.
func (ps *InProcPubSub) deliverCloseSubscriber(s *subscriber, msg Message) bool {
	select {
	case s.ch <- msg:
		ps.metrics.incDelivered(s.topic)
		return false
	default:
		ps.metrics.incDropped(s.topic, CloseSubscriber)
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
func (ps *InProcPubSub) Snapshot() MetricsSnapshot {
	return ps.metrics.Snapshot()
}
