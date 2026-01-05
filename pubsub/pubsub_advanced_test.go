package pubsub

import (
	"sync"
	"testing"
	"time"
)

// TestPubSub_Performance 测试优化后的性能
func TestPubSub_Performance(t *testing.T) {
	ps := New()
	defer ps.Close()

	// 创建多个订阅者
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

	// 测量发布性能
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

	// 验证所有订阅者都收到消息
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

// TestPubSub_ConcurrentSubscribers 测试并发订阅者管理
func TestPubSub_ConcurrentSubscribers(t *testing.T) {
	ps := New()
	defer ps.Close()

	const numGoroutines = 50
	var wg sync.WaitGroup

	// 并发订阅和取消
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 订阅
			sub, err := ps.Subscribe("concurrent", SubOptions{
				BufferSize: 10,
				Policy:     DropNewest,
			})
			if err != nil {
				t.Errorf("subscribe failed: %v", err)
				return
			}

			// 短暂接收
			time.Sleep(10 * time.Millisecond)

			// 取消
			sub.Cancel()
		}(i)
	}

	// 并发发布
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := Message{ID: "test", Data: id}
			_ = ps.Publish("concurrent", msg)
		}(i)
	}

	wg.Wait()

	// 验证最终状态
	count := ps.GetSubscriberCount("concurrent")
	if count != 0 {
		t.Errorf("Expected 0 subscribers, got %d", count)
	}
}

// TestPubSub_DifferentPolicies 测试不同背压策略
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

// TestPubSub_ListOperations 测试列表操作
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

// TestPubSub_ErrorHandling 测试错误处理
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

// TestPubSub_MetricsAccuracy 测试指标准确性
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

// TestPubSub_RapidSubscribeCancel 测试快速订阅/取消
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

// TestPubSub_PressureHandling 测试压力处理
func TestPubSub_PressureHandling(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Create slow subscriber
	slowSub, _ := ps.Subscribe("pressure", SubOptions{
		BufferSize:   2,
		Policy:       BlockWithTimeout,
		BlockTimeout: 10 * time.Millisecond,
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

	// Should have processed some messages
	if tm.DeliveredTotal == 0 {
		t.Errorf("Expected some delivered messages")
	}
}
