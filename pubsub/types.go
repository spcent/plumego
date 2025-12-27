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

type BackpressurePolicy int

const (
	DropOldest BackpressurePolicy = iota
	DropNewest
	BlockWithTimeout
	CloseSubscriber
)

type SubOptions struct {
	BufferSize int
	Policy     BackpressurePolicy

	// Only used when Policy == BlockWithTimeout.
	// If zero, a safe default (50ms) is used.
	BlockTimeout time.Duration
}

type Subscription interface {
	C() <-chan Message
	Cancel()
}

type PubSub interface {
	Publish(topic string, msg Message) error
	Subscribe(topic string, opts SubOptions) (Subscription, error)
	Close() error
}
