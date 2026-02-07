package pubsub

import (
	"context"
	"time"
)

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

// String returns the string representation of a BackpressurePolicy.
func (p BackpressurePolicy) String() string {
	switch p {
	case DropOldest:
		return "drop_oldest"
	case DropNewest:
		return "drop_newest"
	case BlockWithTimeout:
		return "block_with_timeout"
	case CloseSubscriber:
		return "close_subscriber"
	default:
		return "unknown"
	}
}

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

	// ZeroCopy disables message cloning for this subscriber.
	// Only use this if the subscriber guarantees not to modify the message.
	// This can significantly improve performance for read-only subscribers.
	ZeroCopy bool

	// Filter is an optional function to filter messages before delivery.
	// If Filter returns false, the message is not delivered to this subscriber.
	// Filter is called under the subscriber's lock, so keep it fast.
	Filter func(msg Message) bool

	// UseRingBuffer enables ring buffer for this subscriber's DropOldest policy.
	// When enabled, uses an O(1) circular buffer instead of channel drain-and-retry.
	// Only effective when Policy is DropOldest.
	UseRingBuffer bool
}

// SubscriptionStats contains statistics for a subscription.
type SubscriptionStats struct {
	// Received is the total number of messages received
	Received uint64

	// Dropped is the total number of messages dropped due to backpressure
	Dropped uint64

	// QueueLen is the current number of messages in the buffer
	QueueLen int

	// QueueCap is the capacity of the buffer
	QueueCap int
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

	// ID returns the unique subscription ID
	ID() uint64

	// Topic returns the topic or pattern this subscription is for
	Topic() string

	// Stats returns subscription statistics
	Stats() SubscriptionStats

	// Done returns a channel that is closed when the subscription is cancelled
	Done() <-chan struct{}
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

// MQTTPubSub extends PubSub with MQTT-style pattern subscriptions.
//
// MQTT patterns use "/" as a level separator with two wildcards:
//   - "+" matches exactly one level (e.g., "devices/+/temp" matches "devices/sensor1/temp")
//   - "#" matches zero or more levels (must be last, e.g., "devices/#" matches "devices/sensor1/temp")
//
// Example:
//
//	var ps pubsub.MQTTPubSub = pubsub.New()
//	defer ps.Close()
//
//	// Subscribe to all sensor temperature readings
//	sub, err := ps.SubscribeMQTT("devices/+/temperature", pubsub.DefaultSubOptions())
//
//	// Subscribe to all events under a device
//	sub2, err := ps.SubscribeMQTT("devices/sensor1/#", pubsub.DefaultSubOptions())
type MQTTPubSub interface {
	PubSub

	// SubscribeMQTT creates a new subscription using MQTT-style topic patterns.
	SubscribeMQTT(pattern string, opts SubOptions) (Subscription, error)
}

// ContextPubSub extends PatternPubSub with context-aware operations.
type ContextPubSub interface {
	PatternPubSub

	// PublishWithContext publishes a message with context support.
	PublishWithContext(ctx context.Context, topic string, msg Message) error

	// SubscribeWithContext creates a subscription that is cancelled when the context is done.
	SubscribeWithContext(ctx context.Context, topic string, opts SubOptions) (Subscription, error)

	// SubscribePatternWithContext creates a pattern subscription that is cancelled when the context is done.
	SubscribePatternWithContext(ctx context.Context, pattern string, opts SubOptions) (Subscription, error)
}

// BatchPubSub extends ContextPubSub with batch operations.
type BatchPubSub interface {
	ContextPubSub

	// PublishBatch publishes multiple messages to a topic atomically.
	PublishBatch(topic string, msgs []Message) error

	// PublishMulti publishes messages to multiple topics.
	PublishMulti(msgs map[string][]Message) error
}

// DrainablePubSub extends BatchPubSub with graceful shutdown support.
type DrainablePubSub interface {
	BatchPubSub

	// Drain waits for all pending messages to be delivered or until the context is cancelled.
	Drain(ctx context.Context) error
}

// HistoryPubSub extends PubSub with per-topic message history retention.
// It allows querying previously published messages for late subscriber catch-up,
// debugging, and lightweight audit trails.
type HistoryPubSub interface {
	PubSub

	// GetTopicHistory returns all retained messages for a topic (oldest first).
	GetTopicHistory(topic string) ([]Message, error)

	// GetTopicHistorySince returns messages added after the given sequence number.
	GetTopicHistorySince(topic string, sequence uint64) ([]Message, error)

	// GetRecentMessages returns the last N messages for a topic (oldest first).
	GetRecentMessages(topic string, count int) ([]Message, error)

	// GetTopicHistoryByTTL returns messages not older than the given TTL duration.
	GetTopicHistoryByTTL(topic string, ttl time.Duration) ([]Message, error)

	// ClearTopicHistory removes all retained messages for a topic.
	ClearTopicHistory(topic string) error

	// TopicHistoryStats returns history statistics for all topics.
	TopicHistoryStats() (map[string]HistoryStats, error)

	// TopicHistorySequence returns the current sequence number for a topic.
	TopicHistorySequence(topic string) (uint64, error)
}

// RequestReplyPubSub extends PubSub with request-reply pattern support.
//
// Example:
//
//	import "github.com/spcent/plumego/pubsub"
//
//	ps := pubsub.New(pubsub.WithRequestReply())
//	defer ps.Close()
//
//	// Responder
//	sub, _ := ps.Subscribe("math.add", pubsub.DefaultSubOptions())
//	go func() {
//		for msg := range sub.C() {
//			if pubsub.IsRequest(msg) {
//				ps.Reply(msg, pubsub.Message{Data: "result"})
//			}
//		}
//	}()
//
//	// Requester
//	resp, err := ps.RequestWithTimeout("math.add", pubsub.Message{Data: "2+3"}, time.Second)
type RequestReplyPubSub interface {
	PubSub

	// Request sends a message and waits for a response.
	Request(ctx context.Context, topic string, msg Message) (Message, error)

	// RequestWithTimeout sends a message and waits for a response with timeout.
	RequestWithTimeout(topic string, msg Message, timeout time.Duration) (Message, error)

	// Reply sends a reply to a request message.
	Reply(reqMsg Message, respMsg Message) error
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
