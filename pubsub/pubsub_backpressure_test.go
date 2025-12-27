package pubsub

import (
	"testing"
	"time"
)

func TestBackpressure_DropOldest(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.Subscribe("t", SubOptions{BufferSize: 2, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	// fill buffer and overflow
	_ = ps.Publish("t", Message{ID: "m1"})
	_ = ps.Publish("t", Message{ID: "m2"})
	_ = ps.Publish("t", Message{ID: "m3"}) // should evict m1

	got1 := <-sub.C()
	got2 := <-sub.C()

	if got1.ID != "m2" || got2.ID != "m3" {
		t.Fatalf("expected [m2,m3], got [%s,%s]", got1.ID, got2.ID)
	}
}

func TestBackpressure_DropNewest(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.Subscribe("t", SubOptions{BufferSize: 1, Policy: DropNewest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	_ = ps.Publish("t", Message{ID: "m1"})
	_ = ps.Publish("t", Message{ID: "m2"}) // should be dropped

	got := <-sub.C()
	if got.ID != "m1" {
		t.Fatalf("expected m1, got %s", got.ID)
	}
}

func TestBackpressure_BlockWithTimeout(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.Subscribe("t", SubOptions{BufferSize: 1, Policy: BlockWithTimeout, BlockTimeout: 30 * time.Millisecond})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	_ = ps.Publish("t", Message{ID: "m1"})
	// buffer full, publish should not block too long; it may drop
	start := time.Now()
	_ = ps.Publish("t", Message{ID: "m2"})
	if time.Since(start) > 120*time.Millisecond {
		t.Fatalf("publish blocked too long")
	}

	<-sub.C() // drain m1
}

func TestBackpressure_CloseSubscriber(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.Subscribe("t", SubOptions{BufferSize: 1, Policy: CloseSubscriber})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	_ = ps.Publish("t", Message{ID: "m1"})
	_ = ps.Publish("t", Message{ID: "m2"}) // should close subscriber if full

	// Drain first
	_, _ = <-sub.C()

	// Channel should eventually be closed
	select {
	case _, ok := <-sub.C():
		if ok {
			// could still have a buffered value; try next read
			_, ok2 := <-sub.C()
			if ok2 {
				t.Fatalf("expected closed")
			}
		}
	case <-time.After(200 * time.Millisecond):
		// some timings might not close immediately; this should be rare.
		t.Fatalf("expected subscriber to close")
	}
}
