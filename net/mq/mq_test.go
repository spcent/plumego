package mq

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/pubsub"
)

func TestInProcBrokerPublishSubscribe(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	sub, err := broker.Subscribe(context.Background(), "topic", SubOptions{BufferSize: 1})
	if err != nil {
		t.Fatalf("subscribe error: %v", err)
	}
	defer sub.Cancel()

	msg := Message{ID: "1", Data: "payload"}
	if err := broker.Publish(context.Background(), "topic", msg); err != nil {
		t.Fatalf("publish error: %v", err)
	}

	select {
	case got := <-sub.C():
		if got.ID != "1" {
			t.Fatalf("expected message ID 1, got %q", got.ID)
		}
		if got.Topic != "topic" {
			t.Fatalf("expected topic 'topic', got %q", got.Topic)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected message to be delivered")
	}
}

func TestInProcBrokerContextCancel(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := broker.Publish(ctx, "topic", Message{}); err == nil {
		t.Fatalf("expected publish to fail with canceled context")
	}
}

func TestInProcBrokerPanicRecovery(t *testing.T) {
	called := false
	broker := NewInProcBroker(panicPubSub{}, WithPanicHandler(func(ctx context.Context, op Operation, recovered any) {
		called = true
		if op != OpPublish {
			t.Fatalf("expected panic handler op publish, got %s", op)
		}
	}))

	err := broker.Publish(context.Background(), "topic", Message{})
	if !errors.Is(err, ErrRecoveredPanic) {
		t.Fatalf("expected ErrRecoveredPanic, got %v", err)
	}
	if !called {
		t.Fatalf("expected panic handler to be called")
	}
}

func TestInProcBrokerMetrics(t *testing.T) {
	collector := &metricsCollector{}
	broker := NewInProcBroker(pubsub.New(), WithMetricsCollector(collector))
	defer broker.Close()

	if err := broker.Publish(context.Background(), "topic", Message{}); err != nil {
		t.Fatalf("unexpected publish error: %v", err)
	}

	collector.mu.Lock()
	defer collector.mu.Unlock()
	if collector.count == 0 {
		t.Fatalf("expected metrics to be recorded")
	}
	if collector.last.Operation != OpPublish {
		t.Fatalf("expected publish operation, got %s", collector.last.Operation)
	}
	if collector.last.Topic != "topic" {
		t.Fatalf("expected topic 'topic', got %q", collector.last.Topic)
	}
	if collector.last.Panic {
		t.Fatalf("did not expect panic flag")
	}
}

type panicPubSub struct{}

func (p panicPubSub) Publish(topic string, msg pubsub.Message) error {
	panic("boom")
}

func (p panicPubSub) Subscribe(topic string, opts pubsub.SubOptions) (pubsub.Subscription, error) {
	return nil, errors.New("not implemented")
}

func (p panicPubSub) Close() error {
	return nil
}

type metricsCollector struct {
	mu    sync.Mutex
	last  Metrics
	count int
}

func (m *metricsCollector) Record(ctx context.Context, record metrics.MetricRecord) {
	// Not used in this test
}

func (m *metricsCollector) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	// Not used in this test
}

func (m *metricsCollector) ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	// Not used in this test
}

func (m *metricsCollector) ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.last = Metrics{
		Operation: Operation(operation),
		Topic:     topic,
		Duration:  duration,
		Err:       err,
		Panic:     panicked,
	}
	m.count++
}

func (m *metricsCollector) ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool) {
	// Not used in this test
}

func (m *metricsCollector) GetStats() metrics.CollectorStats {
	return metrics.CollectorStats{}
}

func (m *metricsCollector) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.count = 0
	m.last = Metrics{}
}
