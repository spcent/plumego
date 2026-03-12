package pubsub

import (
	"sync"
	"testing"
	"time"
)

func TestPubSub_ConcurrentPublish_NoPanic(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.Subscribe("t", SubOptions{BufferSize: 64, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Cancel()

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = ps.Publish("t", Message{ID: "x"})
			}
		}()
	}
	wg.Wait()

	// Ensure we can still receive without deadlock
	timeout := time.After(200 * time.Millisecond)
	for {
		select {
		case <-sub.C():
			return
		case <-timeout:
			t.Fatalf("no message received")
		}
	}
}
