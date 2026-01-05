package pubsub

import "time"

// Message is an immutable event object by convention.
// After Publish, callers MUST NOT mutate Data/Meta referenced by this message.
type Message struct {
	ID    string
	Topic string
	Type  string
	Time  time.Time
	Data  any
	Meta  map[string]string
}

// BackpressurePolicy defines how to handle slow subscribers.
type BackpressurePolicy int

const (
	// DropOldest drops the oldest message when buffer is full
	DropOldest BackpressurePolicy = iota
	// DropNewest drops the newest message when buffer is full
	DropNewest
	// BlockWithTimeout blocks publish until timeout
	BlockWithTimeout
	// CloseSubscriber closes the subscription when buffer is full
	CloseSubscriber
)

// SubOptions configures a subscription.
type SubOptions struct {
	// BufferSize is the channel buffer size (default: 16)
	BufferSize int
	// Policy defines backpressure behavior
	Policy BackpressurePolicy
	// BlockTimeout is used when Policy == BlockWithTimeout (default: 50ms)
	BlockTimeout time.Duration
}

// Subscription represents a message subscription.
type Subscription interface {
	// C returns the receive-only message channel
	C() <-chan Message
	// Cancel unsubscribes and closes the channel
	Cancel()
}

// PubSub defines the publish-subscribe interface.
type PubSub interface {
	// Publish sends a message to a topic
	Publish(topic string, msg Message) error
	// Subscribe creates a new subscription to a topic
	Subscribe(topic string, opts SubOptions) (Subscription, error)
	// Close shuts down the pubsub system
	Close() error
}

// DefaultSubOptions returns production-ready default options.
func DefaultSubOptions() SubOptions {
	return SubOptions{
		BufferSize:   16,
		Policy:       DropOldest,
		BlockTimeout: 50 * time.Millisecond,
	}
}
