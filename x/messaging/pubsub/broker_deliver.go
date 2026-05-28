package pubsub

import (
	"context"
	"runtime"
	"time"
)


// deliver handles message delivery to a single subscriber with the configured backpressure policy.
func (b *InProcBroker) deliver(ctx context.Context, s *subscriber, msg Message) {
	if b.config.EnablePanicRecovery {
		defer func() {
			if r := recover(); r != nil {
				if b.config.OnPanic != nil {
					b.config.OnPanic(s.topic, s.id, r)
				}
				s.Cancel()
			}
		}()
	}

	metricTopic := s.topic
	if s.kind != subKindTopic {
		metricTopic = msg.Topic
	}

	if s.opts.Filter != nil && !s.opts.Filter(msg) {
		return
	}

	if !s.opts.ZeroCopy {
		msg = cloneMessage(msg)
	}

	if s.closed.Load() {
		return
	}

	if s.opts.Policy == BlockWithTimeout {
		if b.deliverBlockWithTimeout(ctx, s, msg, metricTopic) {
			s.Cancel()
		}
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
		cancelSub = b.deliverDropOldest(s, msg, metricTopic)
	case DropNewest:
		cancelSub = b.deliverDropNewest(s, msg, metricTopic)
	case CloseSubscriber:
		cancelSub = b.deliverCloseSubscriber(s, msg, metricTopic)
	default:
		cancelSub = b.deliverDropOldest(s, msg, metricTopic)
	}

	s.mu.Unlock()

	if cancelSub {
		s.Cancel()
	}
}

// deliverDropOldest drops the oldest message when the buffer is full.
func (b *InProcBroker) deliverDropOldest(s *subscriber, msg Message, metricTopic string) bool {
	if s.ringBuf != nil {
		dropped, wasDropped := s.ringBuf.Push(msg)
		if wasDropped {
			b.metrics.incDropped(metricTopic, DropOldest)
			s.dropped.Add(1)
			b.notifyDrop(metricTopic, s.id, &dropped, DropOldest)
		}
		b.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		b.notifyDeliver(metricTopic, s.id, &msg)
		return false
	}

	for spinCount := 0; ; spinCount++ {
		if s.closed.Load() {
			return false
		}
		select {
		case s.ch <- msg:
			b.metrics.incDelivered(metricTopic)
			s.received.Add(1)
			b.notifyDeliver(metricTopic, s.id, &msg)
			return false
		default:
			select {
			case dropped := <-s.ch:
				b.metrics.incDropped(metricTopic, DropOldest)
				s.dropped.Add(1)
				b.notifyDrop(metricTopic, s.id, &dropped, DropOldest)
			default:
				runtime.Gosched()
			}
		}
		if spinCount > 0 && spinCount%64 == 0 {
			runtime.Gosched()
		}
	}
}

// deliverDropNewest discards the incoming message when the buffer is full.
func (b *InProcBroker) deliverDropNewest(s *subscriber, msg Message, metricTopic string) bool {
	select {
	case s.ch <- msg:
		b.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		b.notifyDeliver(metricTopic, s.id, &msg)
		return false
	default:
		b.metrics.incDropped(metricTopic, DropNewest)
		s.dropped.Add(1)
		b.notifyDrop(metricTopic, s.id, &msg, DropNewest)
		return false
	}
}

// deliverBlockWithTimeout blocks until buffer space is available or the deadline elapses.
func (b *InProcBroker) deliverBlockWithTimeout(ctx context.Context, s *subscriber, msg Message, metricTopic string) bool {
	timeout := s.opts.BlockTimeout
	if timeout <= 0 {
		timeout = b.config.DefaultBlockTimeout
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case s.ch <- msg:
		b.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		b.notifyDeliver(metricTopic, s.id, &msg)
		return false
	case <-timer.C:
		b.metrics.incDropped(metricTopic, BlockWithTimeout)
		s.dropped.Add(1)
		b.notifyDrop(metricTopic, s.id, &msg, BlockWithTimeout)
		return false
	case <-ctx.Done():
		b.metrics.incDropped(metricTopic, BlockWithTimeout)
		s.dropped.Add(1)
		b.notifyDrop(metricTopic, s.id, &msg, BlockWithTimeout)
		return false
	}
}

// deliverCloseSubscriber closes the subscription when the buffer is full.
func (b *InProcBroker) deliverCloseSubscriber(s *subscriber, msg Message, metricTopic string) bool {
	select {
	case s.ch <- msg:
		b.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		b.notifyDeliver(metricTopic, s.id, &msg)
		return false
	default:
		b.metrics.incDropped(metricTopic, CloseSubscriber)
		s.dropped.Add(1)
		b.notifyDrop(metricTopic, s.id, &msg, CloseSubscriber)
		return true
	}
}
