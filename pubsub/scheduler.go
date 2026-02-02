package pubsub

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// delayedMessage represents a message scheduled for future delivery.
type delayedMessage struct {
	topic  string
	msg    Message
	fireAt time.Time
	index  int // heap index
	id     uint64
}

// delayedHeap implements heap.Interface for delayed messages.
type delayedHeap []*delayedMessage

func (h delayedHeap) Len() int           { return len(h) }
func (h delayedHeap) Less(i, j int) bool { return h[i].fireAt.Before(h[j].fireAt) }
func (h delayedHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *delayedHeap) Push(x any) {
	n := len(*h)
	item := x.(*delayedMessage)
	item.index = n
	*h = append(*h, item)
}

func (h *delayedHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

// messageScheduler handles delayed message delivery and TTL expiration.
type messageScheduler struct {
	mu       sync.Mutex
	heap     delayedHeap
	notify   chan struct{}
	closed   atomic.Bool
	nextID   atomic.Uint64
	ps       *InProcPubSub
	cancelFn context.CancelFunc

	// Pending delayed messages by ID (for cancellation)
	pending sync.Map // map[uint64]*delayedMessage
}

// newMessageScheduler creates a new message scheduler.
func newMessageScheduler(ps *InProcPubSub) *messageScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	s := &messageScheduler{
		heap:     make(delayedHeap, 0),
		notify:   make(chan struct{}, 1),
		ps:       ps,
		cancelFn: cancel,
	}

	heap.Init(&s.heap)
	go s.run(ctx)

	return s
}

// Schedule schedules a message for delayed delivery.
// Returns a unique ID that can be used to cancel the scheduled message.
func (s *messageScheduler) Schedule(topic string, msg Message, delay time.Duration) uint64 {
	if s.closed.Load() {
		return 0
	}

	id := s.nextID.Add(1)
	dm := &delayedMessage{
		topic:  topic,
		msg:    msg,
		fireAt: time.Now().Add(delay),
		id:     id,
	}

	s.mu.Lock()
	heap.Push(&s.heap, dm)
	s.mu.Unlock()

	s.pending.Store(id, dm)

	// Notify the scheduler
	select {
	case s.notify <- struct{}{}:
	default:
	}

	return id
}

// ScheduleAt schedules a message for delivery at a specific time.
func (s *messageScheduler) ScheduleAt(topic string, msg Message, at time.Time) uint64 {
	delay := time.Until(at)
	if delay < 0 {
		delay = 0
	}
	return s.Schedule(topic, msg, delay)
}

// Cancel cancels a scheduled message by ID.
// Returns true if the message was found and cancelled.
func (s *messageScheduler) Cancel(id uint64) bool {
	v, ok := s.pending.LoadAndDelete(id)
	if !ok {
		return false
	}

	dm := v.(*delayedMessage)

	s.mu.Lock()
	defer s.mu.Unlock()

	if dm.index >= 0 && dm.index < len(s.heap) && s.heap[dm.index] == dm {
		heap.Remove(&s.heap, dm.index)
		return true
	}

	return false
}

// PendingCount returns the number of pending delayed messages.
func (s *messageScheduler) PendingCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.heap)
}

// Close stops the scheduler.
func (s *messageScheduler) Close() {
	if s.closed.Swap(true) {
		return
	}
	s.cancelFn()
	close(s.notify)
}

// run is the main scheduler loop.
func (s *messageScheduler) run(ctx context.Context) {
	timer := time.NewTimer(time.Hour)
	timer.Stop()

	for {
		s.mu.Lock()
		var nextFire time.Time
		if len(s.heap) > 0 {
			nextFire = s.heap[0].fireAt
		}
		s.mu.Unlock()

		if !nextFire.IsZero() {
			delay := time.Until(nextFire)
			if delay <= 0 {
				s.fireNext()
				continue
			}
			timer.Reset(delay)
		}

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-s.notify:
			timer.Stop()
		case <-timer.C:
			s.fireNext()
		}
	}
}

// fireNext fires the next scheduled message if it's due.
func (s *messageScheduler) fireNext() {
	s.mu.Lock()
	if len(s.heap) == 0 {
		s.mu.Unlock()
		return
	}

	dm := s.heap[0]
	if time.Now().Before(dm.fireAt) {
		s.mu.Unlock()
		return
	}

	heap.Pop(&s.heap)
	s.mu.Unlock()

	s.pending.Delete(dm.id)

	// Publish the message
	if s.ps != nil && !s.ps.closed.Load() {
		_ = s.ps.Publish(dm.topic, dm.msg)
	}
}

// ttlManager handles message TTL expiration.
type ttlManager struct {
	mu         sync.Mutex
	messages   map[string]map[string]time.Time // topic -> msgID -> expiresAt
	ps         *InProcPubSub
	closed     atomic.Bool
	cancelFn   context.CancelFunc
	cleanupInt time.Duration
}

// newTTLManager creates a new TTL manager.
func newTTLManager(ps *InProcPubSub, cleanupInterval time.Duration) *ttlManager {
	if cleanupInterval <= 0 {
		cleanupInterval = time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	tm := &ttlManager{
		messages:   make(map[string]map[string]time.Time),
		ps:         ps,
		cancelFn:   cancel,
		cleanupInt: cleanupInterval,
	}

	go tm.run(ctx)

	return tm
}

// Track tracks a message with TTL.
func (tm *ttlManager) Track(topic, msgID string, ttl time.Duration) {
	if tm.closed.Load() || ttl <= 0 {
		return
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.messages[topic] == nil {
		tm.messages[topic] = make(map[string]time.Time)
	}
	tm.messages[topic][msgID] = time.Now().Add(ttl)
}

// IsExpired checks if a message has expired.
func (tm *ttlManager) IsExpired(topic, msgID string) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if msgs, ok := tm.messages[topic]; ok {
		if expiresAt, ok := msgs[msgID]; ok {
			return time.Now().After(expiresAt)
		}
	}
	return false
}

// Close stops the TTL manager.
func (tm *ttlManager) Close() {
	if tm.closed.Swap(true) {
		return
	}
	tm.cancelFn()
}

// run is the cleanup loop.
func (tm *ttlManager) run(ctx context.Context) {
	ticker := time.NewTicker(tm.cleanupInt)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tm.cleanup()
		}
	}
}

// cleanup removes expired entries.
func (tm *ttlManager) cleanup() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now()
	for topic, msgs := range tm.messages {
		for msgID, expiresAt := range msgs {
			if now.After(expiresAt) {
				delete(msgs, msgID)
			}
		}
		if len(msgs) == 0 {
			delete(tm.messages, topic)
		}
	}
}

// Stats returns TTL manager statistics.
func (tm *ttlManager) Stats() TTLStats {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var totalTracked int
	for _, msgs := range tm.messages {
		totalTracked += len(msgs)
	}

	return TTLStats{
		TrackedTopics:   len(tm.messages),
		TrackedMessages: totalTracked,
	}
}

// TTLStats contains TTL manager statistics.
type TTLStats struct {
	TrackedTopics   int
	TrackedMessages int
}
