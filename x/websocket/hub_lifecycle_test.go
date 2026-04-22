package websocket

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- Stop idempotency ---

func TestHub_Stop_Idempotent(t *testing.T) {
	hub := NewHub(1, 4)
	hub.Stop()
	hub.Stop() // must not panic
}

// --- BroadcastRoom/BroadcastAll after Stop ---

func TestHub_BroadcastRoom_AfterStop_NoOp(t *testing.T) {
	hub := NewHub(1, 4)
	conn := newMockConn()
	defer conn.Close()
	if err := hub.TryJoin("room", conn); err != nil {
		t.Fatalf("TryJoin before stop: %v", err)
	}

	hub.Stop()
	// Must not panic; dropped silently when hub is stopped.
	hub.BroadcastRoom("room", OpcodeText, []byte("hello"))
}

func TestHub_BroadcastAll_AfterStop_NoOp(t *testing.T) {
	hub := NewHub(1, 4)
	hub.Stop()
	hub.BroadcastAll(OpcodeText, []byte("world")) // must not panic
}

// --- TryJoin capacity errors ---

func TestHub_TryJoin_HubFull(t *testing.T) {
	hub := NewHubWithConfig(HubConfig{
		MaxConnections: 1,
		WorkerCount:    1,
		JobQueueSize:   4,
	})
	defer hub.Stop()

	c1 := newMockConn()
	defer c1.Close()
	c2 := newMockConn()
	defer c2.Close()

	if err := hub.TryJoin("r", c1); err != nil {
		t.Fatalf("first join: %v", err)
	}
	err := hub.TryJoin("r", c2)
	if !errors.Is(err, ErrHubFull) {
		t.Errorf("expected ErrHubFull, got %v", err)
	}
}

func TestHub_TryJoin_RoomFull(t *testing.T) {
	hub := NewHubWithConfig(HubConfig{
		MaxRoomConnections: 1,
		WorkerCount:        1,
		JobQueueSize:       4,
	})
	defer hub.Stop()

	c1 := newMockConn()
	defer c1.Close()
	c2 := newMockConn()
	defer c2.Close()

	if err := hub.TryJoin("r", c1); err != nil {
		t.Fatalf("first join: %v", err)
	}
	err := hub.TryJoin("r", c2)
	if !errors.Is(err, ErrRoomFull) {
		t.Errorf("expected ErrRoomFull, got %v", err)
	}
}

// --- CanJoin capacity errors ---

func TestHub_CanJoin_HubFull(t *testing.T) {
	hub := NewHubWithConfig(HubConfig{
		MaxConnections: 1,
		WorkerCount:    1,
		JobQueueSize:   4,
	})
	defer hub.Stop()

	c := newMockConn()
	defer c.Close()
	if err := hub.TryJoin("r", c); err != nil {
		t.Fatalf("join: %v", err)
	}
	if err := hub.CanJoin("r"); !errors.Is(err, ErrHubFull) {
		t.Errorf("expected ErrHubFull, got %v", err)
	}
}

// --- RangeConns ---

func TestHub_RangeConns_VisitsAll(t *testing.T) {
	hub := NewHub(1, 8)
	defer hub.Stop()

	c1 := newMockConn()
	c2 := newMockConn()
	defer c1.Close()
	defer c2.Close()

	hub.Join("room", c1)
	hub.Join("room", c2)

	var count int
	hub.RangeConns("room", func(c *Conn) bool {
		count++
		return true
	})
	if count != 2 {
		t.Errorf("RangeConns visited %d conns, want 2", count)
	}
}

func TestHub_RangeConns_EarlyReturn(t *testing.T) {
	hub := NewHub(1, 8)
	defer hub.Stop()

	for i := 0; i < 3; i++ {
		c := newMockConn()
		defer c.Close()
		hub.Join("room", c)
	}

	var count int
	hub.RangeConns("room", func(c *Conn) bool {
		count++
		return false // stop after first
	})
	if count != 1 {
		t.Errorf("RangeConns with early return visited %d, want 1", count)
	}
}

func TestHub_RangeConns_EmptyRoom_NoOp(t *testing.T) {
	hub := NewHub(1, 4)
	defer hub.Stop()

	var called bool
	hub.RangeConns("nonexistent", func(*Conn) bool {
		called = true
		return true
	})
	if called {
		t.Error("RangeConns on empty room should not invoke fn")
	}
}

// --- Shutdown ---

func TestHub_Shutdown_EmptyHub(t *testing.T) {
	hub := NewHub(1, 4)
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	if err := hub.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown empty hub: %v", err)
	}
}

func TestHub_Shutdown_WithConnections(t *testing.T) {
	hub := NewHub(2, 8)

	c1 := newMockConn()
	c2 := newMockConn()
	hub.Join("r", c1)
	hub.Join("r", c2)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	if err := hub.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown: %v", err)
	}

	// After Shutdown the hub must be stopped: TryJoin must return ErrHubStopped.
	c3 := newMockConn()
	defer c3.Close()
	if err := hub.TryJoin("r", c3); !errors.Is(err, ErrHubStopped) {
		t.Errorf("expected ErrHubStopped after Shutdown, got %v", err)
	}
}

func TestHub_Shutdown_ContextCancel_ReturnsCtxErr(t *testing.T) {
	hub := NewHub(1, 4)

	// Add many connections to slow down Shutdown iteration.
	for i := 0; i < 5; i++ {
		c := newMockConn()
		hub.Join("r", c)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // already cancelled

	err := hub.Shutdown(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled or nil, got %v", err)
	}
}

// --- Metrics ---

func TestHub_Metrics_InitialState(t *testing.T) {
	hub := NewHub(1, 4)
	defer hub.Stop()

	m := hub.Metrics()
	if m.ActiveConnections != 0 {
		t.Errorf("ActiveConnections = %d, want 0", m.ActiveConnections)
	}
	if m.RejectedTotal != 0 {
		t.Errorf("RejectedTotal = %d, want 0", m.RejectedTotal)
	}
}

func TestHub_GetRooms_Empty(t *testing.T) {
	hub := NewHub(1, 4)
	defer hub.Stop()

	rooms := hub.GetRooms()
	if len(rooms) != 0 {
		t.Errorf("expected no rooms, got %v", rooms)
	}
}

func TestHub_Leave_NonMember_NoOp(t *testing.T) {
	hub := NewHub(1, 4)
	defer hub.Stop()

	conn := newMockConn()
	defer conn.Close()
	hub.Leave("room", conn) // conn was never joined — must not panic
}

func TestHub_RemoveConn_NotInAnyRoom_NoOp(t *testing.T) {
	hub := NewHub(1, 4)
	defer hub.Stop()

	conn := newMockConn()
	defer conn.Close()
	hub.RemoveConn(conn) // must not panic
}
