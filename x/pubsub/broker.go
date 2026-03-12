package pubsub

import "context"

// Broker is the core publish-subscribe interface.
// It is intentionally minimal and stable — callers depend on this interface,
// not on concrete types or extended interface chains.
//
// Example:
//
//	var b pubsub.Broker = pubsub.New()
//	defer b.Close()
//
//	sub, err := b.Subscribe("user.created", pubsub.DefaultSubOptions())
//	if err != nil {
//		// handle error
//	}
//	defer sub.Cancel()
//
//	err = b.Publish("user.created", pubsub.Message{Data: "hello"})
type Broker interface {
	// Publish sends a message to a topic.
	Publish(topic string, msg Message) error

	// Subscribe creates a new subscription to an exact topic.
	Subscribe(topic string, opts SubOptions) (Subscription, error)

	// Close shuts down the broker and cancels all subscriptions.
	Close() error
}

// PatternSubscriber extends Broker with glob-pattern subscriptions.
//
// Pattern syntax follows filepath.Match:
//   - "*" matches any sequence of non-separator characters
//   - "?" matches any single non-separator character
//   - "[abc]" matches character class
//
// Example:
//
//	var b pubsub.PatternSubscriber = pubsub.New()
//	sub, err := b.SubscribePattern("user.*", pubsub.DefaultSubOptions())
type PatternSubscriber interface {
	Broker
	// SubscribePattern creates a subscription matching a glob pattern.
	SubscribePattern(pattern string, opts SubOptions) (Subscription, error)
}

// MQTTSubscriber extends Broker with MQTT-style topic pattern subscriptions.
//
// MQTT pattern rules:
//   - "/" is the level separator
//   - "+" matches exactly one level
//   - "#" matches zero or more levels and must appear last
//
// Example:
//
//	var b pubsub.MQTTSubscriber = pubsub.New()
//	sub, err := b.SubscribeMQTT("devices/+/temperature", pubsub.DefaultSubOptions())
type MQTTSubscriber interface {
	Broker
	// SubscribeMQTT creates a subscription using MQTT-style topic patterns.
	SubscribeMQTT(pattern string, opts SubOptions) (Subscription, error)
}

// BatchPublisher extends Broker with batch-publish operations.
type BatchPublisher interface {
	Broker
	// PublishBatch publishes multiple messages to a single topic atomically.
	PublishBatch(topic string, msgs []Message) error

	// PublishMulti publishes messages to multiple topics.
	PublishMulti(msgs map[string][]Message) error
}

// Drainable extends Broker with graceful-shutdown support.
type Drainable interface {
	Broker
	// Drain blocks until all in-flight messages are delivered or ctx is cancelled.
	Drain(ctx context.Context) error
}

// BrokerObserver receives lifecycle notifications from a Broker.
// Register via WithObserver when creating the broker.
//
// All methods are called synchronously; implementations must be non-blocking.
type BrokerObserver interface {
	OnPublish(topic string, msg *Message)
	OnSubscribe(topic string, subID uint64)
	OnUnsubscribe(topic string, subID uint64)
	OnDeliver(topic string, subID uint64, msg *Message)
	OnDrop(topic string, subID uint64, msg *Message, policy BackpressurePolicy)
}
