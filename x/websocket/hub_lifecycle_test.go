package websocket

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"strings"
	"testing"
	"time"
)

// --- Stop idempotency ---

func TestHub_Stop_Idempotent(t *testing.T) {
	hub := mustHub(t, 1, 4)
	hub.Stop()
	hub.Stop() // must not panic
}

// --- BroadcastRoom/BroadcastAll after Stop ---

func TestHub_BroadcastRoom_AfterStop_NoOp(t *testing.T) {
	hub := mustHub(t, 1, 4)
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
	hub := mustHub(t, 1, 4)
	hub.Stop()
	hub.BroadcastAll(OpcodeText, []byte("world")) // must not panic
}

func TestHub_TryBroadcastRoomReportsStopped(t *testing.T) {
	hub := mustHub(t, 1, 4)
	hub.Stop()

	result, err := hub.TryBroadcastRoom("room", OpcodeText, []byte("hello"))
	if !errors.Is(err, ErrHubStopped) {
		t.Fatalf("TryBroadcastRoom error = %v, want ErrHubStopped", err)
	}
	if result != (BroadcastResult{}) {
		t.Fatalf("result = %+v, want zero", result)
	}
}

func TestHub_TryBroadcastRoomEmptyRoom(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	result, err := hub.TryBroadcastRoom("missing", OpcodeText, []byte("hello"))
	if err != nil {
		t.Fatalf("TryBroadcastRoom error = %v", err)
	}
	if result != (BroadcastResult{}) {
		t.Fatalf("result = %+v, want zero", result)
	}
}

func TestHub_TryBroadcastRoomRejectsInvalidRoom(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	result, err := hub.TryBroadcastRoom("bad/room", OpcodeText, []byte("hello"))
	if !errors.Is(err, ErrInvalidRoomName) {
		t.Fatalf("TryBroadcastRoom error = %v, want ErrInvalidRoomName", err)
	}
	if result != (BroadcastResult{}) {
		t.Fatalf("TryBroadcastRoom result = %+v, want zero result", result)
	}
}

func TestHub_TryBroadcastRoomRejectsInvalidOpcode(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	result, err := hub.TryBroadcastRoom("room", opcodeClose, []byte("bad"))
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("TryBroadcastRoom error = %v, want ErrProtocolError", err)
	}
	if result != (BroadcastResult{}) {
		t.Fatalf("TryBroadcastRoom result = %+v, want zero result", result)
	}
}

func TestHub_TryBroadcastAllRejectsInvalidOpcode(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	result, err := hub.TryBroadcastAll(opcodePing, []byte("bad"))
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("TryBroadcastAll error = %v, want ErrProtocolError", err)
	}
	if result != (BroadcastResult{}) {
		t.Fatalf("TryBroadcastAll result = %+v, want zero result", result)
	}
}

func TestHub_TryBroadcastRoomReportsSent(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()
	conn := newMockConn()
	defer conn.Close()
	if err := hub.TryJoin("room", conn); err != nil {
		t.Fatalf("TryJoin: %v", err)
	}

	result, err := hub.TryBroadcastRoom("room", OpcodeText, []byte("hello"))
	if err != nil {
		t.Fatalf("TryBroadcastRoom error = %v", err)
	}
	if result.Sent != 1 || result.Dropped != 0 {
		t.Fatalf("result = %+v, want sent=1 dropped=0", result)
	}
}

func TestHubConnListPoolDropsLargeSlices(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	large := make([]*Conn, 0, maxPooledConnListCap+1)
	hub.putConnList(&large)
	if large != nil {
		t.Fatalf("large conn list cap = %d, want dropped slice", cap(large))
	}

	small := make([]*Conn, 1, maxPooledConnListCap)
	hub.putConnList(&small)
	if small == nil || len(small) != 0 {
		t.Fatalf("small conn list = %#v, want retained empty slice", small)
	}
}

func TestHub_DispatchJobsReportsDropped(t *testing.T) {
	hub := &Hub{
		jobQueue: make(chan hubJob, 1),
		quit:     make(chan struct{}),
	}
	hub.jobQueue <- hubJob{}
	conn := &Conn{closeC: make(chan struct{})}
	defer conn.Close()

	result := hub.dispatchJobs([]*Conn{conn}, OpcodeText, []byte("hello"), "test")
	if result.Sent != 0 || result.Dropped != 1 {
		t.Fatalf("result = %+v, want sent=0 dropped=1", result)
	}
	if got := hub.Metrics().BroadcastDropped; got != 1 {
		t.Fatalf("BroadcastDropped = %d, want 1", got)
	}
}

func TestHub_DispatchJobsDoesNotCopyPayloadWhenQueueFull(t *testing.T) {
	hub := &Hub{
		jobQueue: make(chan hubJob, 1),
		quit:     make(chan struct{}),
	}
	hub.jobQueue <- hubJob{}
	conn := &Conn{closeC: make(chan struct{})}
	defer conn.Close()
	payload := make([]byte, 1<<20)
	conns := []*Conn{conn}

	result := hub.dispatchJobs(conns, OpcodeText, payload, "test")
	if result.Sent != 0 || result.Dropped != 1 {
		t.Fatalf("result = %+v, want sent=0 dropped=1", result)
	}
	allocs := testing.AllocsPerRun(100, func() {
		_ = hub.dispatchJobs(conns, OpcodeText, payload, "test")
	})
	if allocs != 0 {
		t.Fatalf("full broadcast queue allocations = %v, want 0", allocs)
	}
}

func TestHub_SecurityEventHandlerDoesNotBlockProducer(t *testing.T) {
	release := make(chan struct{})
	hub := mustHubWithConfig(t, HubConfig{
		WorkerCount:          1,
		JobQueueSize:         4,
		EnableSecurityEvents: true,
		SecurityEventHandler: func(SecurityEvent) {
			<-release
		},
	})
	defer func() {
		close(release)
		hub.Stop()
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < cap(hub.securityEvents)+10; i++ {
			hub.recordSecurityEvent("test_event", map[string]any{"i": i}, "warning")
		}
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("recordSecurityEvent blocked on SecurityEventHandler")
	}
}

func TestHub_StopDoesNotWaitForBlockedSecurityEventHandler(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	hub := mustHubWithConfig(t, HubConfig{
		WorkerCount:          1,
		JobQueueSize:         4,
		EnableSecurityEvents: true,
		SecurityEventHandler: func(SecurityEvent) {
			close(started)
			<-release
		},
	})
	defer close(release)

	hub.recordSecurityEvent("test_event", map[string]any{"blocked": true}, "warning")
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for security event handler to start")
	}

	done := make(chan struct{})
	go func() {
		hub.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop waited for blocked SecurityEventHandler")
	}
}

func TestHubHandleSecurityEventDropsHandlerEventAfterStop(t *testing.T) {
	hub := &Hub{
		handlerEvents: make(chan SecurityEvent, 1),
		logger:        log.New(io.Discard, "", 0),
	}
	hub.stopped.Store(true)

	hub.handleSecurityEvent(SecurityEvent{Type: "after_stop"})
	if got := len(hub.handlerEvents); got != 0 {
		t.Fatalf("handlerEvents len = %d, want 0 after stop", got)
	}
}

func TestHubSecurityEventHandlerPanicRecovered(t *testing.T) {
	var logBuf bytes.Buffer
	handled := make(chan string, 1)
	hub := mustHubWithConfig(t, HubConfig{
		WorkerCount:          1,
		JobQueueSize:         4,
		EnableSecurityEvents: true,
		Logger:               log.New(&logBuf, "", 0),
		SecurityEventHandler: func(event SecurityEvent) {
			if event.Type == "panic_event" {
				panic("handler failed")
			}
			handled <- event.Type
		},
	})
	defer hub.Stop()

	hub.recordSecurityEvent("panic_event", nil, "warning")
	hub.recordSecurityEvent("normal_event", nil, "warning")

	select {
	case got := <-handled:
		if got != "normal_event" {
			t.Fatalf("handled event = %q, want normal_event", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler after recovered panic")
	}
	if !strings.Contains(logBuf.String(), "security event handler panic recovered") {
		t.Fatalf("panic recovery log missing: %q", logBuf.String())
	}
}

// --- TryJoin capacity errors ---

func TestHub_TryJoin_HubFull(t *testing.T) {
	hub := mustHubWithConfig(t, HubConfig{
		MaxRoomRegistrations: 1,
		WorkerCount:          1,
		JobQueueSize:         4,
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
	hub := mustHubWithConfig(t, HubConfig{
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

func TestHub_TryJoinRejectsInvalidRoomAndNilConn(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	conn := newMockConn()
	defer conn.Close()

	if err := hub.TryJoin("bad/room", conn); !errors.Is(err, ErrInvalidRoomName) {
		t.Fatalf("TryJoin invalid room error = %v, want ErrInvalidRoomName", err)
	}
	if err := hub.TryJoin("room", nil); !errors.Is(err, ErrNilNetConn) {
		t.Fatalf("TryJoin nil conn error = %v, want ErrNilNetConn", err)
	}
	if got := hub.Metrics().RejectedTotal; got != 2 {
		t.Fatalf("RejectedTotal = %d, want 2", got)
	}
}

// --- CanJoin capacity errors ---

func TestHub_CanJoin_HubFull(t *testing.T) {
	hub := mustHubWithConfig(t, HubConfig{
		MaxRoomRegistrations: 1,
		WorkerCount:          1,
		JobQueueSize:         4,
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

func TestHub_CanJoinRejectsInvalidRoom(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	if err := hub.CanJoin("bad/room"); !errors.Is(err, ErrInvalidRoomName) {
		t.Fatalf("CanJoin invalid room error = %v, want ErrInvalidRoomName", err)
	}
}

func TestNewHubWithConfigERejectsNegativeLimits(t *testing.T) {
	tests := []struct {
		name string
		cfg  HubConfig
	}{
		{
			name: "max room registrations",
			cfg: HubConfig{
				WorkerCount:          1,
				JobQueueSize:         4,
				MaxRoomRegistrations: -1,
			},
		},
		{
			name: "max room connections",
			cfg: HubConfig{
				WorkerCount:        1,
				JobQueueSize:       4,
				MaxRoomConnections: -1,
			},
		},
		{
			name: "max connection rate",
			cfg: HubConfig{
				WorkerCount:       1,
				JobQueueSize:      4,
				MaxConnectionRate: -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub, err := NewHubWithConfigE(tt.cfg)
			if err == nil {
				if hub != nil {
					hub.Stop()
				}
				t.Fatal("expected error")
			}
			if !errors.Is(err, ErrInvalidHubConfig) {
				t.Fatalf("NewHubWithConfigE error = %v, want ErrInvalidHubConfig", err)
			}
		})
	}
}

// --- RangeConns ---

func TestHub_RangeConns_VisitsAll(t *testing.T) {
	hub := mustHub(t, 1, 8)
	defer hub.Stop()

	c1 := newMockConn()
	c2 := newMockConn()
	defer c1.Close()
	defer c2.Close()

	mustJoin(t, hub, "room", c1)
	mustJoin(t, hub, "room", c2)

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
	hub := mustHub(t, 1, 8)
	defer hub.Stop()

	for i := 0; i < 3; i++ {
		c := newMockConn()
		defer c.Close()
		mustJoin(t, hub, "room", c)
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
	hub := mustHub(t, 1, 4)
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
	hub := mustHub(t, 1, 4)
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	if err := hub.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown empty hub: %v", err)
	}
}

func TestHub_Shutdown_NilContext(t *testing.T) {
	hub := mustHub(t, 1, 4)

	if err := hub.Shutdown(nil); err != nil {
		t.Errorf("Shutdown nil context: %v", err)
	}
}

func TestHub_Shutdown_WithConnections(t *testing.T) {
	hub := mustHub(t, 2, 8)

	c1 := newMockConn()
	c2 := newMockConn()
	mustJoin(t, hub, "r", c1)
	mustJoin(t, hub, "r", c2)

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

func TestHub_Shutdown_ClearsRoomsAndSendsCloseFrames(t *testing.T) {
	hub := mustHub(t, 2, 8)
	c1, raw1 := newCloseFrameTestConn()
	c2, raw2 := newCloseFrameTestConn()

	mustJoin(t, hub, "r1", c1)
	mustJoin(t, hub, "r2", c1)
	mustJoin(t, hub, "r2", c2)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	if err := hub.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	if got := hub.GetRoomRegistrationCount(); got != 0 {
		t.Fatalf("room registrations after Shutdown = %d, want 0", got)
	}
	if got := len(hub.GetRooms()); got != 0 {
		t.Fatalf("rooms after Shutdown = %d, want 0", got)
	}
	if !c1.IsClosed() || !c2.IsClosed() {
		t.Fatal("expected registered connections to be closed")
	}
	assertCloseFrame(t, raw1, CloseGoingAway)
	assertCloseFrame(t, raw2, CloseGoingAway)
}

func TestHub_Shutdown_ContextCancel_ReturnsCtxErr(t *testing.T) {
	hub := mustHub(t, 1, 4)

	// Add many connections to slow down Shutdown iteration.
	for i := 0; i < 5; i++ {
		c := newMockConn()
		mustJoin(t, hub, "r", c)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // already cancelled

	err := hub.Shutdown(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled or nil, got %v", err)
	}
	if got := hub.GetRoomRegistrationCount(); got != 0 {
		t.Fatalf("room registrations after canceled Shutdown = %d, want 0", got)
	}
}

func newCloseFrameTestConn() (*Conn, *failingWriteConn) {
	rawConn := &failingWriteConn{}
	c := &Conn{
		conn:        rawConn,
		bw:          bufio.NewWriterSize(rawConn, defaultBufSize),
		closeC:      make(chan struct{}),
		sendTimeout: time.Second,
	}
	return c, rawConn
}

func assertCloseFrame(t *testing.T, rawConn *failingWriteConn, wantCode uint16) {
	t.Helper()

	written := rawConn.written.Bytes()
	if len(written) < 4 {
		t.Fatalf("written frame length = %d, want at least 4", len(written))
	}
	if got := written[0] & 0x0F; got != opcodeClose {
		t.Fatalf("opcode = %d, want close", got)
	}
	payloadLen := int(written[1] & 0x7F)
	if payloadLen < 2 || len(written) < 2+payloadLen {
		t.Fatalf("invalid close frame payload length %d for %d bytes", payloadLen, len(written))
	}
	if got := binary.BigEndian.Uint16(written[2:4]); got != wantCode {
		t.Fatalf("close code = %d, want %d", got, wantCode)
	}
}

// --- Metrics ---

func TestHub_Metrics_InitialState(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	m := hub.Metrics()
	if m.RoomRegistrations != 0 {
		t.Errorf("RoomRegistrations = %d, want 0", m.RoomRegistrations)
	}
	if m.RejectedTotal != 0 {
		t.Errorf("RejectedTotal = %d, want 0", m.RejectedTotal)
	}
}

func TestHub_DispatchJobsCountsDroppedAfterStopWithoutMetrics(t *testing.T) {
	hub := mustHubWithConfig(t, HubConfig{
		WorkerCount:  1,
		JobQueueSize: 1,
	})
	hub.Stop()

	c := newMockConn()
	defer c.Close()
	hub.dispatchJobs([]*Conn{c}, OpcodeText, []byte("x"), "test")

	if got := hub.Metrics().BroadcastDropped; got != 1 {
		t.Fatalf("BroadcastDropped = %d, want 1", got)
	}
}

func TestHub_StopDoesNotBlockOnFullSendQueue(t *testing.T) {
	hub := mustHubWithConfig(t, HubConfig{
		WorkerCount:  1,
		JobQueueSize: 1,
	})

	conn := &Conn{
		sendQueue:    make(chan outbound, 1),
		sendBehavior: SendBlock,
		closeC:       make(chan struct{}),
	}
	conn.sendQueue <- outbound{Op: OpcodeText, Data: []byte("queued")}
	defer conn.Close()

	hub.jobQueue <- hubJob{conn: conn, op: OpcodeText, data: []byte("blocked")}

	done := make(chan struct{})
	go func() {
		hub.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop blocked on full connection send queue")
	}
}

func TestHubSecurityEventHandlerReceivesEvents(t *testing.T) {
	events := make(chan SecurityEvent, 1)
	hub := mustHubWithConfig(t, HubConfig{
		WorkerCount:          1,
		JobQueueSize:         1,
		MaxRoomRegistrations: 1,
		EnableSecurityEvents: true,
		SecurityEventHandler: func(event SecurityEvent) {
			events <- event
		},
	})
	defer hub.Stop()

	c1 := newMockConn()
	defer c1.Close()
	if err := hub.TryJoin("r", c1); err != nil {
		t.Fatalf("TryJoin first connection: %v", err)
	}
	c2 := newMockConn()
	defer c2.Close()
	if err := hub.TryJoin("r", c2); !errors.Is(err, ErrHubFull) {
		t.Fatalf("TryJoin second connection error = %v, want ErrHubFull", err)
	}

	select {
	case event := <-events:
		if event.Type != "hub_full" {
			t.Fatalf("event type = %q, want hub_full", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for security event")
	}
}

func TestHubSecurityEventHandlerCanReenterHub(t *testing.T) {
	events := make(chan SecurityEvent, 1)
	var hub *Hub
	hub = mustHubWithConfig(t, HubConfig{
		WorkerCount:          1,
		JobQueueSize:         1,
		MaxRoomRegistrations: 1,
		EnableSecurityEvents: true,
		SecurityEventHandler: func(event SecurityEvent) {
			_ = hub.GetRoomCount("r")
			events <- event
		},
	})
	defer hub.Stop()

	c1 := newMockConn()
	defer c1.Close()
	if err := hub.TryJoin("r", c1); err != nil {
		t.Fatalf("TryJoin first connection: %v", err)
	}
	c2 := newMockConn()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- hub.TryJoin("r", c2)
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, ErrHubFull) {
			t.Fatalf("TryJoin second connection error = %v, want ErrHubFull", err)
		}
	case <-time.After(time.Second):
		t.Fatal("TryJoin deadlocked while invoking security event handler")
	}

	select {
	case event := <-events:
		if event.Type != "hub_full" {
			t.Fatalf("event type = %q, want hub_full", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for reentrant security event")
	}
}

func TestHub_GetRooms_Empty(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	rooms := hub.GetRooms()
	if len(rooms) != 0 {
		t.Errorf("expected no rooms, got %v", rooms)
	}
}

func TestHub_Leave_NonMember_NoOp(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	conn := newMockConn()
	defer conn.Close()
	hub.Leave("room", conn) // conn was never joined — must not panic
}

func TestHub_RemoveConn_NotInAnyRoom_NoOp(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	conn := newMockConn()
	defer conn.Close()
	hub.RemoveConn(conn) // must not panic
}
