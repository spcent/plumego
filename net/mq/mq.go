package mq

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/pubsub"
)

// Message re-exports pubsub.Message for broker compatibility.
type Message = pubsub.Message

// SubOptions re-exports pubsub.SubOptions for broker compatibility.
type SubOptions = pubsub.SubOptions

// Subscription re-exports pubsub.Subscription for broker compatibility.
type Subscription = pubsub.Subscription

// Operation labels broker actions for observability.
type Operation string

const (
	OpPublish   Operation = "publish"
	OpSubscribe Operation = "subscribe"
	OpClose     Operation = "close"
	OpMetrics   Operation = "metrics"
)

// Metrics captures timing and error information for broker actions.
type Metrics struct {
	Operation Operation
	Topic     string
	Duration  time.Duration
	Err       error
	Panic     bool
}

// MetricsCollector can be plugged into the broker to observe activity.
// This is now an alias for the unified metrics collector
type MetricsCollector = metrics.MetricsCollector

// PanicHandler is invoked when a broker operation panics.
type PanicHandler func(ctx context.Context, op Operation, recovered any)

// ErrRecoveredPanic is returned when a broker operation recovers from panic.
var ErrRecoveredPanic = errors.New("mq: panic recovered")

// ErrNotInitialized is returned when the broker is not properly initialized.
var ErrNotInitialized = errors.New("mq: broker not initialized")

// ErrInvalidTopic is returned when a topic is invalid.
var ErrInvalidTopic = errors.New("mq: invalid topic")

// ErrNilMessage is returned when attempting to publish a nil message.
var ErrNilMessage = errors.New("mq: message cannot be nil")

// Broker defines the interface for message queue backends.
type Broker interface {
	Publish(ctx context.Context, topic string, msg Message) error
	Subscribe(ctx context.Context, topic string, opts SubOptions) (Subscription, error)
	Close() error
}

// InProcBroker adapts pubsub.PubSub to the Broker interface.
type InProcBroker struct {
	ps           pubsub.PubSub
	metrics      MetricsCollector
	panicHandler PanicHandler
}

// Option configures the broker.
type Option func(*InProcBroker)

// WithMetricsCollector registers a metrics collector.
func WithMetricsCollector(collector MetricsCollector) Option {
	return func(b *InProcBroker) {
		b.metrics = collector
	}
}

// WithPanicHandler registers a panic handler.
func WithPanicHandler(handler PanicHandler) Option {
	return func(b *InProcBroker) {
		b.panicHandler = handler
	}
}

// validateTopic checks if a topic is valid.
func validateTopic(topic string) error {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return fmt.Errorf("%w: cannot be empty", ErrInvalidTopic)
	}
	if len(topic) > 1024 {
		return fmt.Errorf("%w: topic too long (max 1024 characters)", ErrInvalidTopic)
	}
	return nil
}

// validateMessage checks if a message is valid.
func validateMessage(msg Message) error {
	if msg.ID == "" {
		return fmt.Errorf("%w: ID is required", ErrNilMessage)
	}
	return nil
}

// NewInProcBroker wraps the in-process pubsub implementation.
func NewInProcBroker(ps pubsub.PubSub, opts ...Option) *InProcBroker {
	if ps == nil {
		ps = pubsub.New()
	}
	broker := &InProcBroker{ps: ps}
	for _, opt := range opts {
		if opt != nil {
			opt(broker)
		}
	}
	return broker
}

// executeWithObservability wraps an operation with observability logic.
func (b *InProcBroker) executeWithObservability(
	ctx context.Context,
	op Operation,
	topic string,
	fn func() error,
) (err error) {
	start := time.Now()
	panicked := false
	defer func() {
		if recovered := recover(); recovered != nil {
			panicked = true
			err = b.handlePanic(ctx, op, recovered)
		}
		b.observe(ctx, op, topic, start, err, panicked)
	}()
	return fn()
}

// Publish sends a message to a topic.
func (b *InProcBroker) Publish(ctx context.Context, topic string, msg Message) error {
	return b.executeWithObservability(ctx, OpPublish, topic, func() error {
		// Validate context
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		// Validate broker initialization
		if b == nil || b.ps == nil {
			return fmt.Errorf("%w", ErrNotInitialized)
		}

		// Validate topic
		if err := validateTopic(topic); err != nil {
			return err
		}

		// Validate message
		if err := validateMessage(msg); err != nil {
			return err
		}

		return b.ps.Publish(topic, msg)
	})
}

// Subscribe registers a subscription for a topic.
func (b *InProcBroker) Subscribe(ctx context.Context, topic string, opts SubOptions) (sub Subscription, err error) {
	err = b.executeWithObservability(ctx, OpSubscribe, topic, func() error {
		// Validate context
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		// Validate broker initialization
		if b == nil || b.ps == nil {
			return fmt.Errorf("%w", ErrNotInitialized)
		}

		// Validate topic
		if err := validateTopic(topic); err != nil {
			return err
		}

		// Subscribe
		var subscription Subscription
		subscription, err = b.ps.Subscribe(topic, opts)
		if err != nil {
			return err
		}

		// Store subscription for return
		sub = subscription
		return nil
	})

	return sub, err
}

// Close shuts down the broker.
func (b *InProcBroker) Close() error {
	return b.executeWithObservability(context.Background(), OpClose, "", func() error {
		// Validate broker initialization
		if b == nil || b.ps == nil {
			return nil // Close is idempotent
		}
		return b.ps.Close()
	})
}

// DefaultSubOptions exposes the default subscription settings.
var DefaultSubOptions = pubsub.DefaultSubOptions

// Snapshot exposes in-process pubsub metrics when supported.
func (b *InProcBroker) Snapshot() pubsub.MetricsSnapshot {
	if b == nil || b.ps == nil {
		return pubsub.MetricsSnapshot{}
	}
	if snapper, ok := b.ps.(interface{ Snapshot() pubsub.MetricsSnapshot }); ok {
		return snapper.Snapshot()
	}
	return pubsub.MetricsSnapshot{}
}

func (b *InProcBroker) handlePanic(ctx context.Context, op Operation, recovered any) error {
	if b != nil && b.panicHandler != nil {
		func() {
			defer func() {
				_ = recover()
			}()
			b.panicHandler(ctx, op, recovered)
		}()
	}
	return fmt.Errorf("%w: %s: %v", ErrRecoveredPanic, op, recovered)
}

func (b *InProcBroker) observe(ctx context.Context, op Operation, topic string, start time.Time, err error, panicked bool) {
	if b == nil || b.metrics == nil {
		return
	}

	duration := time.Since(start)

	func() {
		defer func() {
			if recovered := recover(); recovered != nil && b.panicHandler != nil {
				b.panicHandler(ctx, OpMetrics, recovered)
			}
		}()
		// Use the unified interface
		b.metrics.ObserveMQ(ctx, string(op), topic, duration, err, panicked)
	}()
}
