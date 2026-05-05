package websocket

import (
	"bufio"
	"bytes"
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
