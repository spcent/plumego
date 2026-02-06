package pubsub

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestAckableSubscription_AckRemovesPending(t *testing.T) {
	ps := New()
	defer ps.Close()

	ackOpts := AckOptions{
		AckTimeout:      5 * time.Second,
		MaxRetries:      3,
		DeadLetterTopic: "dlq.test",
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	// Publish a message
	msg := Message{ID: "msg-1", Topic: "test.topic", Data: "hello"}
	if err := ps.Publish("test.topic", msg); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	received, err := sub.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if sub.PendingCount() != 1 {
		t.Fatalf("expected 1 pending, got %d", sub.PendingCount())
	}

	if err := sub.Ack(received.ID); err != nil {
		t.Fatal(err)
	}

	if sub.PendingCount() != 0 {
		t.Fatalf("expected 0 pending after ack, got %d", sub.PendingCount())
	}
}

func TestAckableSubscription_NackNoRequeue_SendsToDLQ(t *testing.T) {
	ps := New()
	defer ps.Close()

	ackOpts := AckOptions{
		AckTimeout:      5 * time.Second,
		MaxRetries:      3,
		DeadLetterTopic: "dlq.test",
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	if sub.DeadLetterCount() != 0 {
		t.Fatalf("expected 0 dead letters initially, got %d", sub.DeadLetterCount())
	}

	msg := Message{ID: "msg-nack", Topic: "test.topic", Data: "fail"}
	if err := ps.Publish("test.topic", msg); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	received, err := sub.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Nack without requeue
	if err := sub.Nack(received.ID, false); err != nil {
		t.Fatal(err)
	}

	if sub.PendingCount() != 0 {
		t.Fatalf("expected 0 pending after nack, got %d", sub.PendingCount())
	}

	if sub.DeadLetterCount() != 1 {
		t.Fatalf("expected 1 dead letter, got %d", sub.DeadLetterCount())
	}

	dlMsgs := sub.DeadLetterMessages()
	if len(dlMsgs) != 1 {
		t.Fatalf("expected 1 dead letter message, got %d", len(dlMsgs))
	}
	if dlMsgs[0].Reason != "nack_no_requeue" {
		t.Fatalf("expected reason 'nack_no_requeue', got %q", dlMsgs[0].Reason)
	}
	if dlMsgs[0].Message.ID != "msg-nack" {
		t.Fatalf("expected message ID 'msg-nack', got %q", dlMsgs[0].Message.ID)
	}
}

func TestAckableSubscription_NackRequeue_ExceedsMaxRetries(t *testing.T) {
	ps := New()
	defer ps.Close()

	// With maxRetries=1, the first Nack with requeue immediately exceeds the limit
	ackOpts := AckOptions{
		AckTimeout:      5 * time.Second,
		MaxRetries:      1,
		DeadLetterTopic: "dlq.test",
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{ID: "msg-retry", Topic: "test.topic", Data: "retry-me"}
	if err := ps.Publish("test.topic", msg); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	received, err := sub.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Nack with requeue — retries becomes 1 which equals maxRetries, so goes to DLQ
	if err := sub.Nack(received.ID, true); err != nil {
		t.Fatal(err)
	}

	if sub.DeadLetterCount() != 1 {
		t.Fatalf("expected 1 dead letter after exceeding retries, got %d", sub.DeadLetterCount())
	}

	dlMsgs := sub.DeadLetterMessages()
	if dlMsgs[0].Reason != "max_retries_exceeded" {
		t.Fatalf("expected reason 'max_retries_exceeded', got %q", dlMsgs[0].Reason)
	}
}

func TestAckableSubscription_DLQ_WithCallback(t *testing.T) {
	ps := New()
	defer ps.Close()

	var callbackCount atomic.Int32
	var callbackReason atomic.Value

	ackOpts := AckOptions{
		AckTimeout:      5 * time.Second,
		MaxRetries:      3,
		DeadLetterTopic: "dlq.test",
		DLQConfig: DeadLetterConfig{
			OnDeadLetter: func(msg Message, reason string) {
				callbackCount.Add(1)
				callbackReason.Store(reason)
			},
		},
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{ID: "msg-cb", Topic: "test.topic", Data: "callback"}
	if err := ps.Publish("test.topic", msg); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	received, err := sub.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := sub.Nack(received.ID, false); err != nil {
		t.Fatal(err)
	}

	// Wait for async callback
	time.Sleep(50 * time.Millisecond)

	if callbackCount.Load() != 1 {
		t.Fatalf("expected callback called once, got %d", callbackCount.Load())
	}
	if reason, ok := callbackReason.Load().(string); !ok || reason != "nack_no_requeue" {
		t.Fatalf("expected callback reason 'nack_no_requeue', got %v", callbackReason.Load())
	}
}

func TestAckableSubscription_DLQ_MaxSize(t *testing.T) {
	ps := New()
	defer ps.Close()

	ackOpts := AckOptions{
		AckTimeout:      5 * time.Second,
		MaxRetries:      3,
		DeadLetterTopic: "dlq.test",
		DLQConfig: DeadLetterConfig{
			MaxSize: 2,
		},
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	// Send 3 messages to DLQ, only 2 should be retained
	for i := 0; i < 3; i++ {
		msg := Message{ID: "msg-" + string(rune('a'+i)), Topic: "test.topic", Data: i}
		if err := ps.Publish("test.topic", msg); err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		received, err := sub.Receive(ctx)
		cancel()
		if err != nil {
			t.Fatal(err)
		}

		if err := sub.Nack(received.ID, false); err != nil {
			t.Fatal(err)
		}
	}

	if sub.DeadLetterCount() != 2 {
		t.Fatalf("expected 2 dead letters (max size), got %d", sub.DeadLetterCount())
	}
}

func TestAckableSubscription_ClearDeadLetters(t *testing.T) {
	ps := New()
	defer ps.Close()

	ackOpts := AckOptions{
		AckTimeout:      5 * time.Second,
		MaxRetries:      3,
		DeadLetterTopic: "dlq.test",
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{ID: "msg-clear", Topic: "test.topic", Data: "clear-me"}
	if err := ps.Publish("test.topic", msg); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	received, err := sub.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := sub.Nack(received.ID, false); err != nil {
		t.Fatal(err)
	}

	if sub.DeadLetterCount() != 1 {
		t.Fatalf("expected 1 dead letter, got %d", sub.DeadLetterCount())
	}

	sub.ClearDeadLetters()

	if sub.DeadLetterCount() != 0 {
		t.Fatalf("expected 0 dead letters after clear, got %d", sub.DeadLetterCount())
	}
}

func TestAckableSubscription_ReprocessDeadLetters(t *testing.T) {
	ps := New()
	defer ps.Close()

	ackOpts := AckOptions{
		AckTimeout:      5 * time.Second,
		MaxRetries:      3,
		DeadLetterTopic: "dlq.test",
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	// Subscribe to the original topic to catch reprocessed messages
	reprocessSub, err := ps.Subscribe("test.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer reprocessSub.Cancel()

	msg := Message{ID: "msg-reprocess", Topic: "test.topic", Data: "reprocess-me"}
	if err := ps.Publish("test.topic", msg); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	received, err := sub.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := sub.Nack(received.ID, false); err != nil {
		t.Fatal(err)
	}

	if sub.DeadLetterCount() != 1 {
		t.Fatalf("expected 1 dead letter, got %d", sub.DeadLetterCount())
	}

	// Drain the reprocessSub channel (it also receives the initial publish)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	select {
	case <-reprocessSub.C():
	case <-ctx2.Done():
	}

	// Reprocess all dead letters
	n := sub.ReprocessDeadLetters(nil)
	if n != 1 {
		t.Fatalf("expected 1 reprocessed, got %d", n)
	}

	if sub.DeadLetterCount() != 0 {
		t.Fatalf("expected 0 dead letters after reprocess, got %d", sub.DeadLetterCount())
	}

	// Verify message was republished
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second)
	defer cancel3()
	select {
	case reMsg := <-reprocessSub.C():
		if reMsg.Data != "reprocess-me" {
			t.Fatalf("expected reprocessed data 'reprocess-me', got %v", reMsg.Data)
		}
	case <-ctx3.Done():
		t.Fatal("timeout waiting for reprocessed message")
	}
}

func TestAckableSubscription_ReprocessDeadLetters_WithFilter(t *testing.T) {
	ps := New()
	defer ps.Close()

	ackOpts := AckOptions{
		AckTimeout:      5 * time.Second,
		MaxRetries:      3,
		DeadLetterTopic: "dlq.test",
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	// Send two messages to DLQ with different reasons
	for _, id := range []string{"keep", "reprocess"} {
		msg := Message{ID: id, Topic: "test.topic", Data: id}
		if err := ps.Publish("test.topic", msg); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		received, err := sub.Receive(ctx)
		cancel()
		if err != nil {
			t.Fatal(err)
		}
		if err := sub.Nack(received.ID, false); err != nil {
			t.Fatal(err)
		}
	}

	if sub.DeadLetterCount() != 2 {
		t.Fatalf("expected 2 dead letters, got %d", sub.DeadLetterCount())
	}

	// Only reprocess messages with ID "reprocess"
	n := sub.ReprocessDeadLetters(func(dm DeadLetterMessage) bool {
		return dm.Message.ID == "reprocess"
	})
	if n != 1 {
		t.Fatalf("expected 1 reprocessed, got %d", n)
	}

	if sub.DeadLetterCount() != 1 {
		t.Fatalf("expected 1 remaining dead letter, got %d", sub.DeadLetterCount())
	}

	remaining := sub.DeadLetterMessages()
	if remaining[0].Message.ID != "keep" {
		t.Fatalf("expected remaining message ID 'keep', got %q", remaining[0].Message.ID)
	}
}

func TestAckableSubscription_DLQ_Metadata(t *testing.T) {
	ps := New()
	defer ps.Close()

	ackOpts := AckOptions{
		AckTimeout:      5 * time.Second,
		MaxRetries:      3,
		DeadLetterTopic: "dlq.test",
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{ID: "msg-meta", Topic: "test.topic", Data: "meta-test"}
	if err := ps.Publish("test.topic", msg); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	received, err := sub.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := sub.Nack(received.ID, false); err != nil {
		t.Fatal(err)
	}

	dlMsgs := sub.DeadLetterMessages()
	if len(dlMsgs) != 1 {
		t.Fatalf("expected 1 dead letter, got %d", len(dlMsgs))
	}

	dm := dlMsgs[0]
	if dm.Message.Meta["X-Dead-Letter-Reason"] != "nack_no_requeue" {
		t.Fatalf("expected X-Dead-Letter-Reason 'nack_no_requeue', got %q", dm.Message.Meta["X-Dead-Letter-Reason"])
	}
	if dm.Message.Meta["X-Original-Topic"] != "test.topic" {
		t.Fatalf("expected X-Original-Topic 'test.topic', got %q", dm.Message.Meta["X-Original-Topic"])
	}
	if dm.Message.Meta["X-Dead-Letter-Time"] == "" {
		t.Fatal("expected X-Dead-Letter-Time to be set")
	}
	if dm.Sequence != 1 {
		t.Fatalf("expected sequence 1, got %d", dm.Sequence)
	}
}

func TestAckableSubscription_NoDLQ_WhenNotConfigured(t *testing.T) {
	ps := New()
	defer ps.Close()

	// No DeadLetterTopic, no DLQConfig — DLQ should not be created
	ackOpts := AckOptions{
		AckTimeout: 5 * time.Second,
		MaxRetries: 3,
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	if sub.DeadLetterCount() != 0 {
		t.Fatalf("expected 0 dead letters when DLQ not configured, got %d", sub.DeadLetterCount())
	}
	if sub.DeadLetterMessages() != nil {
		t.Fatal("expected nil dead letter messages when DLQ not configured")
	}
	if sub.ReprocessDeadLetters(nil) != 0 {
		t.Fatal("expected 0 reprocessed when DLQ not configured")
	}
	// ClearDeadLetters should be safe to call
	sub.ClearDeadLetters()
}

func TestAckableSubscription_DLQConfig_OnlyLocalStorage(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Use DLQConfig without DeadLetterTopic — local-only DLQ
	ackOpts := AckOptions{
		AckTimeout: 5 * time.Second,
		MaxRetries: 3,
		DLQConfig: DeadLetterConfig{
			Topic: "local-dlq",
		},
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{ID: "msg-local", Topic: "test.topic", Data: "local-only"}
	if err := ps.Publish("test.topic", msg); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	received, err := sub.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := sub.Nack(received.ID, false); err != nil {
		t.Fatal(err)
	}

	// Message should be stored in local DLQ
	if sub.DeadLetterCount() != 1 {
		t.Fatalf("expected 1 dead letter in local DLQ, got %d", sub.DeadLetterCount())
	}
}

func TestAckableSubscription_Timeout_SendsToDLQ(t *testing.T) {
	ps := New()
	defer ps.Close()

	ackOpts := AckOptions{
		AckTimeout:      50 * time.Millisecond,
		MaxRetries:      1,
		DeadLetterTopic: "dlq.test",
	}

	sub, err := ps.SubscribeAckable("test.topic", DefaultSubOptions(), ackOpts)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	msg := Message{ID: "msg-timeout", Topic: "test.topic", Data: "timeout-me"}
	if err := ps.Publish("test.topic", msg); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := sub.Receive(ctx); err != nil {
		t.Fatal(err)
	}

	// Wait for ack timeout + redelivery timeout to trigger DLQ
	time.Sleep(200 * time.Millisecond)

	if sub.DeadLetterCount() != 1 {
		t.Fatalf("expected 1 dead letter after timeout, got %d", sub.DeadLetterCount())
	}

	dlMsgs := sub.DeadLetterMessages()
	if dlMsgs[0].Reason != "max_retries_exceeded" {
		t.Fatalf("expected reason 'max_retries_exceeded', got %q", dlMsgs[0].Reason)
	}
}

func TestDeadLetterQueue_Unit(t *testing.T) {
	dlq := newDeadLetterQueue(DeadLetterConfig{
		Topic:   "test-dlq",
		MaxSize: 5,
	})

	// Test empty queue
	if dlq.Count() != 0 {
		t.Fatalf("expected 0 count, got %d", dlq.Count())
	}
	if len(dlq.List()) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(dlq.List()))
	}

	// Add messages
	for i := 0; i < 3; i++ {
		dlq.Add(Message{ID: "m" + string(rune('0'+i)), Topic: "test"}, "test_reason")
	}

	if dlq.Count() != 3 {
		t.Fatalf("expected 3 messages, got %d", dlq.Count())
	}

	// Test List returns a copy
	list := dlq.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 in list, got %d", len(list))
	}

	// Verify sequence numbers
	for i, dm := range list {
		if dm.Sequence != uint64(i+1) {
			t.Fatalf("expected sequence %d, got %d", i+1, dm.Sequence)
		}
	}

	// Clear
	dlq.Clear()
	if dlq.Count() != 0 {
		t.Fatalf("expected 0 after clear, got %d", dlq.Count())
	}
}

func TestDeadLetterQueue_MaxSizeTrim(t *testing.T) {
	dlq := newDeadLetterQueue(DeadLetterConfig{
		Topic:   "test-dlq",
		MaxSize: 2,
	})

	dlq.Add(Message{ID: "a"}, "r1")
	dlq.Add(Message{ID: "b"}, "r2")
	dlq.Add(Message{ID: "c"}, "r3")

	if dlq.Count() != 2 {
		t.Fatalf("expected 2 (max size), got %d", dlq.Count())
	}

	list := dlq.List()
	// Oldest message ("a") should have been trimmed
	if list[0].Message.ID != "b" {
		t.Fatalf("expected first message 'b', got %q", list[0].Message.ID)
	}
	if list[1].Message.ID != "c" {
		t.Fatalf("expected second message 'c', got %q", list[1].Message.ID)
	}
}

func TestDeadLetterQueue_Reprocess(t *testing.T) {
	ps := New()
	defer ps.Close()

	dlq := newDeadLetterQueue(DeadLetterConfig{Topic: "test-dlq"})

	// Subscribe to original topic
	sub, err := ps.Subscribe("original.topic", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	// Add a message with original topic metadata
	msg := Message{
		ID:    "reprocess-1",
		Topic: "original.topic",
		Meta: map[string]string{
			"X-Original-Topic":     "original.topic",
			"X-Dead-Letter-Reason": "test",
		},
	}
	dlq.Add(msg, "test")

	n := dlq.Reprocess(ps, nil)
	if n != 1 {
		t.Fatalf("expected 1 reprocessed, got %d", n)
	}
	if dlq.Count() != 0 {
		t.Fatalf("expected 0 after reprocess, got %d", dlq.Count())
	}

	// Verify message was published
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case reMsg := <-sub.C():
		if reMsg.ID != "reprocess-1" {
			t.Fatalf("expected message ID 'reprocess-1', got %q", reMsg.ID)
		}
		// Dead letter metadata should be cleared
		if _, ok := reMsg.Meta["X-Dead-Letter-Reason"]; ok {
			t.Fatal("expected X-Dead-Letter-Reason to be cleared")
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for reprocessed message")
	}
}
