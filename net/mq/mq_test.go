package mq

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

	err := broker.Publish(context.Background(), "topic", Message{ID: "test-1"})
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

	if err := broker.Publish(context.Background(), "topic", Message{ID: "test-1"}); err != nil {
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

func TestInProcBrokerValidation(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	// Test empty topic
	err := broker.Publish(context.Background(), "  ", Message{ID: "test-1"})
	if !errors.Is(err, ErrInvalidTopic) {
		t.Fatalf("expected ErrInvalidTopic for empty topic, got %v", err)
	}

	// Test invalid message (missing ID)
	err = broker.Publish(context.Background(), "topic", Message{})
	if !errors.Is(err, ErrNilMessage) {
		t.Fatalf("expected ErrNilMessage for missing ID, got %v", err)
	}

	// Test nil broker
	var nilBroker *InProcBroker
	err = nilBroker.Publish(context.Background(), "topic", Message{ID: "test-1"})
	if !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("expected ErrNotInitialized for nil broker, got %v", err)
	}

	// Test empty topic for subscribe
	_, err = broker.Subscribe(context.Background(), "  ", SubOptions{})
	if !errors.Is(err, ErrInvalidTopic) {
		t.Fatalf("expected ErrInvalidTopic for empty topic in subscribe, got %v", err)
	}
}

func TestInProcBrokerConcurrent(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	// Subscribe with larger buffer
	sub, err := broker.Subscribe(context.Background(), "topic", SubOptions{BufferSize: 1000})
	if err != nil {
		t.Fatalf("subscribe error: %v", err)
	}
	defer sub.Cancel()

	// Concurrent publishers
	var wg sync.WaitGroup
	numPublishers := 5
	messagesPerPublisher := 50

	wg.Add(numPublishers)
	for i := 0; i < numPublishers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerPublisher; j++ {
				msg := Message{
					ID:   fmt.Sprintf("msg-%d-%d", id, j),
					Data: fmt.Sprintf("data-%d-%d", id, j),
				}
				if err := broker.Publish(context.Background(), "topic", msg); err != nil {
					t.Errorf("publish error: %v", err)
					return
				}
			}
		}(i)
	}

	// Wait for all publishers to finish
	wg.Wait()

	// Verify all messages were received
	expectedTotal := numPublishers * messagesPerPublisher
	received := 0
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-sub.C():
			received++
			if received == expectedTotal {
				return
			}
		case <-timeout:
			t.Fatalf("timeout waiting for messages, received %d out of %d", received, expectedTotal)
		}
	}
}

func TestInProcBrokerIdempotentClose(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())

	// First close
	if err := broker.Close(); err != nil {
		t.Fatalf("first close error: %v", err)
	}

	// Second close should be idempotent
	if err := broker.Close(); err != nil {
		t.Fatalf("second close should be idempotent, got error: %v", err)
	}

	// Third close should also be idempotent
	if err := broker.Close(); err != nil {
		t.Fatalf("third close should be idempotent, got error: %v", err)
	}
}

func TestInProcBrokerLongTopic(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	// Create a topic longer than 1024 characters
	longTopic := strings.Repeat("a", 1025)

	err := broker.Publish(context.Background(), longTopic, Message{ID: "test-1"})
	if !errors.Is(err, ErrInvalidTopic) {
		t.Fatalf("expected ErrInvalidTopic for long topic, got %v", err)
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
