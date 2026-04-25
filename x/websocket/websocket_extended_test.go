package websocket

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestComputeAcceptKey(t *testing.T) {
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	expected := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	result := computeAcceptKey(key)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestHeaderContains(t *testing.T) {
	tests := []struct {
		name     string
		header   http.Header
		key      string
		val      string
		expected bool
	}{
		{
			name:     "exact match",
			header:   http.Header{"Connection": []string{"Upgrade"}},
			key:      "Connection",
			val:      "Upgrade",
			expected: true,
		},
		{
			name:     "case insensitive",
			header:   http.Header{"Connection": []string{"upgrade"}},
			key:      "Connection",
			val:      "Upgrade",
			expected: true,
		},
		{
			name:     "multiple values",
			header:   http.Header{"Connection": []string{"keep-alive, Upgrade"}},
			key:      "Connection",
			val:      "Upgrade",
			expected: true,
		},
		{
			name:     "not found",
			header:   http.Header{"Connection": []string{"keep-alive"}},
			key:      "Connection",
			val:      "Upgrade",
			expected: false,
		},
		{
			name:     "empty header",
			header:   http.Header{},
			key:      "Connection",
			val:      "Upgrade",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := headerContains(tt.header, tt.key, tt.val)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestServeWSWithAuth_MethodNotAllowed(t *testing.T) {
	hub := NewHub(2, 10)
	defer hub.Stop()
	auth := NewSimpleRoomAuth([]byte("secret"))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/ws", nil)

	ServeWSWithAuth(w, r, hub, auth, 10, 5*time.Second, SendDrop)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestServeWSWithAuth_BadRequest(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*http.Request)
		expect int
	}{
		{
			name: "missing connection header",
			setup: func(r *http.Request) {
				r.Header.Set("Upgrade", "websocket")
				r.Header.Set("Sec-WebSocket-Key", "test-key")
			},
			expect: http.StatusBadRequest,
		},
		{
			name: "missing upgrade header",
			setup: func(r *http.Request) {
				r.Header.Set("Connection", "Upgrade")
				r.Header.Set("Sec-WebSocket-Key", "test-key")
			},
			expect: http.StatusBadRequest,
		},
		{
			name: "missing sec-websocket-key",
			setup: func(r *http.Request) {
				r.Header.Set("Connection", "Upgrade")
				r.Header.Set("Upgrade", "websocket")
			},
			expect: http.StatusBadRequest,
		},
	}

	hub := NewHub(2, 10)
	defer hub.Stop()
	auth := NewSimpleRoomAuth([]byte("secret"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/ws", nil)
			tt.setup(r)

			ServeWSWithAuth(w, r, hub, auth, 10, 5*time.Second, SendDrop)

			if w.Code != tt.expect {
				t.Errorf("expected status %d, got %d", tt.expect, w.Code)
			}
		})
	}
}

func TestServeWSWithAuth_BadRoomPassword(t *testing.T) {
	hub := NewHub(2, 10)
	defer hub.Stop()
	auth := NewSimpleRoomAuth([]byte("secret"))
	// Set a room password first
	if err := auth.SetRoomPassword("test", "correct"); err != nil {
		t.Fatalf("SetRoomPassword: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/ws?room=test&room_password=wrong", nil)
	r.Header.Set("Connection", "Upgrade")
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==") // Valid WebSocket Key

	ServeWSWithAuth(w, r, hub, auth, 10, 5*time.Second, SendDrop)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestStreamReaderRead(t *testing.T) {
	sr := &streamReader{
		parent: &Conn{},
		op:     0,
	}

	sr.done = true
	buf := make([]byte, 10)
	n, err := sr.Read(buf)
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 bytes, got %d", n)
	}
}

func TestStreamReaderClose(t *testing.T) {
	sr := &streamReader{}

	err := sr.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = sr.Close()
	if err != nil {
		t.Errorf("unexpected error on second close: %v", err)
	}
}

func TestWriteMessageWithTimeout(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan Outbound, 1),
		closeC:       make(chan struct{}),
		sendTimeout:  10 * time.Millisecond,
		sendBehavior: SendBlock,
	}

	c.sendQueue <- Outbound{Op: OpcodeText, Data: []byte("full")}

	err := c.WriteMessage(OpcodeText, []byte("test"))
	assertErrorIsOrContains(t, err, context.DeadlineExceeded, "timeout", "deadline")
}

func TestWriteMessageDrop(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan Outbound, 1),
		closeC:       make(chan struct{}),
		sendBehavior: SendDrop,
	}

	c.sendQueue <- Outbound{Op: OpcodeText, Data: []byte("full")}

	err := c.WriteMessage(OpcodeText, []byte("test"))
	assertErrorIsOrContains(t, err, ErrQueueFull, "queue full")
}

func TestWriteMessageClose(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan Outbound, 1),
		closeC:       make(chan struct{}),
		sendBehavior: SendClose,
	}

	c.sendQueue <- Outbound{Op: OpcodeText, Data: []byte("full")}

	err := c.WriteMessage(OpcodeText, []byte("test"))
	assertErrorContains(t, err, "closed")
}

func TestWriteText(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan Outbound, 1),
		closeC:       make(chan struct{}),
		sendBehavior: SendDrop,
	}

	err := c.WriteText("hello")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case out := <-c.sendQueue:
		if out.Op != OpcodeText {
			t.Errorf("expected OpcodeText, got %d", out.Op)
		}
		if string(out.Data) != "hello" {
			t.Errorf("expected 'hello', got %s", string(out.Data))
		}
	default:
		t.Error("message not queued")
	}
}

func TestWriteBinary(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan Outbound, 1),
		closeC:       make(chan struct{}),
		sendBehavior: SendDrop,
	}

	data := []byte{0x01, 0x02, 0x03}
	err := c.WriteBinary(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case out := <-c.sendQueue:
		if out.Op != OpcodeBinary {
			t.Errorf("expected OpcodeBinary, got %d", out.Op)
		}
		if !bytes.Equal(out.Data, data) {
			t.Errorf("expected %v, got %v", data, out.Data)
		}
	default:
		t.Error("message not queued")
	}
}

func TestWriteJSONMarshalError(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan Outbound, 1),
		closeC:       make(chan struct{}),
		sendBehavior: SendDrop,
	}

	ch := make(chan int)
	err := c.WriteJSON(ch)
	if err == nil {
		t.Error("expected marshal error")
	}
}

func TestWriteMessageClosed(t *testing.T) {
	c := &Conn{
		closeC: make(chan struct{}),
	}
	c.Close()

	err := c.WriteMessage(OpcodeText, []byte("test"))
	assertErrorIsOrContains(t, err, ErrConnClosed, "closed")
}

func TestWriteMessageUnknownBehavior(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan Outbound, 1),
		closeC:       make(chan struct{}),
		sendBehavior: 99,
	}

	c.WriteMessage(OpcodeText, []byte("test1"))
	err := c.WriteMessage(OpcodeText, []byte("test2"))
	assertErrorContains(t, err, "unknown")
}

func TestWriterPumpFragmentation(t *testing.T) {
	c := &Conn{
		sendQueue: make(chan Outbound, 1),
		closeC:    make(chan struct{}),
	}
	c.pingPeriod.Store(int64(1 * time.Second))

	largeData := bytes.Repeat([]byte("x"), maxFragmentSize*2+100)
	c.sendQueue <- Outbound{Op: OpcodeBinary, Data: largeData}
	c.Close()

	select {
	case out := <-c.sendQueue:
		if len(out.Data) != len(largeData) {
			t.Errorf("expected data length %d, got %d", len(largeData), len(out.Data))
		}
	default:
		t.Error("message not found in queue")
	}
}

func TestPongMonitor(t *testing.T) {
	c := &Conn{
		closeC: make(chan struct{}),
	}
	c.lastPong.Store(time.Now().Add(-2 * time.Second).UnixNano())
	c.pingPeriod.Store(int64(10 * time.Millisecond))
	c.pongWait.Store(int64(5 * time.Millisecond))

	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Duration(c.pongWait.Load()) / 3)
		defer ticker.Stop()

		for {
			select {
			case <-c.closeC:
				done <- true
				return
			case <-ticker.C:
				last := time.Unix(0, c.lastPong.Load())
				if time.Since(last) > time.Duration(c.pongWait.Load()) {
					c.Close()
				}
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("pongMonitor did not close connection")
	}
}

func TestReadMessageStreamClosed(t *testing.T) {
	c := &Conn{
		closeC: make(chan struct{}),
	}
	c.Close()

	_, _, err := c.ReadMessageStream()
	assertErrorIsOrContains(t, err, ErrConnClosed, "closed")
}

func TestStreamReaderWithBuffer(t *testing.T) {
	sr := &streamReader{
		parent: &Conn{},
		op:     0,
	}

	sr.buf.WriteString("testdata")
	sr.done = true

	buf := make([]byte, 4)
	n, err := sr.Read(buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes, got %d", n)
	}
	if string(buf) != "test" {
		t.Errorf("expected 'test', got %s", string(buf))
	}

	n, err = sr.Read(buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes, got %d", n)
	}
	if string(buf[:n]) != "data" {
		t.Errorf("expected 'data', got %s", string(buf[:n]))
	}

	n, err = sr.Read(buf)
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 bytes, got %d", n)
	}
}

func TestStreamReaderReadError(t *testing.T) {
	sr := &streamReader{
		parent:  &Conn{},
		op:      0,
		readErr: errors.New("test error"),
		done:    true,
	}

	buf := make([]byte, 10)
	_, err := sr.Read(buf)
	if err == nil || err.Error() != "test error" {
		t.Errorf("expected 'test error', got %v", err)
	}
}

func TestServeWSWithAuth_HijackFailure(t *testing.T) {
	hub := NewHub(2, 10)
	defer hub.Stop()
	auth := NewSimpleRoomAuth([]byte("secret"))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/ws", nil)
	r.Header.Set("Connection", "Upgrade")
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==") // Valid WebSocket Key

	ServeWSWithAuth(w, r, hub, auth, 10, 5*time.Second, SendDrop)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	assertWebSocketError(t, w, http.StatusInternalServerError, codeWebSocketHijackUnsupported, "websocket hijack unsupported")
}

func TestConnWriteMessageWithBehavior(t *testing.T) {
	tests := []struct {
		name      string
		behavior  SendBehavior
		timeout   time.Duration
		setupFunc func(*Conn)
		expectErr bool
		errMsg    string
	}{
		{
			name:     "block with timeout success",
			behavior: SendBlock,
			timeout:  100 * time.Millisecond,
			setupFunc: func(c *Conn) {
			},
			expectErr: false,
		},
		{
			name:     "block with timeout failure",
			behavior: SendBlock,
			timeout:  10 * time.Millisecond,
			setupFunc: func(c *Conn) {
				c.sendQueue <- Outbound{Op: OpcodeText, Data: []byte("full")}
			},
			expectErr: true,
			errMsg:    "deadline", // context.DeadlineExceeded
		},
		{
			name:     "drop when full",
			behavior: SendDrop,
			timeout:  0,
			setupFunc: func(c *Conn) {
				c.sendQueue <- Outbound{Op: OpcodeText, Data: []byte("full")}
			},
			expectErr: true,
			errMsg:    "queue full", // ErrQueueFull
		},
		{
			name:     "close when full",
			behavior: SendClose,
			timeout:  0,
			setupFunc: func(c *Conn) {
				c.sendQueue <- Outbound{Op: OpcodeText, Data: []byte("full")}
			},
			expectErr: true,
			errMsg:    "closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				sendQueue:    make(chan Outbound, 1),
				closeC:       make(chan struct{}),
				sendTimeout:  tt.timeout,
				sendBehavior: tt.behavior,
			}

			tt.setupFunc(c)

			err := c.WriteMessage(OpcodeText, []byte("test"))

			if tt.expectErr {
				assertErrorContains(t, err, tt.errMsg)
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestServeWSWithAuth_TokenInQuery(t *testing.T) {
	hub := NewHub(2, 10)
	defer hub.Stop()
	auth := NewSimpleRoomAuth([]byte("secret"))
	w := &testHijackWriter{
		httptest.NewRecorder(),
	}

	r := httptest.NewRequest("GET", "/ws?room=test&token=valid", nil)
	r.Header.Set("Connection", "Upgrade")
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Sec-WebSocket-Key", "test-key")

	ServeWSWithAuth(w, r, hub, auth, 10, 5*time.Second, SendDrop)

	if w.Code != http.StatusInternalServerError {
		t.Logf("Got status %d, body: %s", w.Code, w.Body.String())
	}
}

type testHijackWriter struct {
	*httptest.ResponseRecorder
}

func (w *testHijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack not supported")
}

func TestConnWriteMessageBlockWithZeroTimeout(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan Outbound, 1),
		closeC:       make(chan struct{}),
		sendTimeout:  0,
		sendBehavior: SendBlock,
	}

	c.sendQueue <- Outbound{Op: OpcodeText, Data: []byte("full")}

	go func() {
		time.Sleep(50 * time.Millisecond)
		<-c.sendQueue
	}()

	done := make(chan error, 1)
	go func() {
		err := c.WriteMessage(OpcodeText, []byte("test"))
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("write blocked indefinitely")
	}
}

func TestConnWriteMessageBlockWithClose(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan Outbound, 1),
		closeC:       make(chan struct{}),
		sendTimeout:  100 * time.Millisecond,
		sendBehavior: SendBlock,
	}

	c.sendQueue <- Outbound{Op: OpcodeText, Data: []byte("full")}

	go func() {
		time.Sleep(20 * time.Millisecond)
		c.Close()
	}()

	err := c.WriteMessage(OpcodeText, []byte("test"))
	assertErrorContains(t, err, "closed")
}
