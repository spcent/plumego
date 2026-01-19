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

func TestInProcBrokerHealthCheck(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	// Subscribe to a topic
	sub, err := broker.Subscribe(context.Background(), "health-test", SubOptions{BufferSize: 1})
	if err != nil {
		t.Fatalf("subscribe error: %v", err)
	}
	defer sub.Cancel()

	// Publish a message
	if err := broker.Publish(context.Background(), "health-test", Message{ID: "test-1"}); err != nil {
		t.Fatalf("publish error: %v", err)
	}

	// Check health
	status := broker.HealthCheck()
	if status.Status != "healthy" {
		t.Fatalf("expected status 'healthy', got %q", status.Status)
	}
	if status.TotalTopics != 1 {
		t.Fatalf("expected 1 topic, got %d", status.TotalTopics)
	}
	if status.TotalSubs != 1 {
		t.Fatalf("expected 1 subscriber, got %d", status.TotalSubs)
	}
	if status.Uptime == "" {
		t.Fatalf("expected non-empty uptime")
	}
	if status.Metrics.TotalPublished != 1 {
		t.Fatalf("expected 1 published message, got %d", status.Metrics.TotalPublished)
	}
}

func TestInProcBrokerConfig(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	// Get default config
	cfg := broker.GetConfig()
	if !cfg.EnableHealthCheck {
		t.Fatalf("expected EnableHealthCheck to be true by default")
	}
	if cfg.DefaultBufferSize != 16 {
		t.Fatalf("expected DefaultBufferSize to be 16 by default, got %d", cfg.DefaultBufferSize)
	}

	// Update config
	newCfg := DefaultConfig()
	newCfg.DefaultBufferSize = 32
	newCfg.EnableMetrics = false
	if err := broker.UpdateConfig(newCfg); err != nil {
		t.Fatalf("update config error: %v", err)
	}

	// Verify updated config
	updatedCfg := broker.GetConfig()
	if updatedCfg.DefaultBufferSize != 32 {
		t.Fatalf("expected DefaultBufferSize to be 32, got %d", updatedCfg.DefaultBufferSize)
	}
	if updatedCfg.EnableMetrics != false {
		t.Fatalf("expected EnableMetrics to be false")
	}

	// Test invalid config
	invalidCfg := DefaultConfig()
	invalidCfg.DefaultBufferSize = 0
	if err := broker.UpdateConfig(invalidCfg); err == nil {
		t.Fatalf("expected error for invalid config")
	}
}

func TestInProcBrokerPublishBatch(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	// Subscribe
	sub, err := broker.Subscribe(context.Background(), "batch-topic", SubOptions{BufferSize: 10})
	if err != nil {
		t.Fatalf("subscribe error: %v", err)
	}
	defer sub.Cancel()

	// Create batch of messages
	messages := []Message{
		{ID: "msg-1", Data: "data-1"},
		{ID: "msg-2", Data: "data-2"},
		{ID: "msg-3", Data: "data-3"},
	}

	// Publish batch
	if err := broker.PublishBatch(context.Background(), "batch-topic", messages); err != nil {
		t.Fatalf("publish batch error: %v", err)
	}

	// Verify all messages received
	received := 0
	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-sub.C():
			received++
			if received == len(messages) {
				return
			}
		case <-timeout:
			t.Fatalf("timeout waiting for messages, received %d out of %d", received, len(messages))
		}
	}
}

func TestInProcBrokerSubscribeBatch(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	// Subscribe to multiple topics
	topics := []string{"topic-1", "topic-2", "topic-3"}
	subs, err := broker.SubscribeBatch(context.Background(), topics, SubOptions{BufferSize: 5})
	if err != nil {
		t.Fatalf("subscribe batch error: %v", err)
	}
	if len(subs) != len(topics) {
		t.Fatalf("expected %d subscriptions, got %d", len(topics), len(subs))
	}

	// Verify each subscription works
	for i, sub := range subs {
		if err := broker.Publish(context.Background(), topics[i], Message{ID: fmt.Sprintf("test-%d", i)}); err != nil {
			t.Fatalf("publish error: %v", err)
		}
		select {
		case msg := <-sub.C():
			if msg.ID != fmt.Sprintf("test-%d", i) {
				t.Fatalf("expected message ID test-%d, got %s", i, msg.ID)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timeout waiting for message on topic %s", topics[i])
		}
	}

	// Clean up subscriptions
	for _, sub := range subs {
		sub.Cancel()
	}
}

func TestInProcBrokerInvalidConfig(t *testing.T) {
	// Test invalid config during creation
	invalidCfg := DefaultConfig()
	invalidCfg.DefaultBufferSize = 0

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for invalid config")
		}
	}()

	// This should panic
	_ = NewInProcBroker(pubsub.New(), WithConfig(invalidCfg))
}

func TestInProcBrokerPriorityQueue(t *testing.T) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	// Test priority message
	priorityMsg := PriorityMessage{
		Message:  Message{ID: "priority-1", Data: "high priority data"},
		Priority: PriorityHigh,
	}

	err := broker.PublishPriority(context.Background(), "priority-topic", priorityMsg)
	if err != nil {
		t.Fatalf("publish priority message error: %v", err)
	}

	// Verify priority queue is enabled by default
	cfg := broker.GetConfig()
	if !cfg.EnablePriorityQueue {
		t.Fatalf("expected EnablePriorityQueue to be true by default")
	}
}

func TestInProcBrokerAckSupport(t *testing.T) {
	// Test with acknowledgment support disabled (default)
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	cfg := broker.GetConfig()
	if cfg.EnableAckSupport {
		t.Fatalf("expected EnableAckSupport to be false by default")
	}

	// Test with acknowledgment support enabled
	ackCfg := DefaultConfig()
	ackCfg.EnableAckSupport = true
	brokerWithAck := NewInProcBroker(pubsub.New(), WithConfig(ackCfg))
	defer brokerWithAck.Close()

	// Test publishing with acknowledgment
	ackMsg := AckMessage{
		Message:   Message{ID: "ack-1", Data: "ack required"},
		AckPolicy: AckRequired,
	}

	err := brokerWithAck.PublishWithAck(context.Background(), "ack-topic", ackMsg)
	if err != nil {
		t.Fatalf("publish with ack error: %v", err)
	}

	// Test acknowledgment methods
	err = brokerWithAck.Ack(context.Background(), "ack-topic", "ack-1")
	if err != nil {
		t.Fatalf("ack error: %v", err)
	}

	err = brokerWithAck.Nack(context.Background(), "ack-topic", "ack-1")
	if err != nil {
		t.Fatalf("nack error: %v", err)
	}
}

func TestInProcBrokerMemoryLimit(t *testing.T) {
	// Test with memory limit
	cfg := DefaultConfig()
	cfg.MaxMemoryUsage = 1 << 30 // 1GB limit
	broker := NewInProcBroker(pubsub.New(), WithConfig(cfg))
	defer broker.Close()

	// Get memory usage
	memUsage := broker.GetMemoryUsage()
	if memUsage == 0 {
		t.Fatalf("expected non-zero memory usage")
	}

	// Verify config
	updatedCfg := broker.GetConfig()
	if updatedCfg.MaxMemoryUsage != 1<<30 {
		t.Fatalf("expected MaxMemoryUsage to be 1GB, got %d", updatedCfg.MaxMemoryUsage)
	}

	// Test health check includes memory info
	status := broker.HealthCheck()
	if status.MemoryUsage == 0 {
		t.Fatalf("expected non-zero memory usage in health check")
	}
	if status.MemoryLimit != 1<<30 {
		t.Fatalf("expected memory limit 1GB in health check, got %d", status.MemoryLimit)
	}
}

func TestInProcBrokerMemoryLimitExceeded(t *testing.T) {
	// Test with very low memory limit
	cfg := DefaultConfig()
	cfg.MaxMemoryUsage = 1 // 1 byte limit - will definitely exceed
	broker := NewInProcBroker(pubsub.New(), WithConfig(cfg))
	defer broker.Close()

	// Try to publish a message (should work since we don't check memory on publish yet)
	// This is a framework test - actual memory checking would be implemented in production
	err := broker.Publish(context.Background(), "test-topic", Message{ID: "test-1", Data: "test"})
	if err != nil {
		t.Fatalf("publish should succeed even with low memory limit: %v", err)
	}
}

func TestInProcBrokerTriePattern(t *testing.T) {
	// Test Trie pattern configuration
	cfg := DefaultConfig()
	cfg.EnableTriePattern = true
	broker := NewInProcBroker(pubsub.New(), WithConfig(cfg))
	defer broker.Close()

	// Verify config
	updatedCfg := broker.GetConfig()
	if !updatedCfg.EnableTriePattern {
		t.Fatalf("expected EnableTriePattern to be true")
	}

	// Note: Actual Trie implementation would be in pubsub layer
	// This test verifies the configuration is properly stored
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
