package pubsub

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type subscriber struct {
	id     uint64
	topic  string
	ch     chan Message
	opts   SubOptions
	closed atomic.Bool
	once   sync.Once

	// back-reference
	ps *InProcPubSub
}

func (s *subscriber) C() <-chan Message { return s.ch }

func (s *subscriber) Cancel() {
	s.once.Do(func() {
		if s.closed.Swap(true) {
			return
		}
		s.ps.removeSubscriber(s.topic, s.id)
		close(s.ch)
	})
}

type InProcPubSub struct {
	mu     sync.RWMutex
	topics map[string]map[uint64]*subscriber

	closed atomic.Bool
	nextID atomic.Uint64

	metrics metrics
}

func New() *InProcPubSub {
	ps := &InProcPubSub{
		topics: make(map[string]map[uint64]*subscriber),
	}
	ps.nextID.Store(0)
	return ps
}

func (ps *InProcPubSub) Close() error {
	if ps.closed.Swap(true) {
		return nil
	}

	// Copy and close outside locks? Closing needs ownership; we can do it under lock safely,
	// but avoid calling subscriber.Cancel() under lock because it also mutates map.
	ps.mu.Lock()
	all := make([]*subscriber, 0, 64)
	for _, subs := range ps.topics {
		for _, s := range subs {
			all = append(all, s)
		}
	}
	// clear map to reject further delivery
	ps.topics = make(map[string]map[uint64]*subscriber)
	ps.mu.Unlock()

	// Now cancel/close each subscription without holding ps.mu.
	for _, s := range all {
		s.Cancel()
	}

	return nil
}

func (ps *InProcPubSub) Subscribe(topic string, opts SubOptions) (Subscription, error) {
	if ps.closed.Load() {
		return nil, ErrClosed
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil, ErrInvalidTopic
	}

	opts = normalizeSubOptions(opts)
	if opts.BufferSize <= 0 {
		return nil, ErrInvalidOpts
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
		return nil, ErrClosed
	}
	if ps.topics[topic] == nil {
		ps.topics[topic] = make(map[uint64]*subscriber)
	}
	ps.topics[topic][id] = sub
	ps.metrics.addSubs(topic, 1)
	return sub, nil
}

func (ps *InProcPubSub) Publish(topic string, msg Message) error {
	if ps.closed.Load() {
		return ErrClosed
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return ErrInvalidTopic
	}

	// fill message metadata (do not mutate caller's meta map)
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
	subs := make([]*subscriber, 0, len(subsMap))
	for _, s := range subsMap {
		subs = append(subs, s)
	}
	ps.mu.RUnlock()

	// Deliver without holding lock
	for _, s := range subs {
		ps.deliver(s, msg)
	}
	return nil
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

func (ps *InProcPubSub) deliver(s *subscriber, msg Message) {
	// If subscriber already closed, ignore.
	if s.closed.Load() {
		return
	}

	switch s.opts.Policy {
	case DropOldest:
		// Try non-blocking send; if full, evict one oldest then try again.
		select {
		case s.ch <- msg:
			ps.metrics.incDelivered(s.topic)
			return
		default:
			// channel full -> drop oldest by reading one
			select {
			case <-s.ch:
			default:
			}
			select {
			case s.ch <- msg:
				ps.metrics.incDelivered(s.topic)
			default:
				// still cannot send (should be rare)
				ps.metrics.incDropped(s.topic, DropOldest)
			}
		}
		ps.metrics.incDropped(s.topic, DropOldest)
		return

	case DropNewest:
		select {
		case s.ch <- msg:
			ps.metrics.incDelivered(s.topic)
		default:
			ps.metrics.incDropped(s.topic, DropNewest)
		}
		return

	case BlockWithTimeout:
		to := s.opts.BlockTimeout
		if to <= 0 {
			to = 50 * time.Millisecond
		}
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()

		select {
		case s.ch <- msg:
			ps.metrics.incDelivered(s.topic)
		case <-ctx.Done():
			ps.metrics.incDropped(s.topic, BlockWithTimeout)
		}
		return

	case CloseSubscriber:
		select {
		case s.ch <- msg:
			ps.metrics.incDelivered(s.topic)
		default:
			// slow consumer -> close subscription
			ps.metrics.incDropped(s.topic, CloseSubscriber)
			s.Cancel()
		}
		return

	default:
		// safe default: DropOldest
		select {
		case s.ch <- msg:
			ps.metrics.incDelivered(s.topic)
		default:
			select {
			case <-s.ch:
			default:
			}
			select {
			case s.ch <- msg:
				ps.metrics.incDelivered(s.topic)
			default:
				ps.metrics.incDropped(s.topic, DropOldest)
			}
			ps.metrics.incDropped(s.topic, DropOldest)
		}
	}
}

func normalizeSubOptions(opts SubOptions) SubOptions {
	if opts.BufferSize <= 0 {
		opts.BufferSize = 16
	}
	// default policy
	switch opts.Policy {
	case DropOldest, DropNewest, BlockWithTimeout, CloseSubscriber:
	default:
		opts.Policy = DropOldest
	}
	return opts
}

// Snapshot exposes observability metrics.
// Not part of the PRD interface, but safe as an optional method.
func (ps *InProcPubSub) Snapshot() MetricsSnapshot {
	return ps.metrics.Snapshot()
}
