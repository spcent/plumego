package websocket

import (
	"bufio"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

type failingWriteConn struct {
	writeErr error
}

func (c *failingWriteConn) Read(_ []byte) (int, error)  { return 0, io.EOF }
func (c *failingWriteConn) Write(_ []byte) (int, error) { return 0, c.writeErr }
func (c *failingWriteConn) Close() error                { return nil }
func (c *failingWriteConn) LocalAddr() net.Addr         { return &net.TCPAddr{} }
func (c *failingWriteConn) RemoteAddr() net.Addr        { return &net.TCPAddr{} }
func (c *failingWriteConn) SetDeadline(_ time.Time) error {
	return nil
}
func (c *failingWriteConn) SetReadDeadline(_ time.Time) error {
	return nil
}
func (c *failingWriteConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

func TestWriterPumpClosesOnWriteError(t *testing.T) {
	rawConn := &failingWriteConn{writeErr: errors.New("write failed")}
	c := &Conn{
		conn:      rawConn,
		bw:        bufio.NewWriterSize(rawConn, defaultBufSize),
		sendQueue: make(chan Outbound, 1),
		closeC:    make(chan struct{}),
	}
	c.pingPeriod.Store(int64(time.Hour))

	done := make(chan struct{})
	go func() {
		c.writerPump()
		close(done)
	}()

	c.sendQueue <- Outbound{Op: OpcodeText, Data: []byte("x")}

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
		sendQueue: make(chan Outbound, 1),
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
