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
