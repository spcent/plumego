// A single-file WebSocket server using only Go standard library.
// Features:
// - RFC6455 handshake (Hijack) and framing
// - Conn with ReadMessage/ReadMessageStream/WriteMessage
// - Fragmentation support, stream API (io.ReadCloser) for large binary/text messages
// - Per-connection bounded send queue with overflow/timeout policies
// - Hub with rooms/topics, worker-pool broadcast to improve backpressure handling
// - Simple auth: room password and HS256 JWT verification (standard library only)
// - Tests and a benchmark that simulates many clients
package websocket

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spcent/plumego/security/password"
)

const (
	guid                     = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	opcodeContinuation byte  = 0x0
	OpcodeText         byte  = 0x1
	OpcodeBinary       byte  = 0x2
	opcodeClose        byte  = 0x8
	opcodePing         byte  = 0x9
	opcodePong         byte  = 0xA
	finBit             byte  = 0x80
	defaultBufSize     int   = 4096
	defaultPingPeriod        = 20 * time.Second
	defaultPongWait          = 30 * time.Second
	maxControlPayload  int64 = 125
	maxFragmentSize    int   = 64 * 1024 // 64KB
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
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Email  string                 `json:"email"`
	Roles  []string               `json:"roles"`
	Claims map[string]interface{} `json:"claims"`
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
}

// NewConn creates a Conn after handshake
func NewConn(c net.Conn, queueSize int, sendTimeout time.Duration, behavior SendBehavior) *Conn {
	cc := &Conn{
		conn:          c,
		br:            bufio.NewReaderSize(c, 8192),
		bw:            bufio.NewWriterSize(c, 8192),
		sendQueue:     make(chan Outbound, queueSize),
		sendQueueSize: queueSize,
		sendTimeout:   sendTimeout,
		sendBehavior:  behavior,
		closeC:        make(chan struct{}),
		readLimit:     16 << 20, // 16MB
		pingPeriod:    defaultPingPeriod,
		pongWait:      defaultPongWait,
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
		err = c.conn.Close()
	})
	return err
}

// ---------------- low-level frame IO ----------------

func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + guid))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func headerContains(h http.Header, key, val string) bool {
	v := h.Get(key)
	if v == "" {
		return false
	}
	parts := strings.Split(v, ",")
	for _, p := range parts {
		if strings.EqualFold(strings.TrimSpace(p), val) {
			return true
		}
	}
	return false
}

// readFrame reads one webSocket frame from client, unmasked, enforces RFC rules for client frames (must be masked).
// returns op, fin, payload, error
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

// writeFrame writes server->client frame (not masked)
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

// ---------------- stream read API ----------------

// streamReader implements io.ReadCloser to stream frames for one message.
// It reads frames (continuation) from the connection and returns bytes as Read is called.
// It stops after FIN for the message. It must be used by only one reader at a time.
type streamReader struct {
	parent    *Conn
	op        byte
	buf       bytes.Buffer
	done      bool
	readErr   error
	readMu    sync.Mutex
	closeOnce sync.Once
	closed    chan struct{}
}

func (sr *streamReader) Read(p []byte) (int, error) {
	sr.readMu.Lock()
	defer sr.readMu.Unlock()

	for {
		if sr.buf.Len() > 0 {
			return sr.buf.Read(p)
		}
		if sr.done {
			if sr.readErr != nil {
				return 0, sr.readErr
			}
			return 0, io.EOF
		}
		// need to pull next frame(s) from connection
		op, fin, payload, err := sr.parent.readFrame()
		if err != nil {
			sr.readErr = err
			sr.done = true
			return 0, err
		}
		// control frames may appear in middle
		switch op {
		case opcodePing:
			_ = sr.parent.writeFrame(opcodePong, true, payload)
			continue
		case opcodePong:
			atomic.StoreInt64(&sr.parent.lastPong, time.Now().UnixNano())
			continue
		case opcodeClose:
			_ = sr.parent.writeFrame(opcodeClose, true, payload)
			sr.readErr = io.EOF
			sr.done = true
			return 0, io.EOF
		case opcodeContinuation:
			// append
			sr.buf.Write(payload)
			if fin {
				sr.done = true
			}
			continue
		case OpcodeText, OpcodeBinary:
			// if new data opcode arrives while not started, treat as start
			// if already started assembling and new data opcode arrives -> protocol error
			if sr.op == 0 {
				sr.op = op
				sr.buf.Write(payload)
				if fin {
					sr.done = true
				}
				continue
			} else {
				// shouldn't get new data opcode while assembling
				sr.readErr = errors.New("protocol error: new data opcode while assembling")
				sr.done = true
				return 0, sr.readErr
			}
		default:
			// ignore other opcodes
			continue
		}
	}
}

func (sr *streamReader) Close() error {
	sr.closeOnce.Do(func() { close(sr.closed) })
	return nil
}

// ReadMessageStream returns (opcode, io.ReadCloser, error).
// Caller must Close() the returned ReadCloser when finished to allow connection continue.
func (c *Conn) ReadMessageStream() (byte, io.ReadCloser, error) {
	if c.IsClosed() {
		return 0, nil, errors.New("connection closed")
	}
	// read first frame - must be text/binary or control
	for {
		op, fin, payload, err := c.readFrame()
		if err != nil {
			return 0, nil, err
		}
		switch op {
		case OpcodeText, OpcodeBinary:
			// If fin == true and small, can return a reader that contains payload and EOF immediately
			sr := &streamReader{
				parent: c,
				op:     0,
				closed: make(chan struct{}),
			}
			// write payload into buffer
			sr.buf.Write(payload)
			if fin {
				sr.done = true
			} else {
				// not finished; set op marker and let subsequent continuation frames be read by Read
				sr.op = op
			}
			return op, sr, nil
		case opcodePing:
			_ = c.writeFrame(opcodePong, true, payload)
			continue
		case opcodePong:
			atomic.StoreInt64(&c.lastPong, time.Now().UnixNano())
			continue
		case opcodeClose:
			_ = c.writeFrame(opcodeClose, true, payload)
			return 0, nil, io.EOF
		default:
			// ignore
			continue
		}
	}
}

// ---------------- WriteMessage with fragmentation and bounded queue ----------------

// WriteMessage enqueues message to sendQueue. Behavior on full queue depends on c.sendBehavior.
// It waits up to sendTimeout if blocking behavior chosen.
func (c *Conn) WriteMessage(op byte, data []byte) error {
	if c.IsClosed() {
		return errors.New("connection closed")
	}
	out := Outbound{Op: op, Data: data}
	select {
	case c.sendQueue <- out:
		return nil
	default:
		// queue full
		switch c.sendBehavior {
		case SendBlock:
			// wait up to timeout
			if c.sendTimeout <= 0 {
				c.sendQueue <- out // block until available
				return nil
			}
			timer := time.NewTimer(c.sendTimeout)
			defer timer.Stop()
			select {
			case c.sendQueue <- out:
				return nil
			case <-timer.C:
				return errors.New("send timeout")
			case <-c.closeC:
				return errors.New("connection closed")
			}
		case SendDrop:
			// drop silently (or return an error)
			return errors.New("send queue full: dropped")
		case SendClose:
			// close connection
			c.Close()
			return errors.New("send queue full: connection closed")
		default:
			return errors.New("unknown send behavior")
		}
	}
}

// writerPump consumes sendQueue and writes frames to client. It fragments large messages.
func (c *Conn) writerPump() {
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()
	for {
		select {
		case <-c.closeC:
			return
		case out, ok := <-c.sendQueue:
			if !ok {
				return
			}
			// fragment if needed
			data := out.Data
			if len(data) <= maxFragmentSize {
				_ = c.writeFrame(out.Op, true, data)
				continue
			}
			total := len(data)
			offset := 0
			first := true
			for offset < total {
				end := offset + maxFragmentSize
				if end > total {
					end = total
				}
				chunk := data[offset:end]
				fin := end == total
				var op byte
				if first {
					op = out.Op
					first = false
				} else {
					op = opcodeContinuation
				}
				if err := c.writeFrame(op, fin, chunk); err != nil {
					c.Close()
					return
				}
				offset = end
			}
		case <-ticker.C:
			// send ping
			_ = c.writeFrame(opcodePing, true, []byte("ping"))
		}
	}
}

// pongMonitor closes connection if no pong received within pongWait
func (c *Conn) pongMonitor() {
	ticker := time.NewTicker(c.pingPeriod / 2)
	defer ticker.Stop()
	for {
		select {
		case <-c.closeC:
			return
		case <-ticker.C:
			last := time.Unix(0, atomic.LoadInt64(&c.lastPong))
			if time.Since(last) > c.pongWait {
				_ = c.Close()
				return
			}
		}
	}
}

// ---------------- Hub with worker pool broadcast ----------------

type hubJob struct {
	conn *Conn
	op   byte
	data []byte
}

type Hub struct {
	rooms map[string]map[*Conn]struct{}
	mu    sync.RWMutex

	// worker pool
	jobQueue chan hubJob
	workers  int
	wg       sync.WaitGroup
	quit     chan struct{}
}

func NewHub(workerCount int, jobQueueSize int) *Hub {
	h := &Hub{
		rooms:    make(map[string]map[*Conn]struct{}),
		jobQueue: make(chan hubJob, jobQueueSize),
		workers:  workerCount,
		quit:     make(chan struct{}),
	}
	h.startWorkers()
	return h
}

func (h *Hub) startWorkers() {
	for i := 0; i < h.workers; i++ {
		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			for {
				select {
				case j, ok := <-h.jobQueue:
					if !ok {
						return
					}
					// best-effort write: we enqueue to conn's sendQueue with conn's behavior handling overflow
					_ = j.conn.WriteMessage(j.op, j.data)
				case <-h.quit:
					return
				}
			}
		}()
	}
}

func (h *Hub) Stop() {
	close(h.quit)
	close(h.jobQueue)
	h.wg.Wait()
}

// Join room
func (h *Hub) Join(room string, c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rs, ok := h.rooms[room]
	if !ok {
		rs = make(map[*Conn]struct{})
		h.rooms[room] = rs
	}
	rs[c] = struct{}{}
}

// Leave room
func (h *Hub) Leave(room string, c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if rs, ok := h.rooms[room]; ok {
		delete(rs, c)
		if len(rs) == 0 {
			delete(h.rooms, room)
		}
	}
}

// RemoveConn from all rooms
func (h *Hub) RemoveConn(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for room, rs := range h.rooms {
		if _, ok := rs[c]; ok {
			delete(rs, c)
			if len(rs) == 0 {
				delete(h.rooms, room)
			}
		}
	}
}

// BroadcastRoom enqueues jobs to jobQueue for workers to send.
func (h *Hub) BroadcastRoom(room string, op byte, data []byte) {
	h.mu.RLock()
	rs, ok := h.rooms[room]
	h.mu.RUnlock()
	if !ok || len(rs) == 0 {
		return
	}
	for c := range rs {
		select {
		case h.jobQueue <- hubJob{conn: c, op: op, data: data}:
		default:
			// if jobQueue full, drop to avoid blocking; alternative: block or expand queue
		}
	}
}

// BroadcastAll broadcasts to all clients
func (h *Hub) BroadcastAll(op byte, data []byte) {
	h.mu.RLock()
	var conns []*Conn
	for _, rs := range h.rooms {
		for c := range rs {
			conns = append(conns, c)
		}
	}
	h.mu.RUnlock()
	for _, c := range conns {
		select {
		case h.jobQueue <- hubJob{conn: c, op: op, data: data}:
		default:
		}
	}
}

// ---------------- Authentication: room password and HS256 JWT verification ----------------

// simpleRoomAuth stores metadata about rooms (password)
type simpleRoomAuth struct {
	roomPasswords map[string]string // room -> password (hashed)
	mu            sync.RWMutex
	jwtSecret     []byte // HMAC secret for HS256
}

func NewSimpleRoomAuth(secret []byte) *simpleRoomAuth {
	return &simpleRoomAuth{
		roomPasswords: make(map[string]string),
		jwtSecret:     secret,
	}
}
func (s *simpleRoomAuth) SetRoomPassword(room, pwd string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Hash password before storing
	hashed, err := password.HashPassword(pwd)
	if err != nil {
		log.Printf("Error hashing room password: %v", err)
		return
	}
	s.roomPasswords[room] = string(hashed)
}
func (s *simpleRoomAuth) CheckRoomPassword(room, provided string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if hashed, ok := s.roomPasswords[room]; ok {
		// Verify hash
		err := password.CheckPassword(hashed, provided)
		return err == nil
	}
	// no password set => allowed
	return true
}

// verifyJWTHS256 verifies HS256 token and returns payload map
func (s *simpleRoomAuth) VerifyJWT(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid jwt")
	}
	hdrb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var hdr map[string]any
	if err = json.Unmarshal(hdrb, &hdr); err != nil {
		return nil, err
	}
	if alg, ok := hdr["alg"].(string); !ok || alg != "HS256" {
		return nil, errors.New("unsupported jwt alg")
	}
	payloadb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err = json.Unmarshal(payloadb, &payload); err != nil {
		return nil, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, s.jwtSecret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expected := mac.Sum(nil)
	if !hmac.Equal(expected, sig) {
		return nil, errors.New("invalid jwt signature")
	}
	// Verify exp if present
	if expv, ok := payload["exp"]; ok {
		switch t := expv.(type) {
		case float64:
			if time.Now().Unix() > int64(t) {
				return nil, errors.New("jwt expired")
			}
		case int64:
			if time.Now().Unix() > t {
				return nil, errors.New("jwt expired")
			}
		}
	}
	return payload, nil
}

// ---------------- ServeWS with auth ----------------

// ServeWSWithAuth does the handshake and basic auth checks before accepting connection.
// Accepts token from Authorization header "Bearer <token>" or ?token= in query.
// room password from ?room_password= param.
// onConn will be called when Conn is ready.
func ServeWSWithAuth(w http.ResponseWriter, r *http.Request, hub *Hub, auth *simpleRoomAuth, queueSize int, sendTimeout time.Duration, behavior SendBehavior) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !headerContains(r.Header, "Connection", "Upgrade") || !headerContains(r.Header, "Upgrade", "websocket") {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// auth: room and JWT
	room := r.URL.Query().Get("room")
	if room == "" {
		room = "default"
	}
	// check room password
	roomPwd := r.URL.Query().Get("room_password")
	if !auth.CheckRoomPassword(room, roomPwd) {
		http.Error(w, "forbidden: bad room password", http.StatusForbidden)
		return
	}
	// check token if present
	token := ""
	var userInfo *UserInfo
	if ah := r.Header.Get("Authorization"); ah != "" && strings.HasPrefix(strings.ToLower(ah), "bearer ") {
		token = strings.TrimSpace(ah[len("bearer "):])
	} else if t := r.URL.Query().Get("token"); t != "" {
		token = t
	}
	if token != "" {
		payload, err := auth.VerifyJWT(token)
		if err != nil {
			http.Error(w, "forbidden: invalid token", http.StatusForbidden)
			return
		}
		// Extract user information from JWT payload
		userInfo = &UserInfo{
			Claims: payload,
		}
		// Extract common claims if present
		if id, ok := payload["sub"].(string); ok {
			userInfo.ID = id
		}
		if name, ok := payload["name"].(string); ok {
			userInfo.Name = name
		}
		if email, ok := payload["email"].(string); ok {
			userInfo.Email = email
		}
		if roles, ok := payload["roles"].([]string); ok {
			userInfo.Roles = roles
		}
	}
	accept := computeAcceptKey(key)
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "server does not support hijacking", http.StatusInternalServerError)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		http.Error(w, "hijack failed", http.StatusInternalServerError)
		return
	}
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n" +
		"\r\n"
	if _, err := buf.WriteString(resp); err != nil {
		conn.Close()
		return
	}
	if err := buf.Flush(); err != nil {
		conn.Close()
		return
	}

	c := NewConn(conn, queueSize, sendTimeout, behavior)
	// Set user information if authenticated
	c.UserInfo = userInfo
	// override br/bw with hijacked conn
	c.br = bufio.NewReaderSize(conn, 8192)
	c.bw = bufio.NewWriterSize(conn, 8192)

	// register in hub
	hub.Join(room, c)
	// cleanup on close
	go func() {
		<-c.closeC
		hub.Leave(room, c)
		hub.RemoveConn(c)
	}()
	// spawn a goroutine to read frames and push to room broadcast on completed stream
	go func() {
		for {
			op, rstream, err := c.ReadMessageStream()
			if err != nil {
				if err != io.EOF {
					log.Println("ReadMessageStream error:", err)
				}
				c.Close()
				return
			}
			// For demonstration: stream copy to a buffer but streaming-friendly.
			// If message is huge, we stream to temp file or process chunk-by-chunk.
			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, rstream) // streaming
			_ = rstream.Close()
			// broadcast to room
			hub.BroadcastRoom(room, op, buf.Bytes())
		}
	}()
}
