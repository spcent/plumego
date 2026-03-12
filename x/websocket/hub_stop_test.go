package websocket

import (
	"errors"
	"testing"
)

func TestHubStopBlocksNewJoins(t *testing.T) {
	hub := NewHub(1, 4)
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
	hub := NewHub(1, 4)
	hub.Stop()

	err := hub.CanJoin("room")
	if !errors.Is(err, ErrHubStopped) {
		t.Fatalf("expected ErrHubStopped, got %v", err)
	}
}

func TestHubJoinNoopAfterStop(t *testing.T) {
	hub := NewHub(1, 4)
	hub.Stop()

	conn := newMockConn()
	defer conn.Close()

	hub.Join("room", conn)

	if got := hub.GetTotalCount(); got != 0 {
		t.Fatalf("expected no registrations after stop, got %d", got)
	}
}
