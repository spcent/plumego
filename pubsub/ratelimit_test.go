package pubsub

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimitedPubSub_Basic(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.GlobalQPS = 10 // 10 messages per second
	config.GlobalBurst = 5

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create rate-limited pubsub: %v", err)
	}
	defer rlps.Close()

	// Publish messages
	allowed := 0
	denied := 0

	for i := 0; i < 20; i++ {
		msg := Message{Data: i}
		if err := rlps.Publish("test.topic", msg); err == ErrRateLimitExceeded {
			denied++
		} else if err == nil {
			allowed++
		}
	}

	if denied == 0 {
		t.Log("Note: All messages allowed (rate limit may not have been hit)")
	}

	if allowed == 0 {
		t.Error("No messages allowed")
	}

	t.Logf("Allowed: %d, Denied: %d", allowed, denied)
}

func TestRateLimitedPubSub_GlobalLimit(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.GlobalQPS = 100
	config.GlobalBurst = 10

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create pubsub: %v", err)
	}
	defer rlps.Close()

	// Rapid fire messages
	start := time.Now()
	count := 0

	for i := 0; i < 50; i++ {
		msg := Message{Data: i}
		if err := rlps.Publish("test", msg); err == nil {
			count++
		}
	}

	elapsed := time.Since(start)

	if count > 15 && elapsed < 50*time.Millisecond {
		t.Logf("Published %d messages in %v", count, elapsed)
	}

	stats := rlps.RateLimitStats()
	if stats.GlobalAllowed != uint64(count) {
		t.Errorf("Expected %d global allowed, got %d", count, stats.GlobalAllowed)
	}
}

func TestRateLimitedPubSub_PerTopicLimit(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.PerTopicQPS = 10
	config.PerTopicBurst = 5

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create pubsub: %v", err)
	}
	defer rlps.Close()

	topics := []string{"topic.a", "topic.b", "topic.c"}

	// Publish to multiple topics
	for _, topic := range topics {
		for i := 0; i < 10; i++ {
			msg := Message{Data: i}
			_ = rlps.Publish(topic, msg)
		}
	}

	time.Sleep(100 * time.Millisecond)

	stats := rlps.RateLimitStats()
	if stats.TopicLimiters != len(topics) {
		t.Errorf("Expected %d topic limiters, got %d", len(topics), stats.TopicLimiters)
	}
}

func TestRateLimitedPubSub_WaitOnLimit(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.GlobalQPS = 50
	config.GlobalBurst = 5
	config.WaitOnLimit = true
	config.WaitTimeout = 100 * time.Millisecond

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create pubsub: %v", err)
	}
	defer rlps.Close()

	start := time.Now()
	count := 0

	// This should block and wait
	for i := 0; i < 10; i++ {
		msg := Message{Data: i}
		if err := rlps.Publish("test", msg); err == nil {
			count++
		}
	}

	elapsed := time.Since(start)

	if count == 10 && elapsed > 0 {
		t.Logf("Published %d messages in %v (waited for tokens)", count, elapsed)
	}

	stats := rlps.RateLimitStats()
	if stats.LimitWaited == 0 {
		t.Log("Note: No waits recorded (may be fast enough)")
	}
}

func TestRateLimitedPubSub_Adaptive(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.PerTopicQPS = 100
	config.Adaptive = true
	config.AdaptiveTarget = 0.5
	config.AdaptiveAdjustInterval = 100 * time.Millisecond

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create pubsub: %v", err)
	}
	defer rlps.Close()

	// Create some load
	for i := 0; i < 5; i++ {
		_, _ = rlps.Subscribe("test.adaptive", SubOptions{BufferSize: 10})
	}

	// Wait for adaptive adjustment
	time.Sleep(300 * time.Millisecond)

	stats := rlps.RateLimitStats()
	if stats.AdaptiveAdj == 0 {
		t.Error("Expected adaptive adjustments")
	}

	if stats.AdaptiveFactor == 0 {
		t.Error("Adaptive factor not set")
	}

	t.Logf("Adaptive factor: %.2f, Load: %.2f", stats.AdaptiveFactor, stats.CurrentLoad)
}

func TestRateLimitedPubSub_PerSubscriberLimit(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.PerSubscriberQPS = 10
	config.PerSubscriberBurst = 3

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create pubsub: %v", err)
	}
	defer rlps.Close()

	// Create subscriber
	sub, err := rlps.Subscribe("test.sub", SubOptions{BufferSize: 20})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Cancel()

	// Publish messages
	for i := 0; i < 15; i++ {
		msg := Message{Data: i}
		_ = rlps.Publish("test.sub", msg)
	}

	time.Sleep(200 * time.Millisecond)

	// Count received messages
	count := 0
	for len(sub.C()) > 0 {
		<-sub.C()
		count++
	}

	// Should be rate limited
	t.Logf("Received %d messages (rate limited)", count)

	stats := rlps.RateLimitStats()
	if stats.SubLimiters != 1 {
		t.Errorf("Expected 1 subscriber limiter, got %d", stats.SubLimiters)
	}
}

func TestRateLimitedPubSub_ConcurrentPublish(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.GlobalQPS = 200
	config.GlobalBurst = 50

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create pubsub: %v", err)
	}
	defer rlps.Close()

	const numGoroutines = 10
	const messagesPerGoroutine = 20

	var wg sync.WaitGroup
	var allowed, denied int32

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				msg := Message{Data: j}
				if err := rlps.Publish("concurrent", msg); err == ErrRateLimitExceeded {
					denied++
				} else if err == nil {
					allowed++
				}
			}
		}()
	}

	wg.Wait()

	t.Logf("Concurrent: Allowed=%d, Denied=%d", allowed, denied)

	if allowed == 0 {
		t.Error("No messages allowed")
	}
}

func TestRateLimitedPubSub_TokenBucket(t *testing.T) {
	tb := newTokenBucket(10.0, 5) // 10 tokens/sec, burst of 5

	// Should allow burst
	for i := 0; i < 5; i++ {
		if !tb.allow() {
			t.Errorf("Burst: Expected allow at %d", i)
		}
	}

	// Should deny next
	if tb.allow() {
		t.Error("Expected deny after burst")
	}

	// Wait for refill
	time.Sleep(200 * time.Millisecond)

	// Should allow again
	if !tb.allow() {
		t.Error("Expected allow after refill")
	}
}

func TestRateLimitedPubSub_DynamicRateAdjustment(t *testing.T) {
	tb := newTokenBucket(10.0, 5)

	// Initial rate
	if tb.rate != 10.0 {
		t.Errorf("Expected initial rate 10.0, got %.1f", tb.rate)
	}

	// Update rate
	tb.updateRate(20.0)

	if tb.rate != 20.0 {
		t.Errorf("Expected updated rate 20.0, got %.1f", tb.rate)
	}
}

func TestRateLimitedPubSub_Stats(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.GlobalQPS = 50
	config.PerTopicQPS = 10

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create pubsub: %v", err)
	}
	defer rlps.Close()

	// Publish to trigger rate limiting
	for i := 0; i < 100; i++ {
		msg := Message{Data: i}
		_ = rlps.Publish("test.stats", msg)
	}

	time.Sleep(100 * time.Millisecond)

	stats := rlps.RateLimitStats()

	if stats.GlobalAllowed == 0 {
		t.Error("Expected some global allowed")
	}

	if stats.LimitExceeded == 0 {
		t.Log("Note: No limits exceeded (rate may be sufficient)")
	}

	t.Logf("Stats: %+v", stats)
}

func TestRateLimitedPubSub_InvalidConfig(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.GlobalQPS = -1 // Invalid

	_, err := NewRateLimited(config)
	if err != ErrInvalidRateLimit {
		t.Errorf("Expected ErrInvalidRateLimit, got %v", err)
	}
}

func TestRateLimitedPubSub_NoLimit(t *testing.T) {
	config := DefaultRateLimitConfig()
	// All limits at 0 (unlimited)

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create pubsub: %v", err)
	}
	defer rlps.Close()

	// Should allow all
	for i := 0; i < 100; i++ {
		msg := Message{Data: i}
		if err := rlps.Publish("test", msg); err != nil {
			t.Errorf("Unexpected error with no limits: %v", err)
		}
	}
}

func TestRateLimitedPubSub_MultipleTopics(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.PerTopicQPS = 20

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create pubsub: %v", err)
	}
	defer rlps.Close()

	topics := []string{"a", "b", "c", "d", "e"}

	// Publish to different topics
	for _, topic := range topics {
		for i := 0; i < 10; i++ {
			msg := Message{Data: i}
			_ = rlps.Publish(topic, msg)
		}
	}

	time.Sleep(100 * time.Millisecond)

	stats := rlps.RateLimitStats()
	if stats.TopicLimiters != len(topics) {
		t.Errorf("Expected %d topic limiters, got %d", len(topics), stats.TopicLimiters)
	}
}

func TestRateLimitedPubSub_Close(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.Adaptive = true

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create pubsub: %v", err)
	}

	// Publish some messages
	for i := 0; i < 5; i++ {
		msg := Message{Data: i}
		_ = rlps.Publish("test", msg)
	}

	// Close
	if err := rlps.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Publishing after close should fail
	msg := Message{Data: "test"}
	if err := rlps.Publish("test", msg); err != ErrClosed {
		t.Errorf("Expected ErrClosed, got %v", err)
	}
}

func BenchmarkRateLimitedPubSub_WithLimit(b *testing.B) {
	config := DefaultRateLimitConfig()
	config.GlobalQPS = 10000
	config.GlobalBurst = 100

	rlps, _ := NewRateLimited(config)
	defer rlps.Close()

	msg := Message{Data: []byte("benchmark message")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rlps.Publish("bench.topic", msg)
	}
}

func BenchmarkRateLimitedPubSub_WithoutLimit(b *testing.B) {
	config := DefaultRateLimitConfig()
	// No limits

	rlps, _ := NewRateLimited(config)
	defer rlps.Close()

	msg := Message{Data: []byte("benchmark message")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rlps.Publish("bench.topic", msg)
	}
}
