package pubsub

import (
	"sync"
	"sync/atomic"
)

// ringBuffer is a thread-safe ring buffer for messages.
// It provides O(1) push/pop operations with proper DropOldest semantics.
// count is maintained as both a plain int (under mu) for push/pop logic and as
// an atomic int32 so that Len() can be read without acquiring the mutex.
type ringBuffer struct {
	mu       sync.Mutex
	buf      []Message
	head     int  // index of the oldest message
	tail     int  // index where next message will be written
	count    int  // current number of messages (protected by mu)
	capacity int  // maximum capacity
	closed   bool // whether the buffer is closed

	// atomicCount mirrors count for lock-free Len() reads.
	atomicCount atomic.Int32

	// Notification channel for consumers
	notify chan struct{}
}

// newRingBuffer creates a new ring buffer with the given capacity.
func newRingBuffer(capacity int) *ringBuffer {
	if capacity < 1 {
		capacity = 1
	}
	return &ringBuffer{
		buf:      make([]Message, capacity),
		capacity: capacity,
		notify:   make(chan struct{}, 1),
	}
}

// Push adds a message to the buffer.
// If the buffer is full, it drops the oldest message and returns it.
// Returns (dropped message, true) if a message was dropped, (zero, false) otherwise.
func (rb *ringBuffer) Push(msg Message) (Message, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.closed {
		return Message{}, false
	}

	var dropped Message
	var wasDropped bool

	if rb.count == rb.capacity {
		// Buffer is full, drop the oldest message
		dropped = rb.buf[rb.head]
		wasDropped = true
		rb.buf[rb.head] = Message{} // clear for GC
		rb.head = (rb.head + 1) % rb.capacity
		rb.count--
		rb.atomicCount.Add(-1)
	}

	// Add the new message
	rb.buf[rb.tail] = msg
	rb.tail = (rb.tail + 1) % rb.capacity
	rb.count++
	rb.atomicCount.Add(1)

	// Notify waiting consumers
	select {
	case rb.notify <- struct{}{}:
	default:
	}

	return dropped, wasDropped
}

// Pop removes and returns the oldest message from the buffer.
// Returns (message, true) if a message was available, (zero, false) otherwise.
func (rb *ringBuffer) Pop() (Message, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	// A closed buffer can still be drained; closure only blocks new pushes.
	if rb.count == 0 {
		return Message{}, false
	}

	msg := rb.buf[rb.head]
	rb.buf[rb.head] = Message{} // Clear reference for GC
	rb.head = (rb.head + 1) % rb.capacity
	rb.count--
	rb.atomicCount.Add(-1)

	return msg, true
}

// Len returns the current number of messages in the buffer.
// Uses an atomic load so callers (e.g. Stats()) do not need to acquire the mutex.
func (rb *ringBuffer) Len() int {
	return int(rb.atomicCount.Load())
}

// Cap returns the capacity of the buffer.
func (rb *ringBuffer) Cap() int {
	return rb.capacity
}

// Close closes the buffer and wakes up any waiting consumers.
func (rb *ringBuffer) Close() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.closed {
		return
	}
	rb.closed = true
	close(rb.notify)
}

// IsClosed returns whether the buffer is closed.
func (rb *ringBuffer) IsClosed() bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.closed
}

// Notify returns the notification channel for consumers.
// A value is sent when new messages are available.
func (rb *ringBuffer) Notify() <-chan struct{} {
	return rb.notify
}

// Drain removes all messages from the buffer and returns them.
func (rb *ringBuffer) Drain() []Message {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		return nil
	}

	msgs := make([]Message, rb.count)
	for i := 0; i < rb.count; i++ {
		idx := (rb.head + i) % rb.capacity
		msgs[i] = rb.buf[idx]
		rb.buf[idx] = Message{} // Clear reference for GC
	}

	rb.head = 0
	rb.tail = 0
	rb.count = 0
	rb.atomicCount.Store(0)

	return msgs
}
