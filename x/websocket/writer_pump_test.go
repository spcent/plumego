package websocket

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

type failingWriteConn struct {
	writeErr           error
	writeDeadlineCalls int
	lastWriteDeadline  time.Time
	written            bytes.Buffer
}

func (c *failingWriteConn) Read(_ []byte) (int, error) { return 0, io.EOF }
func (c *failingWriteConn) Write(p []byte) (int, error) {
	if c.writeErr != nil {
		return 0, c.writeErr
	}
	c.written.Write(p)
	return len(p), nil
}
func (c *failingWriteConn) Close() error         { return nil }
func (c *failingWriteConn) LocalAddr() net.Addr  { return &net.TCPAddr{} }
func (c *failingWriteConn) RemoteAddr() net.Addr { return &net.TCPAddr{} }
func (c *failingWriteConn) SetDeadline(_ time.Time) error {
	return nil
}
func (c *failingWriteConn) SetReadDeadline(_ time.Time) error {
	return nil
}
func (c *failingWriteConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadlineCalls++
	c.lastWriteDeadline = t
	return nil
}

func TestWriteFrameSetsNetworkWriteDeadline(t *testing.T) {
	rawConn := &failingWriteConn{}
	c := &Conn{
		conn:        rawConn,
		bw:          bufio.NewWriterSize(rawConn, defaultBufSize),
		sendTimeout: time.Second,
	}

	if err := c.writeFrame(OpcodeText, true, []byte("x")); err != nil {
		t.Fatalf("writeFrame() error = %v", err)
	}
	if rawConn.writeDeadlineCalls < 2 {
		t.Fatalf("SetWriteDeadline calls = %d, want at least 2", rawConn.writeDeadlineCalls)
	}
	if !rawConn.lastWriteDeadline.IsZero() {
		t.Fatal("expected write deadline to be cleared after write")
	}
}

func TestWriteFrameSetsDefaultNetworkWriteDeadline(t *testing.T) {
	rawConn := &failingWriteConn{}
	c := &Conn{
		conn: rawConn,
		bw:   bufio.NewWriterSize(rawConn, defaultBufSize),
	}

	if err := c.writeFrame(OpcodeText, true, []byte("x")); err != nil {
		t.Fatalf("writeFrame() error = %v", err)
	}
	if rawConn.writeDeadlineCalls < 2 {
		t.Fatalf("SetWriteDeadline calls = %d, want at least 2", rawConn.writeDeadlineCalls)
	}
	if !rawConn.lastWriteDeadline.IsZero() {
		t.Fatal("expected write deadline to be cleared after write")
	}
}

func TestWriteCloseSendsCloseFrameBeforeClosing(t *testing.T) {
	rawConn := &failingWriteConn{}
	c := &Conn{
		conn:        rawConn,
		bw:          bufio.NewWriterSize(rawConn, defaultBufSize),
		closeC:      make(chan struct{}),
		sendTimeout: time.Second,
	}

	if err := c.WriteClose(CloseNormalClosure, "bye"); err != nil {
		t.Fatalf("WriteClose() error = %v", err)
	}
	written := rawConn.written.Bytes()
	if len(written) < 2 {
		t.Fatalf("written frame length = %d, want at least 2", len(written))
	}
	if got := written[0] & 0x0F; got != opcodeClose {
		t.Fatalf("opcode = %d, want close", got)
	}
	if !c.IsClosed() {
		t.Fatal("expected connection to be closed")
	}
}

func TestWriteCloseReturnsFrameWriteErrorAndCloses(t *testing.T) {
	writeErr := errors.New("close frame write failed")
	rawConn := &failingWriteConn{writeErr: writeErr}
	c := &Conn{
		conn:        rawConn,
		bw:          bufio.NewWriterSize(rawConn, defaultBufSize),
		closeC:      make(chan struct{}),
		sendTimeout: time.Second,
	}

	if err := c.WriteClose(CloseNormalClosure, "bye"); !errors.Is(err, writeErr) {
		t.Fatalf("WriteClose() error = %v, want %v", err, writeErr)
	}
	if !c.IsClosed() {
		t.Fatal("expected connection to be closed after failed close frame write")
	}
}

func TestWriteCloseRejectsInvalidClosePayload(t *testing.T) {
	tests := []struct {
		name   string
		code   uint16
		reason string
		want   error
	}{
		{
			name: "invalid close code",
			code: 1006,
			want: ErrProtocolError,
		},
		{
			name:   "invalid utf8 reason",
			code:   CloseNormalClosure,
			reason: string([]byte{0xff}),
			want:   ErrProtocolError,
		},
		{
			name:   "control payload too large",
			code:   CloseNormalClosure,
			reason: string(bytes.Repeat([]byte("x"), int(maxControlPayload))),
			want:   ErrControlTooLarge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rawConn := &failingWriteConn{}
			c := &Conn{
				conn:        rawConn,
				bw:          bufio.NewWriterSize(rawConn, defaultBufSize),
				closeC:      make(chan struct{}),
				sendTimeout: time.Second,
			}
			if err := c.WriteClose(tc.code, tc.reason); !errors.Is(err, tc.want) {
				t.Fatalf("WriteClose() error = %v, want %v", err, tc.want)
			}
			if rawConn.written.Len() != 0 {
				t.Fatalf("WriteClose wrote %d bytes for invalid close payload", rawConn.written.Len())
			}
			if c.IsClosed() {
				t.Fatal("invalid WriteClose closed connection")
			}
		})
	}
}

func TestWriteMessageRejectsInvalidDataOpcode(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan outbound, 1),
		sendBehavior: SendBlock,
		closeC:       make(chan struct{}),
	}

	if err := c.WriteMessageContext(nil, opcodeClose, []byte("bad")); !errors.Is(err, ErrProtocolError) {
		t.Fatalf("WriteMessageContext invalid opcode error = %v, want ErrProtocolError", err)
	}
	if len(c.sendQueue) != 0 {
		t.Fatal("invalid opcode was enqueued")
	}
}

func TestWriteMessageOwnsPayloadBytes(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan outbound, 1),
		sendBehavior: SendBlock,
		closeC:       make(chan struct{}),
	}
	payload := []byte("before")
	if err := c.WriteMessageContext(nil, OpcodeText, payload); err != nil {
		t.Fatalf("WriteMessageContext error = %v", err)
	}
	payload[0] = 'a'

	out := <-c.sendQueue
	if got := string(out.Data); got != "before" {
		t.Fatalf("queued data = %q, want caller-owned snapshot", got)
	}
}

func TestWriteMessageDropDoesNotCopyPayloadWhenQueueFull(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan outbound, 1),
		sendBehavior: SendDrop,
		closeC:       make(chan struct{}),
	}
	c.sendQueue <- outbound{Op: OpcodeText, Data: []byte("full")}
	payload := bytes.Repeat([]byte("x"), 1<<20)

	if err := c.WriteMessageContext(nil, OpcodeText, payload); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("WriteMessageContext error = %v, want ErrQueueFull", err)
	}
	allocs := testing.AllocsPerRun(100, func() {
		_ = c.WriteMessageContext(nil, OpcodeText, payload)
	})
	if allocs != 0 {
		t.Fatalf("full SendDrop path allocations = %v, want 0", allocs)
	}
}

func TestWriteMessageFastPathObservesClosedChannel(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan outbound, 1),
		sendBehavior: SendBlock,
		closeC:       make(chan struct{}),
	}
	close(c.closeC)

	if err := c.WriteMessageContext(nil, OpcodeText, []byte("lost")); !errors.Is(err, ErrConnClosed) {
		t.Fatalf("WriteMessageContext error = %v, want ErrConnClosed", err)
	}
	if len(c.sendQueue) != 0 {
		t.Fatal("message enqueued after close was visible")
	}
}

func TestWriteMessageUsesContextWriteDeadline(t *testing.T) {
	c := &Conn{sendTimeout: time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	timeout := c.writeTimeoutForContext(ctx)
	if timeout <= 0 || timeout > 100*time.Millisecond {
		t.Fatalf("write timeout = %v, want context-derived short timeout", timeout)
	}
}

func TestQueuedWriteKeepsAbsoluteContextDeadline(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan outbound, 1),
		sendBehavior: SendBlock,
		closeC:       make(chan struct{}),
		sendTimeout:  time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := c.WriteMessageContext(ctx, OpcodeText, []byte("late")); err != nil {
		t.Fatalf("WriteMessageContext error = %v", err)
	}
	out := <-c.sendQueue
	time.Sleep(20 * time.Millisecond)

	if _, err := c.writeDeadlineForOutbound(out); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("writeDeadlineForOutbound error = %v, want context deadline exceeded", err)
	}
}

func TestWriterPumpDoesNotWriteExpiredContextMessage(t *testing.T) {
	rawConn := &failingWriteConn{}
	c := &Conn{
		conn:      rawConn,
		bw:        bufio.NewWriterSize(rawConn, defaultBufSize),
		sendQueue: make(chan outbound, 1),
		closeC:    make(chan struct{}),
	}
	c.pingPeriod.Store(int64(time.Hour))
	c.sendQueue <- outbound{
		Op:            OpcodeText,
		Data:          []byte("late"),
		WriteTimeout:  time.Second,
		WriteDeadline: time.Now().Add(-time.Nanosecond),
	}

	done := make(chan struct{})
	go func() {
		c.writerPump()
		close(done)
	}()

	select {
	case <-c.closeC:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("connection was not closed for expired queued write")
	}
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("writerPump did not exit after expired queued write")
	}
	if rawConn.written.Len() != 0 {
		t.Fatalf("expired queued write wrote %d bytes", rawConn.written.Len())
	}
}

func TestWriteMessageContextNilContext(t *testing.T) {
	c := &Conn{
		sendQueue:    make(chan outbound, 1),
		sendBehavior: SendBlock,
		closeC:       make(chan struct{}),
	}
	c.sendQueue <- outbound{Op: OpcodeText, Data: []byte("queued")}

	go func() {
		time.Sleep(10 * time.Millisecond)
		_ = c.Close()
	}()

	if err := c.WriteMessageContext(nil, OpcodeText, []byte("blocked")); !errors.Is(err, ErrConnClosed) {
		t.Fatalf("WriteMessageContext(nil) error = %v, want ErrConnClosed", err)
	}
}

func TestWriterPumpClosesOnWriteError(t *testing.T) {
	rawConn := &failingWriteConn{writeErr: errors.New("write failed")}
	c := &Conn{
		conn:      rawConn,
		bw:        bufio.NewWriterSize(rawConn, defaultBufSize),
		sendQueue: make(chan outbound, 1),
		closeC:    make(chan struct{}),
	}
	c.pingPeriod.Store(int64(time.Hour))

	done := make(chan struct{})
	go func() {
		c.writerPump()
		close(done)
	}()

	c.sendQueue <- outbound{Op: OpcodeText, Data: []byte("x")}

	select {
	case <-c.closeC:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("connection was not closed on write error")
	}

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("writerPump did not exit after write error")
	}
}

func TestWriterPumpClosesOnPingWriteError(t *testing.T) {
	rawConn := &failingWriteConn{writeErr: errors.New("ping write failed")}
	c := &Conn{
		conn:      rawConn,
		bw:        bufio.NewWriterSize(rawConn, defaultBufSize),
		sendQueue: make(chan outbound, 1),
		closeC:    make(chan struct{}),
	}
	c.pingPeriod.Store(int64(10 * time.Millisecond))

	done := make(chan struct{})
	go func() {
		c.writerPump()
		close(done)
	}()

	select {
	case <-c.closeC:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("connection was not closed after ping write error")
	}

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("writerPump did not exit after ping write error")
	}
}
