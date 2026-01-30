package pubsub

import (
	"testing"
	"time"
)

func TestPubSub_MultiSubscriberReceive(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub1, err := ps.Subscribe("user.created", SubOptions{BufferSize: 8, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe sub1: %v", err)
	}
	sub2, err := ps.Subscribe("user.created", SubOptions{BufferSize: 8, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe sub2: %v", err)
	}

	msg := Message{ID: "m1", Type: "user.created", Data: map[string]any{"x": 1}}
	if err := ps.Publish("user.created", msg); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case got := <-sub1.C():
		if got.ID != "m1" || got.Topic != "user.created" {
			t.Fatalf("sub1 got unexpected: %+v", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("sub1 timeout")
	}

	select {
	case got := <-sub2.C():
		if got.ID != "m1" || got.Topic != "user.created" {
			t.Fatalf("sub2 got unexpected: %+v", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("sub2 timeout")
	}
}

func TestPubSub_CancelStopsDelivery(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.Subscribe("t", SubOptions{BufferSize: 1, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	sub.Cancel()

	_ = ps.Publish("t", Message{ID: "m1"})
	// channel should be closed
	_, ok := <-sub.C()
	if ok {
		t.Fatalf("expected channel closed")
	}
}

func TestPubSub_PatternSubscription(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribePattern("user.*", SubOptions{BufferSize: 4, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe pattern: %v", err)
	}
	defer sub.Cancel()

	if err := ps.Publish("user.created", Message{ID: "m1"}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case got := <-sub.C():
		if got.Topic != "user.created" {
			t.Fatalf("unexpected topic: %s", got.Topic)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("pattern subscriber timeout")
	}
}

func TestPubSub_InvalidPattern(t *testing.T) {
	ps := New()
	defer ps.Close()

	if _, err := ps.SubscribePattern("[", DefaultSubOptions()); err != ErrInvalidPattern {
		t.Fatalf("expected ErrInvalidPattern, got %v", err)
	}
}

func TestPubSub_MessageImmutability(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub1, err := ps.Subscribe("topic", SubOptions{BufferSize: 4, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe sub1: %v", err)
	}
	sub2, err := ps.Subscribe("topic", SubOptions{BufferSize: 4, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe sub2: %v", err)
	}

	payload := map[string]any{"count": 1}
	meta := map[string]string{"source": "test"}
	msg := Message{ID: "m1", Data: payload, Meta: meta}
	if err := ps.Publish("topic", msg); err != nil {
		t.Fatalf("publish: %v", err)
	}

	// Mutate original payload/meta after publish.
	payload["count"] = 99
	meta["source"] = "mutated"

	var got1 Message
	select {
	case got1 = <-sub1.C():
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("sub1 timeout")
	}

	data1, ok := got1.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", got1.Data)
	}
	if data1["count"] != 1 {
		t.Fatalf("expected original payload, got %v", data1["count"])
	}
	if got1.Meta["source"] != "test" {
		t.Fatalf("expected original meta, got %v", got1.Meta["source"])
	}

	// Mutate sub1 payload; sub2 should be unaffected.
	data1["count"] = 2

	var got2 Message
	select {
	case got2 = <-sub2.C():
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("sub2 timeout")
	}

	data2, ok := got2.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", got2.Data)
	}
	if data2["count"] != 1 {
		t.Fatalf("expected isolated payload, got %v", data2["count"])
	}
	if got2.Meta["source"] != "test" {
		t.Fatalf("expected isolated meta, got %v", got2.Meta["source"])
	}
}
