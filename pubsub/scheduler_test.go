package pubsub

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// --- messageScheduler integration tests ---

func TestPublishDelayed_Basic(t *testing.T) {
	ps := New(WithScheduler())
	defer ps.Close()

	sub, err := ps.Subscribe("delayed.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{ID: "d-1", Data: "delayed-hello"}
	id, err := ps.PublishDelayed("delayed.topic", msg, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("expected non-zero schedule ID")
	}

	// Message should not be delivered immediately
	select {
	case <-sub.C():
		t.Fatal("message delivered too early")
	case <-time.After(10 * time.Millisecond):
	}

	// Wait for delivery
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case received := <-sub.C():
		if received.Data != "delayed-hello" {
			t.Fatalf("expected 'delayed-hello', got %v", received.Data)
		}
		if received.Topic != "delayed.topic" {
			t.Fatalf("expected topic 'delayed.topic', got %q", received.Topic)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for delayed message")
	}
}

func TestPublishAt_Basic(t *testing.T) {
	ps := New(WithScheduler())
	defer ps.Close()

	sub, err := ps.Subscribe("at.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	at := time.Now().Add(50 * time.Millisecond)
	msg := Message{ID: "a-1", Data: "at-hello"}
	id, err := ps.PublishAt("at.topic", msg, at)
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("expected non-zero schedule ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case received := <-sub.C():
		if received.Data != "at-hello" {
			t.Fatalf("expected 'at-hello', got %v", received.Data)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for scheduled message")
	}
}

func TestPublishDelayed_DisabledScheduler(t *testing.T) {
	ps := New() // No WithScheduler()
	defer ps.Close()

	_, err := ps.PublishDelayed("topic", Message{}, time.Second)
	if err != ErrSchedulerDisabled {
		t.Fatalf("expected ErrSchedulerDisabled, got %v", err)
	}

	_, err = ps.PublishAt("topic", Message{}, time.Now().Add(time.Second))
	if err != ErrSchedulerDisabled {
		t.Fatalf("expected ErrSchedulerDisabled, got %v", err)
	}
}

func TestPublishDelayed_ClosedPubSub(t *testing.T) {
	ps := New(WithScheduler())
	ps.Close()

	_, err := ps.PublishDelayed("topic", Message{}, time.Second)
	if err != ErrPublishToClosed {
		t.Fatalf("expected ErrPublishToClosed, got %v", err)
	}
}

func TestPublishDelayed_InvalidTopic(t *testing.T) {
	ps := New(WithScheduler())
	defer ps.Close()

	_, err := ps.PublishDelayed("", Message{}, time.Second)
	if err != ErrInvalidTopic {
		t.Fatalf("expected ErrInvalidTopic, got %v", err)
	}

	_, err = ps.PublishAt("  ", Message{}, time.Now())
	if err != ErrInvalidTopic {
		t.Fatalf("expected ErrInvalidTopic, got %v", err)
	}
}

func TestCancelScheduled(t *testing.T) {
	ps := New(WithScheduler())
	defer ps.Close()

	sub, err := ps.Subscribe("cancel.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{ID: "c-1", Data: "cancel-me"}
	id, err := ps.PublishDelayed("cancel.topic", msg, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	if ps.ScheduledCount() != 1 {
		t.Fatalf("expected 1 scheduled, got %d", ps.ScheduledCount())
	}

	ok := ps.CancelScheduled(id)
	if !ok {
		t.Fatal("expected cancel to succeed")
	}

	if ps.ScheduledCount() != 0 {
		t.Fatalf("expected 0 scheduled after cancel, got %d", ps.ScheduledCount())
	}

	// Message should not be delivered
	select {
	case <-sub.C():
		t.Fatal("cancelled message was delivered")
	case <-time.After(200 * time.Millisecond):
		// Expected
	}
}

func TestCancelScheduled_DisabledScheduler(t *testing.T) {
	ps := New()
	defer ps.Close()

	if ps.CancelScheduled(1) {
		t.Fatal("expected false when scheduler disabled")
	}
	if ps.ScheduledCount() != 0 {
		t.Fatalf("expected 0, got %d", ps.ScheduledCount())
	}
}

func TestScheduledCount(t *testing.T) {
	ps := New(WithScheduler())
	defer ps.Close()

	for i := 0; i < 5; i++ {
		_, err := ps.PublishDelayed("count.topic", Message{Data: i}, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
	}

	if ps.ScheduledCount() != 5 {
		t.Fatalf("expected 5 scheduled, got %d", ps.ScheduledCount())
	}
}

func TestPublishDelayed_MultipleMessages(t *testing.T) {
	ps := New(WithScheduler())
	defer ps.Close()

	sub, err := ps.Subscribe("multi.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	// Schedule messages in reverse order: second message should arrive first
	_, err = ps.PublishDelayed("multi.topic", Message{ID: "m-2", Data: "second"}, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ps.PublishDelayed("multi.topic", Message{ID: "m-1", Data: "first"}, 30*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// First delivered should be "first" (shorter delay)
	select {
	case received := <-sub.C():
		if received.Data != "first" {
			t.Fatalf("expected 'first', got %v", received.Data)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for first message")
	}

	// Second delivered should be "second" (longer delay)
	select {
	case received := <-sub.C():
		if received.Data != "second" {
			t.Fatalf("expected 'second', got %v", received.Data)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for second message")
	}
}

func TestPublishAt_PastTime(t *testing.T) {
	ps := New(WithScheduler())
	defer ps.Close()

	sub, err := ps.Subscribe("past.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	// Schedule for the past — should be delivered immediately
	past := time.Now().Add(-time.Hour)
	_, err = ps.PublishAt("past.topic", Message{ID: "p-1", Data: "past"}, past)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case received := <-sub.C():
		if received.Data != "past" {
			t.Fatalf("expected 'past', got %v", received.Data)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for past-scheduled message")
	}
}

// --- ttlManager integration tests ---

func TestPublishWithTTL_Basic(t *testing.T) {
	ps := New(WithTTL())
	defer ps.Close()

	sub, err := ps.Subscribe("ttl.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{Data: "ttl-hello"}
	err = ps.PublishWithTTL("ttl.topic", msg, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case received := <-sub.C():
		if received.Data != "ttl-hello" {
			t.Fatalf("expected 'ttl-hello', got %v", received.Data)
		}
		// Verify TTL metadata
		if received.Meta["X-Message-TTL"] != "5s" {
			t.Fatalf("expected TTL metadata '5s', got %q", received.Meta["X-Message-TTL"])
		}
		if received.Meta["X-Message-ExpiresAt"] == "" {
			t.Fatal("expected X-Message-ExpiresAt to be set")
		}
		if received.ID == "" {
			t.Fatal("expected message ID to be auto-generated")
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for TTL message")
	}
}

func TestPublishWithTTL_DisabledTTL(t *testing.T) {
	ps := New() // No WithTTL()
	defer ps.Close()

	err := ps.PublishWithTTL("topic", Message{}, time.Second)
	if err != ErrTTLDisabled {
		t.Fatalf("expected ErrTTLDisabled, got %v", err)
	}
}

func TestPublishWithTTL_ClosedPubSub(t *testing.T) {
	ps := New(WithTTL())
	ps.Close()

	err := ps.PublishWithTTL("topic", Message{}, time.Second)
	if err != ErrPublishToClosed {
		t.Fatalf("expected ErrPublishToClosed, got %v", err)
	}
}

func TestPublishWithTTL_InvalidTopic(t *testing.T) {
	ps := New(WithTTL())
	defer ps.Close()

	err := ps.PublishWithTTL("", Message{}, time.Second)
	if err != ErrInvalidTopic {
		t.Fatalf("expected ErrInvalidTopic, got %v", err)
	}
}

func TestPublishWithTTL_ExpiredSkipped(t *testing.T) {
	ps := New(WithScheduler(), WithTTL())
	defer ps.Close()

	sub, err := ps.Subscribe("expire.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	// Publish a message with TTL metadata set to already be expired,
	// then schedule it for delayed delivery. The expired message should
	// be skipped during delivery.
	msg := Message{ID: "exp-1", Data: "should-expire"}
	msg.Meta = map[string]string{
		"X-Message-TTL":       (50 * time.Millisecond).String(),
		"X-Message-ExpiresAt": time.Now().Add(50 * time.Millisecond).UTC().Format(time.RFC3339Nano),
	}

	// Schedule for delivery well after TTL expires
	_, err = ps.PublishDelayed("expire.topic", msg, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// Wait long enough for the scheduled delivery attempt
	select {
	case <-sub.C():
		t.Fatal("expired message should not have been delivered")
	case <-time.After(400 * time.Millisecond):
		// Expected: message expired before delivery
	}
}

func TestGetTTLStats(t *testing.T) {
	ps := New(WithTTL())
	defer ps.Close()

	stats := ps.GetTTLStats()
	if stats.TrackedTopics != 0 || stats.TrackedMessages != 0 {
		t.Fatalf("expected empty stats, got %+v", stats)
	}

	if err := ps.PublishWithTTL("topic-a", Message{Data: "a"}, time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := ps.PublishWithTTL("topic-a", Message{Data: "b"}, time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := ps.PublishWithTTL("topic-b", Message{Data: "c"}, time.Hour); err != nil {
		t.Fatal(err)
	}

	stats = ps.GetTTLStats()
	if stats.TrackedTopics != 2 {
		t.Fatalf("expected 2 tracked topics, got %d", stats.TrackedTopics)
	}
	if stats.TrackedMessages != 3 {
		t.Fatalf("expected 3 tracked messages, got %d", stats.TrackedMessages)
	}
}

func TestGetTTLStats_DisabledTTL(t *testing.T) {
	ps := New()
	defer ps.Close()

	stats := ps.GetTTLStats()
	if stats.TrackedTopics != 0 || stats.TrackedMessages != 0 {
		t.Fatalf("expected zero stats when TTL disabled, got %+v", stats)
	}
}

// --- Combined scheduler + TTL tests ---

func TestSchedulerAndTTL_Combined(t *testing.T) {
	ps := New(WithScheduler(), WithTTL())
	defer ps.Close()

	sub, err := ps.Subscribe("combo.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	// Publish with TTL and verify delivery
	err = ps.PublishWithTTL("combo.topic", Message{Data: "combo-msg"}, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case received := <-sub.C():
		if received.Data != "combo-msg" {
			t.Fatalf("expected 'combo-msg', got %v", received.Data)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for combo message")
	}

	// Delayed publish and verify delivery
	_, err = ps.PublishDelayed("combo.topic", Message{Data: "delayed-combo"}, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	select {
	case received := <-sub.C():
		if received.Data != "delayed-combo" {
			t.Fatalf("expected 'delayed-combo', got %v", received.Data)
		}
	case <-ctx2.Done():
		t.Fatal("timeout waiting for delayed combo message")
	}
}

func TestPublishDelayed_SetsMessageFields(t *testing.T) {
	ps := New(WithScheduler())
	defer ps.Close()

	sub, err := ps.Subscribe("fields.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{Data: "check-fields"}
	_, err = ps.PublishDelayed("fields.topic", msg, 10*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case received := <-sub.C():
		if received.Topic != "fields.topic" {
			t.Fatalf("expected topic 'fields.topic', got %q", received.Topic)
		}
		if received.Time.IsZero() {
			t.Fatal("expected Time to be set")
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for message")
	}
}

// --- messageScheduler unit tests ---

func TestMessageScheduler_Schedule(t *testing.T) {
	ps := New()
	defer ps.Close()

	sched := newMessageScheduler(ps)
	defer sched.Close()

	sub, err := ps.Subscribe("sched.unit", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	id := sched.Schedule("sched.unit", Message{Data: "unit-test"}, 30*time.Millisecond)
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}

	if sched.PendingCount() != 1 {
		t.Fatalf("expected 1 pending, got %d", sched.PendingCount())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case received := <-sub.C():
		if received.Data != "unit-test" {
			t.Fatalf("expected 'unit-test', got %v", received.Data)
		}
	case <-ctx.Done():
		t.Fatal("timeout")
	}
}

func TestMessageScheduler_Cancel(t *testing.T) {
	ps := New()
	defer ps.Close()

	sched := newMessageScheduler(ps)
	defer sched.Close()

	id := sched.Schedule("cancel.unit", Message{Data: "cancel"}, time.Hour)
	if !sched.Cancel(id) {
		t.Fatal("expected cancel to return true")
	}
	if sched.PendingCount() != 0 {
		t.Fatalf("expected 0 pending, got %d", sched.PendingCount())
	}

	// Cancel again should return false
	if sched.Cancel(id) {
		t.Fatal("expected second cancel to return false")
	}
}

func TestMessageScheduler_ClosedScheduler(t *testing.T) {
	ps := New()
	defer ps.Close()

	sched := newMessageScheduler(ps)
	sched.Close()

	id := sched.Schedule("closed", Message{}, time.Millisecond)
	if id != 0 {
		t.Fatalf("expected 0 from closed scheduler, got %d", id)
	}
}

// --- ttlManager unit tests ---

func TestTTLManager_TrackAndExpire(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Use a long cleanup interval so entries aren't removed before IsExpired check
	tm := newTTLManager(ps, time.Minute)
	defer tm.Close()

	tm.Track("topic", "msg-1", 50*time.Millisecond)

	if tm.IsExpired("topic", "msg-1") {
		t.Fatal("message should not be expired yet")
	}

	time.Sleep(100 * time.Millisecond)

	if !tm.IsExpired("topic", "msg-1") {
		t.Fatal("message should be expired")
	}
}

func TestTTLManager_NotTracked(t *testing.T) {
	ps := New()
	defer ps.Close()

	tm := newTTLManager(ps, time.Minute)
	defer tm.Close()

	// Non-tracked messages should not be considered expired
	if tm.IsExpired("topic", "unknown") {
		t.Fatal("non-tracked message should not be expired")
	}
}

func TestTTLManager_Stats(t *testing.T) {
	ps := New()
	defer ps.Close()

	tm := newTTLManager(ps, time.Minute)
	defer tm.Close()

	tm.Track("topic-a", "m1", time.Hour)
	tm.Track("topic-a", "m2", time.Hour)
	tm.Track("topic-b", "m3", time.Hour)

	stats := tm.Stats()
	if stats.TrackedTopics != 2 {
		t.Fatalf("expected 2 topics, got %d", stats.TrackedTopics)
	}
	if stats.TrackedMessages != 3 {
		t.Fatalf("expected 3 messages, got %d", stats.TrackedMessages)
	}
}

func TestTTLManager_Cleanup(t *testing.T) {
	ps := New()
	defer ps.Close()

	tm := newTTLManager(ps, 50*time.Millisecond)
	defer tm.Close()

	tm.Track("topic", "m1", 30*time.Millisecond)
	tm.Track("topic", "m2", time.Hour)

	// Wait for cleanup to run
	time.Sleep(120 * time.Millisecond)

	stats := tm.Stats()
	if stats.TrackedMessages != 1 {
		t.Fatalf("expected 1 remaining after cleanup, got %d", stats.TrackedMessages)
	}
}

func TestTTLManager_ClosedManager(t *testing.T) {
	ps := New()
	defer ps.Close()

	tm := newTTLManager(ps, time.Minute)
	tm.Close()

	// Track should be no-op when closed
	tm.Track("topic", "msg", time.Hour)

	stats := tm.Stats()
	if stats.TrackedMessages != 0 {
		t.Fatalf("expected 0 tracked after close, got %d", stats.TrackedMessages)
	}
}

func TestPublishWithTTL_AutoGeneratesID(t *testing.T) {
	ps := New(WithTTL())
	defer ps.Close()

	sub, err := ps.Subscribe("autoid.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	// Publish without ID — should auto-generate
	err = ps.PublishWithTTL("autoid.topic", Message{Data: "no-id"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case received := <-sub.C():
		if received.ID == "" {
			t.Fatal("expected auto-generated ID")
		}
	case <-ctx.Done():
		t.Fatal("timeout")
	}
}

func TestScheduler_ConcurrentScheduleCancel(t *testing.T) {
	ps := New(WithScheduler())
	defer ps.Close()

	var ids atomic.Int64
	done := make(chan struct{})

	// Concurrently schedule and cancel
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			id, err := ps.PublishDelayed("conc.topic", Message{Data: i}, time.Hour)
			if err != nil {
				return
			}
			if id > 0 {
				ids.Add(1)
			}
			ps.CancelScheduled(id)
		}
	}()

	<-done
	if ids.Load() == 0 {
		t.Fatal("expected some successful schedules")
	}
}
