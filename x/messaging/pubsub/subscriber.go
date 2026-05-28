package pubsub

import (
	"context"
	"sync"
	"sync/atomic"
)

// subKind distinguishes the type of subscription.
type subKind int

const (
	subKindTopic   subKind = iota // exact topic subscription
	subKindPattern                // glob/filepath pattern subscription
	subKindMQTT                   // MQTT-style pattern subscription
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
	kind   subKind
	done   chan struct{}

	// Ring buffer for O(1) DropOldest (nil when not using ring buffer)
	ringBuf  *ringBuffer
	pumpDone chan struct{}

	// Statistics
	received atomic.Uint64
	dropped  atomic.Uint64

	// back-reference to parent
	ps *InProcBroker

	// context cancellation
	cancel context.CancelFunc
}

type patternSnapshot struct {
	pattern string
	isGlob  bool // true when pattern contains glob metacharacters (*?[]\\)
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
// sync.Once guarantees this runs at most once; no secondary closed check needed.
func (s *subscriber) Cancel() {
	s.once.Do(func() {
		s.mu.Lock()
		s.closed.Store(true)
		if s.ringBuf != nil {
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

		s.ps.notifyUnsubscribe(s.topic, s.id)
		s.ps.removeSubscriber(s.topic, s.id, s.kind)
	})
}

// pumpRingBuffer moves messages from the ring buffer to the output channel.
// Termination is driven exclusively by s.done and ringBuf.Notify() closure.
func (s *subscriber) pumpRingBuffer() {
	defer close(s.ch)
	defer close(s.pumpDone)

	for {
		for {
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

		select {
		case _, ok := <-s.ringBuf.Notify():
			if !ok {
				return
			}
		case <-s.done:
			return
		}
	}
}


// initSubscriberChannel sets up the subscriber's message channel.
func initSubscriberChannel(sub *subscriber, opts SubOptions) {
	if opts.UseRingBuffer && opts.Policy == DropOldest {
		sub.ringBuf = newRingBuffer(opts.BufferSize)
		sub.ch = make(chan Message)
		sub.pumpDone = make(chan struct{})
		go sub.pumpRingBuffer()
	} else {
		sub.ch = make(chan Message, opts.BufferSize)
	}
}

