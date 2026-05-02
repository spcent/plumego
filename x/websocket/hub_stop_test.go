package websocket

import (
	"errors"
	"testing"
	"time"
)

func TestHubStopBlocksNewJoins(t *testing.T) {
	hub := mustNewHubConfig(t, HubConfig{WorkerCount: 1, JobQueueSize: 4})
	hub.Stop()

	conn := newMockConn()
	defer conn.Close()

	err := hub.TryJoin("room", conn)
	if !errors.Is(err, ErrHubStopped) {
		t.Fatalf("expected ErrHubStopped, got %v", err)
	}

	if got := hub.Metrics().RejectedTotal; got != 1 {
		t.Fatalf("expected RejectedTotal=1, got %d", got)
	}
}

func TestHubCanJoinAfterStop(t *testing.T) {
	hub := mustNewHubConfig(t, HubConfig{WorkerCount: 1, JobQueueSize: 4})
	hub.Stop()

	err := hub.CanJoin("room")
	if !errors.Is(err, ErrHubStopped) {
		t.Fatalf("expected ErrHubStopped, got %v", err)
	}
}

func TestHubTryJoinAfterStopDoesNotRegister(t *testing.T) {
	hub := mustNewHubConfig(t, HubConfig{WorkerCount: 1, JobQueueSize: 4})
	hub.Stop()

	conn := newMockConn()
	defer conn.Close()

	if err := hub.TryJoin("room", conn); !errors.Is(err, ErrHubStopped) {
		t.Fatalf("expected ErrHubStopped, got %v", err)
	}

	if got := hub.GetRoomRegistrationCount(); got != 0 {
		t.Fatalf("expected no registrations after stop, got %d", got)
	}
}

func TestHubStopReturnsWhenBlockingSendQueueIsFull(t *testing.T) {
	hub := mustNewHubConfig(t, HubConfig{WorkerCount: 1, JobQueueSize: 1})

	conn := &Conn{
		sendQueue:    make(chan Outbound, 1),
		closeC:       make(chan struct{}),
		sendBehavior: SendBlock,
	}
	conn.sendQueue <- Outbound{Op: OpcodeText, Data: []byte("full")}
	mustTryJoin(t, hub, "room", conn)

	hub.BroadcastRoom("room", OpcodeText, []byte("blocked"))
	time.Sleep(25 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		hub.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("hub Stop blocked on full connection send queue")
	}
}

func TestHubStopConcurrentBroadcast(t *testing.T) {
	hub := mustNewHubConfig(t, HubConfig{WorkerCount: 1, JobQueueSize: 16})
	conn := newMockConn()
	defer conn.Close()
	mustTryJoin(t, hub, "room", conn)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			_ = hub.TryBroadcastRoom("room", OpcodeText, []byte("hello"))
		}
	}()

	hub.Stop()
	<-done

	if result := hub.TryBroadcastRoom("room", OpcodeText, []byte("after")); result.Enqueued != 0 {
		t.Fatalf("stopped hub reported enqueued jobs: %+v", result)
	}
}
