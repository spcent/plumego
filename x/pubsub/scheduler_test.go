package pubsub

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// publishWithTTL is a local test helper that sets TTL metadata and publishes.
func publishWithTTL(b *InProcBroker, tm *ttlManager, topic string, msg Message, ttl time.Duration) error {
	if msg.ID == "" {
		msg.ID = generateCorrelationID()
	}
	if msg.Meta == nil {
		msg.Meta = make(map[string]string)
	}
	msg.Meta["X-Message-TTL"] = ttl.String()
	msg.Meta["X-Message-ExpiresAt"] = time.Now().Add(ttl).UTC().Format(time.RFC3339Nano)
	tm.Track(topic, msg.ID, ttl)
	return b.Publish(topic, msg)
}

// --- messageScheduler integration tests ---

func TestPublishDelayed_Basic(t *testing.T) {
	ps := New()
	defer ps.Close()

	sched := newMessageScheduler(ps)
	defer sched.Close()

	sub, err := ps.Subscribe("delayed.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{ID: "d-1", Data: "delayed-hello"}
	id := sched.Schedule("delayed.topic", msg, 50*time.Millisecond)
	if id == 0 {
		t.Fatal("expected non-zero schedule ID")
	}

	select {
	case <-sub.C():
		t.Fatal("message delivered too early")
	case <-time.After(10 * time.Millisecond):
	}

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
	ps := New()
	defer ps.Close()

	sched := newMessageScheduler(ps)
	defer sched.Close()

	sub, err := ps.Subscribe("at.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	at := time.Now().Add(50 * time.Millisecond)
	msg := Message{ID: "a-1", Data: "at-hello"}
	id := sched.ScheduleAt("at.topic", msg, at)
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

func TestCancelScheduled(t *testing.T) {
	ps := New()
	defer ps.Close()

	sched := newMessageScheduler(ps)
	defer sched.Close()

	sub, err := ps.Subscribe("cancel.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{ID: "c-1", Data: "cancel-me"}
	id := sched.Schedule("cancel.topic", msg, 100*time.Millisecond)

	if sched.PendingCount() != 1 {
		t.Fatalf("expected 1 scheduled, got %d", sched.PendingCount())
	}

	ok := sched.Cancel(id)
	if !ok {
		t.Fatal("expected cancel to succeed")
	}

	if sched.PendingCount() != 0 {
		t.Fatalf("expected 0 scheduled after cancel, got %d", sched.PendingCount())
	}

	select {
	case <-sub.C():
		t.Fatal("cancelled message was delivered")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestScheduledCount(t *testing.T) {
	ps := New()
	defer ps.Close()

	sched := newMessageScheduler(ps)
	defer sched.Close()

	for i := 0; i < 5; i++ {
		sched.Schedule("count.topic", Message{Data: i}, time.Hour)
	}

	if sched.PendingCount() != 5 {
		t.Fatalf("expected 5 scheduled, got %d", sched.PendingCount())
	}
}

func TestPublishDelayed_MultipleMessages(t *testing.T) {
	ps := New()
	defer ps.Close()

	sched := newMessageScheduler(ps)
	defer sched.Close()

	sub, err := ps.Subscribe("multi.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	sched.Schedule("multi.topic", Message{ID: "m-2", Data: "second"}, 100*time.Millisecond)
	sched.Schedule("multi.topic", Message{ID: "m-1", Data: "first"}, 30*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	select {
	case received := <-sub.C():
		if received.Data != "first" {
			t.Fatalf("expected 'first', got %v", received.Data)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for first message")
	}

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
	ps := New()
	defer ps.Close()

	sched := newMessageScheduler(ps)
	defer sched.Close()

	sub, err := ps.Subscribe("past.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	past := time.Now().Add(-time.Hour)
	sched.ScheduleAt("past.topic", Message{ID: "p-1", Data: "past"}, past)

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
	ps := New()
	defer ps.Close()

	tm := newTTLManager(ps, time.Minute)
	defer tm.Close()

	sub, err := ps.Subscribe("ttl.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	if err := publishWithTTL(ps, tm, "ttl.topic", Message{Data: "ttl-hello"}, 5*time.Second); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case received := <-sub.C():
		if received.Data != "ttl-hello" {
			t.Fatalf("expected 'ttl-hello', got %v", received.Data)
		}
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

func TestPublishWithTTL_AutoGeneratesID(t *testing.T) {
	ps := New()
	defer ps.Close()

	tm := newTTLManager(ps, time.Minute)
	defer tm.Close()

	sub, err := ps.Subscribe("autoid.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	if err := publishWithTTL(ps, tm, "autoid.topic", Message{Data: "no-id"}, time.Hour); err != nil {
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

	tm.Track("topic", "msg", time.Hour)

	stats := tm.Stats()
	if stats.TrackedMessages != 0 {
		t.Fatalf("expected 0 tracked after close, got %d", stats.TrackedMessages)
	}
}

func TestScheduler_ConcurrentScheduleCancel(t *testing.T) {
	ps := New()
	defer ps.Close()

	sched := newMessageScheduler(ps)
	defer sched.Close()

	var ids atomic.Int64
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			id := sched.Schedule("conc.topic", Message{Data: i}, time.Hour)
			if id > 0 {
				ids.Add(1)
			}
			sched.Cancel(id)
		}
	}()

	<-done
	if ids.Load() == 0 {
		t.Fatal("expected some successful schedules")
	}
}
