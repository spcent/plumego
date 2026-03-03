package websocket

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// msgBufPool reuses bytes.Buffer instances across read operations
// to reduce allocator pressure from per-message buffer creation.
var msgBufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// SendBehavior determines behavior on enqueue timeout / full queue.
//
// SendBehavior controls how the connection handles message sending when the
// send queue is full or a timeout occurs.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	// Block until space available (default)
//	conn := websocket.NewConn(netConn, 100, 5*time.Second, websocket.SendBlock)
//
//	// Drop message when queue full
//	conn := websocket.NewConn(netConn, 100, 5*time.Second, websocket.SendDrop)
//
//	// Close connection when queue full
//	conn := websocket.NewConn(netConn, 100, 5*time.Second, websocket.SendClose)
type SendBehavior int

const (
	// SendBlock blocks until space is available in the queue.
	// Can still timeout if context is used.
	SendBlock SendBehavior = iota

	// SendDrop drops the message when the queue is full.
	SendDrop

	// SendClose closes the connection when the queue is full.
	SendClose
)

// Outbound is an internal message for sending.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	msg := websocket.Outbound{
//		Op:   websocket.OpcodeText,
//		Data: []byte("Hello, World!"),
//	}
type Outbound struct {
	Op   byte
	Data []byte
}

// UserInfo stores authenticated user information.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	userInfo := websocket.UserInfo{
//		ID:    "user-123",
//		Name:  "John Doe",
//		Email: "john@example.com",
//		Roles: []string{"admin", "user"},
//	}
type UserInfo struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Email  string         `json:"email"`
	Roles  []string       `json:"roles"`
	Claims map[string]any `json:"claims"`
}

// Conn is a websocket connection wrapper with stream API and bounded queue send.
//
// Conn provides a production-ready WebSocket connection with:
//   - Bounded send queue with configurable behavior
//   - Ping/pong heartbeat monitoring
//   - Read message size limits
//   - User authentication support
//   - Connection metadata storage
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	// Create connection with blocking send
//	conn := websocket.NewConn(netConn, 100, 5*time.Second, websocket.SendBlock)
//	defer conn.Close()
//
//	// Send a message
//	err := conn.WriteMessage(websocket.OpcodeText, []byte("Hello"))
//
//	// Read messages
//	for {
//		op, data, err := conn.ReadMessage()
//		if err != nil {
//			break
//		}
//		// Process message
//	}
//
//	// Set user info after authentication
//	conn.UserInfo = &websocket.UserInfo{
//		ID:    "user-123",
//		Name:  "John Doe",
//		Roles: []string{"admin"},
//	}
type Conn struct {
	conn net.Conn
	br   *bufio.Reader
	bw   *bufio.Writer

	writeMu sync.Mutex

	// send queue
	sendQueue chan Outbound
	// config
	sendQueueSize int
	sendTimeout   time.Duration
	sendBehavior  SendBehavior

	closeOnce sync.Once
	closed    atomic.Int32
	closeC    chan struct{}

	readLimit  atomic.Int64
	pingPeriod atomic.Int64 // stores time.Duration as int64 nanoseconds
	pongWait   atomic.Int64 // stores time.Duration as int64 nanoseconds
	lastPong   atomic.Int64

	// User information (set after authentication)
	UserInfo *UserInfo

	// Connection metadata (concurrent-safe)
	metadata sync.Map
}

// NewConn creates a Conn after handshake.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	// Create connection with blocking send
//	conn := websocket.NewConn(netConn, 100, 5*time.Second, websocket.SendBlock)
//	defer conn.Close()
//
//	// Create connection with drop behavior
//	conn := websocket.NewConn(netConn, 100, 5*time.Second, websocket.SendDrop)
// NewConn creates a Conn after handshake, allocating its own buffered I/O.
//
// For server-side connections obtained via http.Hijacker, prefer
// newConnFromHijack to reuse the bufio.ReadWriter that the HTTP server
// already created, avoiding a redundant allocation.
func NewConn(c net.Conn, queueSize int, sendTimeout time.Duration, behavior SendBehavior) *Conn {
	return newConnFromHijack(
		c,
		bufio.NewReaderSize(c, defaultBufSize),
		bufio.NewWriterSize(c, defaultBufSize),
		queueSize, sendTimeout, behavior,
	)
}

// newConnFromHijack creates a Conn using buffers already allocated by the
// HTTP server's hijack operation, avoiding a redundant allocation pair.
func newConnFromHijack(c net.Conn, br *bufio.Reader, bw *bufio.Writer, queueSize int, sendTimeout time.Duration, behavior SendBehavior) *Conn {
	cc := &Conn{
		conn:          c,
		br:            br,
		bw:            bw,
		sendQueue:     make(chan Outbound, queueSize),
		sendQueueSize: queueSize,
		sendTimeout:   sendTimeout,
		sendBehavior:  behavior,
		closeC:        make(chan struct{}),
	}
	cc.readLimit.Store(16 << 20) // 16MB
	cc.pingPeriod.Store(int64(defaultPingPeriod))
	cc.pongWait.Store(int64(defaultPongWait))
	cc.lastPong.Store(time.Now().UnixNano())
	go cc.writerPump()
	go cc.pongMonitor()
	return cc
}

func (c *Conn) IsClosed() bool { return c.closed.Load() == 1 }

func (c *Conn) Close() error {
	var err error
	c.closeOnce.Do(func() {
		c.closed.Store(1)
		close(c.closeC)
		if c.conn != nil {
			err = c.conn.Close()
		}
	})
	return err
}

// SetReadLimit sets the maximum message size
func (c *Conn) SetReadLimit(limit int64) {
	c.readLimit.Store(limit)
}

// SetPingPeriod sets the ping interval
func (c *Conn) SetPingPeriod(d time.Duration) {
	c.pingPeriod.Store(int64(d))
}

// SetPongWait sets the pong wait time
func (c *Conn) SetPongWait(d time.Duration) {
	c.pongWait.Store(int64(d))
}

// GetLastPong returns the last pong time
func (c *Conn) GetLastPong() time.Time {
	return time.Unix(0, c.lastPong.Load())
}

// WriteClose sends a WebSocket close frame with the given RFC 6455 status code
// and human-readable reason, then closes the underlying connection.
//
// This initiates a proper WebSocket-level closing handshake. Use the Close*
// constants (CloseNormalClosure, CloseGoingAway, etc.) for the status code.
// Calling Close() directly skips the close frame and tears down TCP immediately,
// which is correct for error conditions but not for clean shutdowns.
//
// Example:
//
//	conn.WriteClose(websocket.CloseNormalClosure, "goodbye")
//	conn.WriteClose(websocket.CloseGoingAway, "server shutting down")
func (c *Conn) WriteClose(code uint16, reason string) error {
	if c.IsClosed() {
		return ErrConnClosed
	}
	// RFC 6455 §5.5.1: close payload is 2-byte big-endian status code followed
	// by a UTF-8 reason phrase (may be empty).
	payload := make([]byte, 2+len(reason))
	binary.BigEndian.PutUint16(payload[:2], code)
	copy(payload[2:], reason)
	// Write directly, bypassing sendQueue so the frame is sent even when the
	// queue is full (e.g. during a slow-consumer shutdown).
	_ = c.writeFrame(opcodeClose, true, payload)
	return c.Close()
}

// ---------------- low-level frame IO ----------------

func (c *Conn) readFrame() (byte, bool, []byte, error) {
	var h [2]byte
	if _, err := io.ReadFull(c.br, h[:]); err != nil {
		return 0, false, nil, err
	}
	fin := h[0]&finBit != 0
	op := h[0] & 0x0F
	mask := h[1]&0x80 != 0
	prefix := int64(h[1] & 0x7F)

	if !mask {
		return 0, false, nil, ErrUnmaskedFrame
	}

	var payloadLen int64
	switch prefix {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(c.br, ext[:]); err != nil {
			return 0, false, nil, err
		}
		payloadLen = int64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(c.br, ext[:]); err != nil {
			return 0, false, nil, err
		}
		// RFC 6455 §5.2: the MSB of a 64-bit payload length MUST be 0.
		// A negative int64 means the MSB is set; reject to prevent a panic in make().
		payloadLen = int64(binary.BigEndian.Uint64(ext[:]))
		if payloadLen < 0 {
			return 0, false, nil, ErrProtocolError
		}
	default:
		payloadLen = prefix
	}

	// Reject oversized frames before reading the mask key so we avoid
	// a 4-byte network round-trip for frames we will reject anyway.
	if payloadLen > c.readLimit.Load() {
		return 0, false, nil, ErrPayloadTooLarge
	}

	var maskKey [4]byte
	if _, err := io.ReadFull(c.br, maskKey[:]); err != nil {
		return 0, false, nil, err
	}
	payload := make([]byte, payloadLen)
	if payloadLen > 0 {
		if _, err := io.ReadFull(c.br, payload); err != nil {
			return 0, false, nil, err
		}
		// Unmask the payload. Process 4 bytes per iteration to let the compiler
		// optimise the loop; avoids the modulo in the byte-by-byte version.
		n := len(payload)
		i := 0
		for ; i+3 < n; i += 4 {
			payload[i] ^= maskKey[0]
			payload[i+1] ^= maskKey[1]
			payload[i+2] ^= maskKey[2]
			payload[i+3] ^= maskKey[3]
		}
		for ; i < n; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	// control frame checks
	if op >= 0x8 {
		if !fin {
			return 0, false, nil, ErrFragmentedControl
		}
		if int64(len(payload)) > maxControlPayload {
			return 0, false, nil, ErrControlTooLarge
		}
	}
	return op, fin, payload, nil
}

func (c *Conn) writeFrame(op byte, fin bool, payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	var header [14]byte
	hlen := 0
	b0 := byte(0)
	if fin {
		b0 |= finBit
	}
	b0 |= op & 0x0F
	header[0] = b0
	hlen = 1

	n := len(payload)
	switch {
	case n <= 125:
		header[hlen] = byte(n)
		hlen++
	case n <= 0xFFFF:
		header[hlen] = 126
		hlen++
		header[hlen] = byte(n >> 8)
		header[hlen+1] = byte(n)
		hlen += 2
	default:
		header[hlen] = 127
		hlen++
		binary.BigEndian.PutUint64(header[hlen:hlen+8], uint64(n))
		hlen += 8
	}

	if _, err := c.bw.Write(header[:hlen]); err != nil {
		return err
	}
	if n > 0 {
		if _, err := c.bw.Write(payload); err != nil {
			return err
		}
	}
	return c.bw.Flush()
}

// SetMetadata sets a metadata value for the connection (concurrent-safe).
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	conn := websocket.NewConn(...)
//	conn.SetMetadata("session_id", "abc123")
//	conn.SetMetadata("client_ip", "192.168.1.1")
func (c *Conn) SetMetadata(key string, value any) {
	c.metadata.Store(key, value)
}

// GetMetadata retrieves a metadata value for the connection (concurrent-safe).
//
// Returns the value and true if found, nil and false otherwise.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	conn := websocket.NewConn(...)
//	conn.SetMetadata("session_id", "abc123")
//	if sessionID, ok := conn.GetMetadata("session_id"); ok {
//		fmt.Printf("Session ID: %v\n", sessionID)
//	}
func (c *Conn) GetMetadata(key string) (any, bool) {
	return c.metadata.Load(key)
}

// DeleteMetadata removes a metadata value from the connection (concurrent-safe).
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	conn := websocket.NewConn(...)
//	conn.SetMetadata("temp_data", "value")
//	conn.DeleteMetadata("temp_data")
func (c *Conn) DeleteMetadata(key string) {
	c.metadata.Delete(key)
}

// RangeMetadata iterates over all metadata key-value pairs (concurrent-safe).
//
// The iteration stops if the function returns false.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	conn := websocket.NewConn(...)
//	conn.SetMetadata("key1", "value1")
//	conn.SetMetadata("key2", "value2")
//	conn.RangeMetadata(func(key, value any) bool {
//		fmt.Printf("%v: %v\n", key, value)
//		return true // continue iteration
//	})
func (c *Conn) RangeMetadata(f func(key, value any) bool) {
	c.metadata.Range(f)
}
