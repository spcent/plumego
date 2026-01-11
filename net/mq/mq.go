package mq

import (
	"context"
	"errors"
	"fmt"
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

// Publish sends a message to a topic.
func (b *InProcBroker) Publish(ctx context.Context, topic string, msg Message) (err error) {
	start := time.Now()
	panicked := false
	defer func() {
		if recovered := recover(); recovered != nil {
			panicked = true
			err = b.handlePanic(ctx, OpPublish, recovered)
		}
		b.observe(ctx, OpPublish, topic, start, err, panicked)
	}()

	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	if b == nil || b.ps == nil {
		return errors.New("mq: broker not initialized")
	}

	return b.ps.Publish(topic, msg)
}

// Subscribe registers a subscription for a topic.
func (b *InProcBroker) Subscribe(ctx context.Context, topic string, opts SubOptions) (sub Subscription, err error) {
	start := time.Now()
	panicked := false
	defer func() {
		if recovered := recover(); recovered != nil {
			panicked = true
			err = b.handlePanic(ctx, OpSubscribe, recovered)
		}
		b.observe(ctx, OpSubscribe, topic, start, err, panicked)
	}()

	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}

	if b == nil || b.ps == nil {
		return nil, errors.New("mq: broker not initialized")
	}

	return b.ps.Subscribe(topic, opts)
}

// Close shuts down the broker.
func (b *InProcBroker) Close() (err error) {
	start := time.Now()
	panicked := false
	defer func() {
		if recovered := recover(); recovered != nil {
			panicked = true
			err = b.handlePanic(context.Background(), OpClose, recovered)
		}
		b.observe(context.Background(), OpClose, "", start, err, panicked)
	}()

	if b == nil || b.ps == nil {
		return nil
	}
	return b.ps.Close()
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
