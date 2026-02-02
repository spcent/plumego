package mq

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
)

// MessagePriority represents the priority of a message.
type MessagePriority int

const (
	PriorityLowest  MessagePriority = 0
	PriorityLow     MessagePriority = 10
	PriorityNormal  MessagePriority = 20
	PriorityHigh    MessagePriority = 30
	PriorityHighest MessagePriority = 40
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
			return ErrNotInitialized
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
