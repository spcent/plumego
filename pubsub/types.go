package pubsub

import "time"

// Message is treated as immutable once published.
// PubSub defensively copies Data/Meta on publish and per-delivery for common
// reference types (maps/slices) to avoid accidental mutation.
// Callers and subscribers should still treat Data/Meta as read-only.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	msg := pubsub.Message{
//		ID:    "msg-123",
//		Type:  "user.created",
//		Data: map[string]any{
//			"user_id": "123",
//			"email":   "user@example.com",
//		},
//		Meta: map[string]string{
//			"source": "api",
//		},
//	}
type Message struct {
	// ID is a unique identifier for the message
	ID string

	// Topic is the topic the message was published to
	Topic string

	// Type is the type of the message (optional)
	Type string

	// Time is the timestamp when the message was created
	Time time.Time

	// Data is the payload of the message
	Data any

	// Meta contains metadata about the message
	Meta map[string]string
}

// BackpressurePolicy defines how to handle slow subscribers.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	opts := pubsub.SubOptions{
//		BufferSize: 16,
//		Policy:     pubsub.DropOldest,
//	}
type BackpressurePolicy int

const (
	// DropOldest drops the oldest message when buffer is full.
	// Use this when you want to keep the most recent messages.
	DropOldest BackpressurePolicy = iota

	// DropNewest drops the newest message when buffer is full.
	// Use this when you want to keep the oldest messages.
	DropNewest

	// BlockWithTimeout blocks publish until timeout.
	// Use this when you want to wait for buffer space.
	BlockWithTimeout

	// CloseSubscriber closes the subscription when buffer is full.
	// Use this when you want to fail fast on slow subscribers.
	CloseSubscriber
)

// SubOptions configures a subscription.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	opts := pubsub.SubOptions{
//		BufferSize:   32,
//		Policy:       pubsub.BlockWithTimeout,
//		BlockTimeout: 100 * time.Millisecond,
//	}
type SubOptions struct {
	// BufferSize is the channel buffer size (default: 16)
	BufferSize int

	// Policy defines backpressure behavior
	Policy BackpressurePolicy

	// BlockTimeout is used when Policy == BlockWithTimeout (default: 50ms)
	BlockTimeout time.Duration
}

// Subscription represents a message subscription.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New()
//	defer ps.Close()
//
//	sub, err := ps.Subscribe("user.created", pubsub.DefaultSubOptions())
//	if err != nil {
//		// Handle error
//	}
//	defer sub.Cancel()
//
//	// Receive messages
//	for msg := range sub.C() {
//		// Process message
//	}
type Subscription interface {
	// C returns the receive-only message channel
	C() <-chan Message

	// Cancel unsubscribes and closes the channel
	Cancel()
}

// PubSub defines the publish-subscribe interface.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	var ps pubsub.PubSub = pubsub.New()
//	defer ps.Close()
//
//	// Publish a message
//	msg := pubsub.Message{Data: "hello"}
//	err := ps.Publish("topic", msg)
//
//	// Subscribe to a topic
//	sub, err := ps.Subscribe("topic", pubsub.DefaultSubOptions())
type PubSub interface {
	// Publish sends a message to a topic
	Publish(topic string, msg Message) error

	// Subscribe creates a new subscription to a topic
	Subscribe(topic string, opts SubOptions) (Subscription, error)

	// Close shuts down the pubsub system
	Close() error
}

// PatternPubSub extends PubSub with pattern subscriptions.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	var ps pubsub.PatternPubSub = pubsub.New()
//	defer ps.Close()
//
//	// Subscribe to all user events
//	sub, err := ps.SubscribePattern("user.*", pubsub.DefaultSubOptions())
type PatternPubSub interface {
	PubSub

	// SubscribePattern creates a new subscription to a topic pattern.
	SubscribePattern(pattern string, opts SubOptions) (Subscription, error)
}

// DefaultSubOptions returns production-ready default options.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	opts := pubsub.DefaultSubOptions()
func DefaultSubOptions() SubOptions {
	return SubOptions{
		BufferSize:   16,
		Policy:       DropOldest,
		BlockTimeout: 50 * time.Millisecond,
	}
}
