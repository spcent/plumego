package websocket

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// SendBehavior determines behavior on enqueue timeout / full queue.
type SendBehavior int

const (
	SendBlock SendBehavior = iota // Block until space available (can still timeout if context used)
	SendDrop                      // Drop message when queue full
	SendClose                     // Close connection when queue full
)

// Outbound is an internal message for sending
type Outbound struct {
	Op   byte
	Data []byte
}

// UserInfo stores authenticated user information
type UserInfo struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Email  string         `json:"email"`
	Roles  []string       `json:"roles"`
	Claims map[string]any `json:"claims"`
}

// Conn is a websocket connection wrapper with stream API and bounded queue send.
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
	closed    int32
	closeC    chan struct{}

	readLimit  int64
	pingPeriod time.Duration
	pongWait   time.Duration
	lastPong   int64

	// User information (set after authentication)
	UserInfo *UserInfo

	// Connection metadata
	Metadata map[string]any
}

// NewConn creates a Conn after handshake
func NewConn(c net.Conn, queueSize int, sendTimeout time.Duration, behavior SendBehavior) *Conn {
	cc := &Conn{
		conn:          c,
		br:            bufio.NewReaderSize(c, defaultBufSize),
		bw:            bufio.NewWriterSize(c, defaultBufSize),
		sendQueue:     make(chan Outbound, queueSize),
		sendQueueSize: queueSize,
		sendTimeout:   sendTimeout,
		sendBehavior:  behavior,
		closeC:        make(chan struct{}),
		readLimit:     16 << 20, // 16MB
		pingPeriod:    defaultPingPeriod,
		pongWait:      defaultPongWait,
		Metadata:      make(map[string]any),
	}
	atomic.StoreInt64(&cc.lastPong, time.Now().UnixNano())
	// start writer pump
	go cc.writerPump()
	// start ping/pong monitor
	go cc.pongMonitor()
	return cc
}

func (c *Conn) IsClosed() bool { return atomic.LoadInt32(&c.closed) == 1 }

func (c *Conn) Close() error {
	var err error
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closed, 1)
		close(c.closeC)
		if c.conn != nil {
			err = c.conn.Close()
		}
	})
	return err
}

// SetReadLimit sets the maximum message size
func (c *Conn) SetReadLimit(limit int64) {
	c.readLimit = limit
}

// SetPingPeriod sets the ping interval
func (c *Conn) SetPingPeriod(d time.Duration) {
	c.pingPeriod = d
}

// SetPongWait sets the pong wait time
func (c *Conn) SetPongWait(d time.Duration) {
	c.pongWait = d
}

// GetLastPong returns the last pong time
func (c *Conn) GetLastPong() time.Time {
	return time.Unix(0, atomic.LoadInt64(&c.lastPong))
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
		return 0, false, nil, errors.New("protocol error: unmasked client frame")
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
		payloadLen = int64(binary.BigEndian.Uint64(ext[:]))
	default:
		payloadLen = prefix
	}

	if payloadLen > c.readLimit {
		return 0, false, nil, errors.New("payload too large")
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
		for i := int64(0); i < payloadLen; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	// control frame checks
	if op >= 0x8 {
		if !fin {
			return 0, false, nil, errors.New("protocol error: fragmented control frame")
		}
		if int64(len(payload)) > maxControlPayload {
			return 0, false, nil, errors.New("protocol error: control frame too large")
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
