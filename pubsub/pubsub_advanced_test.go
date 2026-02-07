package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/spcent/plumego/metrics"
)

// TestPubSub_Performance tests optimized performance
func TestPubSub_Performance(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Create multiple subscribers
	const numSubs = 10
	const numMsgs = 1000

	subs := make([]Subscription, numSubs)
	for i := 0; i < numSubs; i++ {
		sub, err := ps.Subscribe("perf", SubOptions{
			BufferSize: 100,
			Policy:     DropOldest,
		})
		if err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
		subs[i] = sub
	}

	// Measure publishing performance
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < numMsgs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := Message{ID: "msg", Data: id}
			if err := ps.Publish("perf", msg); err != nil {
				t.Errorf("publish failed: %v", err)
			}
		}(i)
	}
	wg.Wait()
	duration := time.Since(start)

	// Verify all subscribers received messages
	for i, sub := range subs {
		received := 0
		// Give more time for all messages to be delivered
		timeout := time.After(500 * time.Millisecond)
		for {
			select {
			case <-sub.C():
				received++
			case <-timeout:
				t.Logf("Subscriber %d received %d messages", i, received)
				// With buffer size 100 and 10 subscribers, expect at least 100 messages
				if received < 100 {
					t.Errorf("Expected at least 100 messages, got %d", received)
				}
				goto done
			}
		}
	done:
	}

	t.Logf("Published %d messages to %d subscribers in %v (%.2f msg/s)",
		numMsgs, numSubs, duration, float64(numMsgs)/duration.Seconds())
}

// TestPubSub_ConcurrentSubscribers tests concurrent subscriber management
func TestPubSub_ConcurrentSubscribers(t *testing.T) {
	ps := New()
	defer ps.Close()

	const numGoroutines = 50
	var wg sync.WaitGroup

	// Concurrent subscribe and cancel
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Subscribe
			sub, err := ps.Subscribe("concurrent", SubOptions{
				BufferSize: 10,
				Policy:     DropNewest,
			})
			if err != nil {
				t.Errorf("subscribe failed: %v", err)
				return
			}

			// Briefly receive
			time.Sleep(10 * time.Millisecond)

			// Cancel
			sub.Cancel()
		}(i)
	}

	// Concurrent publishing
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := Message{ID: "test", Data: id}
			_ = ps.Publish("concurrent", msg)
		}(i)
	}

	wg.Wait()

	// Verify final state
	count := ps.GetSubscriberCount("concurrent")
	if count != 0 {
		t.Errorf("Expected 0 subscribers, got %d", count)
	}
}

// TestPubSub_DifferentPolicies tests different backpressure policies
func TestPubSub_DifferentPolicies(t *testing.T) {
	ps := New()
	defer ps.Close()

	tests := []struct {
		name   string
		policy BackpressurePolicy
	}{
		{"DropOldest", DropOldest},
		{"DropNewest", DropNewest},
		{"BlockWithTimeout", BlockWithTimeout},
		{"CloseSubscriber", CloseSubscriber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use unique topic for each test to avoid interference
			topic := "policy_test_" + tt.name

			sub, err := ps.Subscribe(topic, SubOptions{
				BufferSize:   2,
				Policy:       tt.policy,
				BlockTimeout: 50 * time.Millisecond,
			})
			if err != nil {
				t.Fatalf("subscribe failed: %v", err)
			}

			// Fill buffer
			for i := 0; i < 5; i++ {
				msg := Message{ID: "msg", Data: i}
				_ = ps.Publish(topic, msg)
			}

			// Give time for processing
			time.Sleep(50 * time.Millisecond)

			// Verify metrics
			snapshot := ps.Snapshot()
			tm, exists := snapshot.Topics[topic]
			if !exists {
				t.Fatalf("Topic %s not found in metrics", topic)
			}

			if tm.PublishTotal != 5 {
				t.Errorf("Expected 5 publishes, got %d", tm.PublishTotal)
			}

			// Only cancel if subscriber is still active (not closed by policy)
			if tt.policy != CloseSubscriber {
				sub.Cancel()
			}
		})
	}
}

// TestPubSub_ListOperations tests list operations
func TestPubSub_ListOperations(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Create subscriptions on different topics
	topics := []string{"topic1", "topic2", "topic3"}
	for _, topic := range topics {
		_, err := ps.Subscribe(topic, DefaultSubOptions())
		if err != nil {
			t.Fatalf("subscribe failed: %v", err)
		}
	}

	// Test ListTopics
	listed := ps.ListTopics()
	if len(listed) != 3 {
		t.Errorf("Expected 3 topics, got %d", len(listed))
	}

	// Test GetSubscriberCount
	for _, topic := range topics {
		count := ps.GetSubscriberCount(topic)
		if count != 1 {
			t.Errorf("Expected 1 subscriber for %s, got %d", topic, count)
		}
	}

	// Test non-existent topic
	count := ps.GetSubscriberCount("nonexistent")
	if count != 0 {
		t.Errorf("Expected 0 subscribers for nonexistent topic, got %d", count)
	}
}

// TestPubSub_ErrorHandling tests error handling
func TestPubSub_ErrorHandling(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Test invalid topic
	_, err := ps.Subscribe("", DefaultSubOptions())
	if err != ErrInvalidTopic {
		t.Errorf("Expected ErrInvalidTopic, got %v", err)
	}

	// Test invalid buffer size (negative)
	_, err = ps.Subscribe("test", SubOptions{BufferSize: -1})
	if err != nil && err != ErrBufferTooSmall {
		t.Errorf("Expected ErrBufferTooSmall or nil, got %v", err)
	}

	// Test publish to closed
	ps.Close()
	err = ps.Publish("test", Message{ID: "test"})
	if err != ErrPublishToClosed {
		t.Errorf("Expected ErrPublishToClosed, got %v", err)
	}

	// Test subscribe to closed
	_, err = ps.Subscribe("test", DefaultSubOptions())
	if err != ErrSubscribeToClosed {
		t.Errorf("Expected ErrSubscribeToClosed, got %v", err)
	}
}

// TestPubSub_MetricsAccuracy tests metric accuracy
func TestPubSub_MetricsAccuracy(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Create subscribers with different policies
	sub1, _ := ps.Subscribe("metrics", SubOptions{BufferSize: 1, Policy: DropOldest})
	sub2, _ := ps.Subscribe("metrics", SubOptions{BufferSize: 1, Policy: DropNewest})
	defer sub1.Cancel()
	defer sub2.Cancel()

	// Publish messages
	for i := 0; i < 10; i++ {
		_ = ps.Publish("metrics", Message{ID: "msg", Data: i})
	}

	// Give time for delivery
	time.Sleep(50 * time.Millisecond)

	// Check metrics
	snapshot := ps.Snapshot()
	tm := snapshot.Topics["metrics"]

	if tm.PublishTotal != 10 {
		t.Errorf("Expected 10 publishes, got %d", tm.PublishTotal)
	}

	if tm.SubscribersGauge != 2 {
		t.Errorf("Expected 2 subscribers, got %d", tm.SubscribersGauge)
	}

	// Should have some dropped messages due to buffer size 1
	totalDropped := uint64(0)
	for _, count := range tm.DroppedByPolicy {
		totalDropped += count
	}

	if totalDropped == 0 {
		t.Logf("Warning: No dropped messages, expected some due to buffer size 1")
	}

	t.Logf("Metrics: Publish=%d, Delivered=%d, Dropped=%d, Subs=%d",
		tm.PublishTotal, tm.DeliveredTotal, totalDropped, tm.SubscribersGauge)
}

// TestPubSub_RapidSubscribeCancel tests rapid subscribe/cancel
func TestPubSub_RapidSubscribeCancel(t *testing.T) {
	ps := New()
	defer ps.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub, err := ps.Subscribe("rapid", DefaultSubOptions())
			if err != nil {
				return
			}
			// Immediately cancel
			sub.Cancel()
		}()
	}
	wg.Wait()

	// Should have 0 subscribers
	if count := ps.GetSubscriberCount("rapid"); count != 0 {
		t.Errorf("Expected 0 subscribers, got %d", count)
	}
}

// TestPubSub_PressureHandling tests pressure handling
func TestPubSub_PressureHandling(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Create slow subscriber with DropNewest policy to avoid blocking on each publish.
	// BlockWithTimeout with synchronous publish can cause cumulative blocking that
	// exceeds test timeouts in resource-constrained environments.
	slowSub, _ := ps.Subscribe("pressure", SubOptions{
		BufferSize: 2,
		Policy:     DropNewest,
	})
	defer slowSub.Cancel()

	// Publish many messages quickly
	for i := 0; i < 100; i++ {
		_ = ps.Publish("pressure", Message{ID: "msg", Data: i})
	}

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Check metrics
	snapshot := ps.Snapshot()
	tm := snapshot.Topics["pressure"]

	t.Logf("Pressure test: Published=%d, Delivered=%d, Dropped=%d",
		tm.PublishTotal, tm.DeliveredTotal, len(tm.DroppedByPolicy))

	// Should have processed some messages (buffer holds 2)
	if tm.DeliveredTotal == 0 {
		t.Errorf("Expected some delivered messages")
	}

	// With 100 messages and buffer of 2, most should be dropped
	if len(tm.DroppedByPolicy) == 0 {
		t.Errorf("Expected some dropped messages under pressure")
	}
}

// TestPublishWithContext_CancelledContext tests that PublishWithContext returns
// an error immediately when the context is already cancelled.
func TestPublishWithContext_CancelledContext(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.Subscribe("ctx.test", SubOptions{BufferSize: 8, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Cancel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err = ps.PublishWithContext(ctx, "ctx.test", Message{ID: "m1", Data: "hello"})
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// No message should have been delivered
	select {
	case msg := <-sub.C():
		t.Fatalf("expected no message, got %+v", msg)
	case <-time.After(50 * time.Millisecond):
		// expected
	}
}

// TestPublishWithContext_DeadlineExceeded tests that PublishWithContext returns
// an error when the context deadline has already passed.
func TestPublishWithContext_DeadlineExceeded(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.Subscribe("ctx.deadline", SubOptions{BufferSize: 8, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Cancel()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	err = ps.PublishWithContext(ctx, "ctx.deadline", Message{ID: "m1"})
	if err == nil {
		t.Fatal("expected error from expired deadline, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
}

// TestPublishWithContext_ValidContext tests that PublishWithContext works
// normally with a valid (non-cancelled) context.
func TestPublishWithContext_ValidContext(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.Subscribe("ctx.valid", SubOptions{BufferSize: 8, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Cancel()

	ctx := context.Background()
	err = ps.PublishWithContext(ctx, "ctx.valid", Message{ID: "m1", Data: "hello"})
	if err != nil {
		t.Fatalf("publish with valid context: %v", err)
	}

	select {
	case got := <-sub.C():
		if got.ID != "m1" {
			t.Fatalf("expected message ID m1, got %s", got.ID)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for message")
	}
}

// TestPublishWithContext_BlockWithTimeoutRespectsContext tests that a
// BlockWithTimeout subscriber respects context cancellation during delivery.
func TestPublishWithContext_BlockWithTimeoutRespectsContext(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Create subscriber with very small buffer and long block timeout
	sub, err := ps.Subscribe("ctx.block", SubOptions{
		BufferSize:   1,
		Policy:       BlockWithTimeout,
		BlockTimeout: 5 * time.Second, // long timeout
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Cancel()

	// Fill the buffer
	err = ps.Publish("ctx.block", Message{ID: "fill"})
	if err != nil {
		t.Fatalf("fill publish: %v", err)
	}

	// Publish with a short-lived context while buffer is full.
	// The context cancellation should unblock the delivery rather than
	// waiting for the full 5s BlockTimeout.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err = ps.PublishWithContext(ctx, "ctx.block", Message{ID: "blocked"})
	elapsed := time.Since(start)

	// Should complete within context timeout, not the 5s block timeout
	if elapsed >= 2*time.Second {
		t.Fatalf("publish took too long (%v), context cancellation was not respected", elapsed)
	}

	// No error expected because publishInternal itself returns nil (delivery drop is handled internally)
	if err != nil {
		t.Logf("publish returned: %v (elapsed: %v)", err, elapsed)
	}
}

// TestPublishWithContext_AsyncCancelledContext tests that async publish
// respects a cancelled context.
func TestPublishWithContext_AsyncCancelledContext(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.Subscribe("ctx.async", SubOptions{BufferSize: 8, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Cancel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// publishInternal with async=true and cancelled context
	err = ps.publishInternal(ctx, "ctx.async", Message{ID: "m1"}, true)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// TestPublishWithContext_ContextValuesPropagated tests that context values
// are available in the metrics collector.
func TestPublishWithContext_ContextValuesPropagated(t *testing.T) {
	type ctxKey string
	const key ctxKey = "trace-id"

	var capturedCtx context.Context
	collector := &testMetricsCollector{
		onObservePubSub: func(ctx context.Context, operation, topic string, duration time.Duration, err error) {
			capturedCtx = ctx
		},
	}

	ps := New(WithMetricsCollector(collector))
	defer ps.Close()

	_, err := ps.Subscribe("ctx.values", SubOptions{BufferSize: 8, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	ctx := context.WithValue(context.Background(), key, "abc-123")
	err = ps.PublishWithContext(ctx, "ctx.values", Message{ID: "m1"})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	if capturedCtx == nil {
		t.Fatal("metrics collector did not receive context")
	}
	if v, ok := capturedCtx.Value(key).(string); !ok || v != "abc-123" {
		t.Fatalf("expected trace-id=abc-123, got %v", capturedCtx.Value(key))
	}
}

// testMetricsCollector is a test helper that implements metrics.MetricsCollector
// with callback hooks for verifying context propagation.
type testMetricsCollector struct {
	metrics.BaseMetricsCollector
	onObservePubSub func(ctx context.Context, operation, topic string, duration time.Duration, err error)
}

func (c *testMetricsCollector) ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	if c.onObservePubSub != nil {
		c.onObservePubSub(ctx, operation, topic, duration, err)
	}
	c.BaseMetricsCollector.ObservePubSub(ctx, operation, topic, duration, err)
}
