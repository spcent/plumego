package websocket

import (
	"context"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type hubJob struct {
	conn *Conn
	op   byte
	data []byte
}

const hubWorkerDefaultSendTimeout = 100 * time.Millisecond

// BroadcastResult describes a broadcast fanout attempt.
type BroadcastResult struct {
	Attempted int  `json:"attempted"`
	Enqueued  int  `json:"enqueued"`
	Skipped   int  `json:"skipped"`
	Dropped   int  `json:"dropped"`
	Invalid   bool `json:"invalid"`
	Stopped   bool `json:"stopped"`
}

// Rejected reports whether the broadcast targeted at least one connection but
// could not enqueue the message for any connection, or was rejected before
// fanout because the input was invalid or the hub was stopped.
func (r BroadcastResult) Rejected() bool {
	return r.Invalid || r.Stopped || (r.Attempted > 0 && r.Enqueued == 0)
}

// Hub manages rooms and broadcast.
//
// Hub provides an experimental WebSocket hub with:
//   - Room-based message broadcasting
//   - Worker pool for concurrent message delivery
//   - Connection limits (total and per-room)
//   - Metrics collection
//   - Security event monitoring
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	// Create hub with 4 workers and 1024 job queue size
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	defer hub.Stop()
//
//	// Join a connection to a room
//	err := hub.TryJoin("chat-room", conn)
//	if err != nil {
//		// Handle capacity limits
//	}
//
//	// Broadcast to a room
//	hub.BroadcastRoom("chat-room", websocket.OpcodeText, []byte("Hello"))
//
//	// Get metrics
//	metrics := hub.Metrics()
type Hub struct {
	rooms map[string]map[*Conn]struct{}
	mu    sync.RWMutex

	// worker pool
	jobQueue    chan hubJob
	workers     int
	wg          sync.WaitGroup
	quit        chan struct{}
	lifecycleMu sync.RWMutex

	// stopped is set to true by Stop() before closing quit.
	// BroadcastRoom/BroadcastAll check this to avoid sending on a channel
	// whose readers have already exited, preventing any post-stop confusion.
	stopped atomic.Bool

	maxRoomRegistrations int
	maxRoomConns         int
	roomRegistrations    atomic.Uint64 // Atomic counter for active room registrations
	accepted             atomic.Uint64
	rejected             atomic.Uint64

	// Per-hub security/broadcast metrics (replaces global securityMetrics writes)
	broadcastAttempted atomic.Uint64 // broadcast target attempts
	broadcastEnqueued  atomic.Uint64 // messages enqueued to workers
	broadcastSkipped   atomic.Uint64 // targets skipped before enqueue
	broadcastDropped   atomic.Uint64 // messages dropped due to full job queue
	securityRejections atomic.Uint64 // connections rejected (origin/auth/capacity)
	invalidWSKeys      atomic.Uint64 // invalid Sec-WebSocket-Key headers seen
	successfulAuths    atomic.Uint64 // successful JWT authentications

	// Rate limiting (simple token bucket implementation)
	rateLimiter *simpleRateLimiter

	// Message pooling to reduce allocations
	connListPool sync.Pool // Pool of []*Conn slices

	// Hub configuration
	config HubConfig
	// Logger for production events
	logger *log.Logger
	// Channel for security events
	securityEvents chan securityEvent
}

// getConnList gets a connection list from the pool
func (h *Hub) getConnList() *[]*Conn {
	return h.connListPool.Get().(*[]*Conn)
}

// putConnList returns a connection list to the pool after clearing it
func (h *Hub) putConnList(conns *[]*Conn) {
	// Clear the slice but keep the underlying array
	*conns = (*conns)[:0]
	h.connListPool.Put(conns)
}

func (h *Hub) roomNameValidator() RoomNameValidator {
	if h == nil {
		return defaultRoomNameValidator
	}
	return h.config.RoomNameValidator
}

// simpleRateLimiter implements a basic token bucket rate limiter using only standard library.
// The mutex protects all fields; plain integer fields are used since the mutex already
// provides mutual exclusion (atomic types inside a mutex are redundant).
type simpleRateLimiter struct {
	rate       int // Tokens per second
	burst      int // Maximum burst size
	tokens     uint64
	lastRefill int64 // nanoseconds
	mu         sync.Mutex
}

// newRateLimiter creates a rate limiter that allows 'rate' events per second with burst capacity
func newRateLimiter(rate, burst int) *simpleRateLimiter {
	if rate <= 0 {
		return nil
	}
	if burst <= 0 {
		burst = rate
	}
	return &simpleRateLimiter{
		rate:       rate,
		burst:      burst,
		tokens:     uint64(burst * 1000), // Start with full burst capacity
		lastRefill: time.Now().UnixNano(),
	}
}

// allow checks if an event can proceed under rate limits
func (rl *simpleRateLimiter) allow() bool {
	if rl == nil {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UnixNano()
	elapsed := now - rl.lastRefill

	// Integer-only token refill: milliTokens per nanosecond = rate/1_000_000.
	// Avoids float64 arithmetic on the hot connection-accept path.
	tokensToAdd := elapsed * int64(rl.rate) / 1_000_000
	if tokensToAdd > 0 {
		maxTokens := uint64(rl.burst * 1000)
		rl.tokens += uint64(tokensToAdd)
		if rl.tokens > maxTokens {
			rl.tokens = maxTokens
		}
		rl.lastRefill = now
	}

	// Try to consume one token
	if rl.tokens >= 1000 {
		rl.tokens -= 1000
		return true
	}

	return false
}

type securityEvent struct {
	Timestamp time.Time
	Type      string
	Details   map[string]any
	Severity  string // "info", "warning", "error"
}

// HubConfig configures a hub instance with production features.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	config := websocket.HubConfig{
//		WorkerCount:            4,
//		JobQueueSize:           1024,
//		MaxRoomRegistrations: 10000,
//		MaxRoomConnections:     100,
//		EnableDebugLogging:     true,
//		RejectOnQueueFull:      true,
//		MaxConnectionRate:      100, // 100 connections per second
//		EnableSecurityMetrics:  true,
//	}
//	hub, err := websocket.NewHubWithConfigE(config)
type HubConfig struct {
	// WorkerCount is the number of worker goroutines for message delivery
	WorkerCount int

	// JobQueueSize is the size of the job queue
	JobQueueSize int

	// MaxRoomRegistrations is the maximum active room registrations allowed.
	// A single connection joined to N rooms contributes N registrations.
	MaxRoomRegistrations int

	// MaxRoomConnections is the maximum connections per room
	MaxRoomConnections int

	// EnableDebugLogging enables detailed logging for debugging
	EnableDebugLogging bool

	// RejectOnQueueFull determines behavior when broadcast queue is full
	// true: reject message and log error
	// false: drop message silently
	RejectOnQueueFull bool

	// MaxConnectionRate limits new connections per second
	// 0 means no limit
	MaxConnectionRate int

	// RoomNameValidator validates application room identifiers for public hub
	// room APIs. Nil uses the default validator.
	RoomNameValidator RoomNameValidator

	// EnableSecurityMetrics enables the internal security event monitor.
	// Metrics counters are always collected in HubMetrics.
	EnableSecurityMetrics bool

	// Logger is optional and caller-owned. When nil, the hub discards logs.
	Logger *log.Logger
}

// HubMetrics describes hub connection metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	metrics := hub.Metrics()
//	fmt.Printf("Active: %d, Rooms: %d\n", metrics.ActiveConnections, metrics.Rooms)
type HubMetrics struct {
	// ActiveConnections is the number of unique open connections registered in
	// at least one room.
	ActiveConnections int `json:"active_connections"`
	// RoomRegistrations is the sum of (connection x room) registrations.
	// A connection joined to N rooms contributes N to this count.
	RoomRegistrations    int    `json:"room_registrations"`
	Rooms                int    `json:"rooms"`
	AcceptedTotal        uint64 `json:"accepted_total"`
	RejectedTotal        uint64 `json:"rejected_total"`
	MaxRoomRegistrations int    `json:"max_room_registrations"`
	MaxRoomConnections   int    `json:"max_room_connections"`
	// Per-hub security/broadcast metrics
	BroadcastAttempted uint64 `json:"broadcast_attempted"`
	BroadcastEnqueued  uint64 `json:"broadcast_enqueued"`
	BroadcastSkipped   uint64 `json:"broadcast_skipped"`
	BroadcastDropped   uint64 `json:"broadcast_dropped"`
	SecurityRejections uint64 `json:"security_rejections"`
	InvalidWSKeys      uint64 `json:"invalid_ws_keys"`
	SuccessfulAuths    uint64 `json:"successful_auths"`
}

// NewHubWithConfigE creates a new WebSocket hub with custom configuration and
// returns explicit validation errors for invalid public inputs.
func NewHubWithConfigE(cfg HubConfig) (*Hub, error) {
	if cfg.WorkerCount < 0 {
		return nil, ErrNegativeWorkerCount
	}
	if cfg.JobQueueSize < 0 {
		return nil, ErrNegativeJobQueue
	}
	if cfg.MaxRoomRegistrations < 0 || cfg.MaxRoomConnections < 0 || cfg.MaxConnectionRate < 0 {
		return nil, ErrNegativeLimit
	}
	if cfg.WorkerCount == 0 {
		cfg.WorkerCount = 4
	}
	if cfg.JobQueueSize == 0 {
		cfg.JobQueueSize = 1024
	}
	return newHubWithNormalizedConfig(cfg)
}

func newHubWithNormalizedConfig(cfg HubConfig) (*Hub, error) {
	logger := cfg.Logger
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	h := &Hub{
		rooms:                make(map[string]map[*Conn]struct{}),
		jobQueue:             make(chan hubJob, cfg.JobQueueSize),
		workers:              cfg.WorkerCount,
		quit:                 make(chan struct{}),
		maxRoomRegistrations: cfg.MaxRoomRegistrations,
		maxRoomConns:         cfg.MaxRoomConnections,
		config:               cfg,
		securityEvents:       make(chan securityEvent, 100),
		logger:               logger,
	}

	// Initialize connection list pool
	h.connListPool = sync.Pool{
		New: func() any {
			// Pre-allocate with reasonable capacity
			conns := make([]*Conn, 0, 64)
			return &conns
		},
	}

	// Initialize rate limiter if configured
	if cfg.MaxConnectionRate > 0 {
		// Allow burst up to 2x the rate for handling spikes
		burst := cfg.MaxConnectionRate * 2
		h.rateLimiter = newRateLimiter(cfg.MaxConnectionRate, burst)
		if cfg.EnableDebugLogging {
			h.logger.Printf("Rate limiter initialized: %d connections/sec (burst: %d)", cfg.MaxConnectionRate, burst)
		}
	}

	h.startWorkers()
	h.startSecurityMonitor()
	return h, nil
}

func (h *Hub) startWorkers() {
	for i := 0; i < h.workers; i++ {
		h.wg.Add(1)
		go func(workerID int) {
			defer h.wg.Done()
			for {
				select {
				case j, ok := <-h.jobQueue:
					if !ok {
						return
					}
					// Write with error handling
					err := h.writeJob(j)
					if err != nil {
						if h.config.EnableDebugLogging {
							h.logger.Printf("Worker %d: failed to write to connection: %v", workerID, err)
						}
						if h.config.EnableSecurityMetrics {
							h.recordSecurityEvent("broadcast_error", map[string]any{
								"error":  err.Error(),
								"worker": workerID,
							}, "warning")
						}
					}
				case <-h.quit:
					// Drain remaining jobs so in-flight messages are not silently dropped.
					// No new sends arrive after Stop() sets stopped=true before closing quit,
					// so this loop terminates as soon as the buffered channel is empty.
					for {
						select {
						case j := <-h.jobQueue:
							_ = h.writeJob(j)
						default:
							return
						}
					}
				}
			}
		}(i)
	}
}

func (h *Hub) writeJob(j hubJob) error {
	if j.conn.sendBehavior != SendBlock || j.conn.sendTimeout > 0 {
		return j.conn.WriteMessage(j.op, j.data)
	}
	ctx, cancel := context.WithTimeout(context.Background(), hubWorkerDefaultSendTimeout)
	defer cancel()
	return j.conn.WriteMessageContext(ctx, j.op, j.data)
}

// startSecurityMonitor processes security events
func (h *Hub) startSecurityMonitor() {
	if !h.config.EnableSecurityMetrics {
		return
	}
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		for {
			select {
			case event := <-h.securityEvents:
				if h.config.EnableDebugLogging {
					h.logger.Printf("[%s] %s: %v", event.Severity, event.Type, event.Details)
				}
			case <-h.quit:
				// Drain remaining security events before exiting.
				for {
					select {
					case event := <-h.securityEvents:
						if h.config.EnableDebugLogging {
							h.logger.Printf("[%s] %s: %v", event.Severity, event.Type, event.Details)
						}
					default:
						return
					}
				}
			}
		}
	}()
}

// recordSecurityEvent records a security event
func (h *Hub) recordSecurityEvent(eventType string, details map[string]any, severity string) {
	select {
	case h.securityEvents <- securityEvent{
		Timestamp: time.Now(),
		Type:      eventType,
		Details:   details,
		Severity:  severity,
	}:
	default:
		// Channel full, drop event
	}
}

// Stop gracefully shuts down the hub.
//
// Stop is idempotent — calling it multiple times is safe.
//
// This method:
//   - Marks the hub as stopped so new broadcasts are rejected immediately
//   - Signals all worker goroutines to exit via the quit channel
//   - Waits for all workers and the security monitor to finish
//   - Clears room registrations and room membership metrics
//
// Note: h.jobQueue is intentionally NOT closed here. Workers exit via the quit
// channel, so closing the queue is unnecessary. More importantly, closing it
// while a concurrent BroadcastRoom/BroadcastAll is still executing a channel
// send would cause a panic ("send on closed channel"). The stopped flag prevents
// new sends after Stop is called.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	defer hub.Stop()
func (h *Hub) Stop() {
	if !h.stopped.CompareAndSwap(false, true) {
		return // already stopped
	}
	h.lifecycleMu.Lock()
	close(h.quit)
	h.lifecycleMu.Unlock()
	h.wg.Wait()
	h.clearRooms()
}

// Shutdown closes all open connections and then stops the hub.
//
// It collects every unique connection across all rooms, calls Close() on each
// one to tear down TCP immediately, and finally calls Stop() to drain in-flight
// jobs and shut down workers. Shutdown intentionally does not send WebSocket
// close frames because clean close-frame delivery can block on slow clients
// during process shutdown.
//
// ctx controls the overall deadline for the close loop. A nil context is treated
// as context.Background(). If the context is
// cancelled before all connections are closed, Shutdown returns ctx.Err()
// immediately after stopping the hub and clearing room registrations.
//
// Example:
//
//	import (
//	    "context"
//	    "time"
//	    "github.com/spcent/plumego/x/websocket"
//	)
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := hub.Shutdown(ctx); err != nil {
//	    log.Printf("shutdown incomplete: %v", err)
//	}
func (h *Hub) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Collect all unique connections under the read lock.
	h.mu.RLock()
	seen := make(map[*Conn]struct{})
	for _, rs := range h.rooms {
		for c := range rs {
			if c == nil {
				continue
			}
			seen[c] = struct{}{}
		}
	}
	h.mu.RUnlock()

	// Close each connection, respecting context cancellation.
	for c := range seen {
		select {
		case <-ctx.Done():
			h.Stop()
			return ctx.Err()
		default:
		}
		c.Close()
	}

	h.clearRooms()
	h.Stop()
	return nil
}

func (h *Hub) clearRooms() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rooms = make(map[string]map[*Conn]struct{})
	h.roomRegistrations.Store(0)
}

// RangeConns calls fn for each non-closed connection in room.
//
// The iteration stops when fn returns false or all connections are visited.
// fn is called outside the hub lock, so it is safe to perform I/O or call
// other hub methods inside it.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub.RangeConns("chat-room", func(c *websocket.Conn) bool {
//	    if userID, ok := c.GetMetadata("user_id"); ok {
//	        fmt.Println("connected:", userID)
//	    }
//	    return true // continue
//	})
func (h *Hub) RangeConns(room string, fn func(*Conn) bool) {
	if err := validateRoomName(room, h.roomNameValidator()); err != nil {
		return
	}
	// Take a snapshot under the read lock so fn is called without holding it.
	conns := h.getConnList()
	defer h.putConnList(conns)

	h.mu.RLock()
	for c := range h.rooms[room] {
		if c != nil && !c.IsClosed() {
			*conns = append(*conns, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range *conns {
		if c == nil {
			continue
		}
		if !fn(c) {
			break
		}
	}
}

// TryJoin registers a connection in a room when limits allow it.
//
// Returns ErrHubFull if the hub has reached its maximum connection limit.
// Returns ErrRoomFull if the room has reached its maximum connection limit.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	conn, err := websocket.NewConnE(...)
//	err := hub.TryJoin("chat-room", conn)
//	if err != nil {
//		if errors.Is(err, websocket.ErrHubFull) {
//			// Hub is at capacity
//		} else if errors.Is(err, websocket.ErrRoomFull) {
//			// Room is at capacity
//		}
//	}
func (h *Hub) TryJoin(room string, c *Conn) error {
	return h.tryJoin(room, c, h.roomNameValidator())
}

func (h *Hub) tryJoin(room string, c *Conn, validator RoomNameValidator) error {
	if h.stopped.Load() {
		h.rejected.Add(1)
		return ErrHubStopped
	}
	if err := validateRoomName(room, validator); err != nil {
		h.rejected.Add(1)
		return err
	}
	if c == nil {
		h.rejected.Add(1)
		return ErrNilConn
	}

	h.mu.RLock()
	if rs, ok := h.rooms[room]; ok {
		if _, exists := rs[c]; exists {
			h.mu.RUnlock()
			if h.config.EnableDebugLogging {
				h.logger.Printf("Connection already in room: %s", room)
			}
			return nil
		}
	}
	h.mu.RUnlock()

	// Rate limiting check (before acquiring lock for better performance)
	if h.rateLimiter != nil && !h.rateLimiter.allow() {
		h.rejected.Add(1)
		if h.config.EnableSecurityMetrics {
			h.recordSecurityEvent("rate_limit_exceeded", map[string]any{
				"room": room,
			}, "warning")
		}
		return ErrRateLimitExceeded
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if rs, ok := h.rooms[room]; ok {
		if _, exists := rs[c]; exists {
			if h.config.EnableDebugLogging {
				h.logger.Printf("Connection already in room: %s", room)
			}
			return nil
		}
	}

	// Fast path: check atomic counter for room registrations (O(1) instead of O(n))
	if h.maxRoomRegistrations > 0 {
		currentRegistrations := int(h.roomRegistrations.Load())
		if currentRegistrations >= h.maxRoomRegistrations {
			h.rejected.Add(1)
			if h.config.EnableSecurityMetrics {
				h.recordSecurityEvent("hub_full", map[string]any{
					"room":  room,
					"total": currentRegistrations,
				}, "warning")
			}
			return ErrHubFull
		}
	}

	rs, ok := h.rooms[room]
	if !ok {
		rs = make(map[*Conn]struct{})
		h.rooms[room] = rs
	}

	if h.maxRoomConns > 0 && len(rs) >= h.maxRoomConns {
		h.rejected.Add(1)
		if h.config.EnableSecurityMetrics {
			h.recordSecurityEvent("room_full", map[string]any{
				"room":  room,
				"count": len(rs),
			}, "warning")
		}
		return ErrRoomFull
	}

	rs[c] = struct{}{}
	h.roomRegistrations.Add(1) // Increment atomic counter
	h.accepted.Add(1)

	if h.config.EnableDebugLogging {
		h.logger.Printf("Connection joined room: %s (registrations: %d, room: %d)", room, h.roomRegistrations.Load(), len(rs))
	}

	return nil
}

// CanJoin checks if a room can accept another connection.
//
// This is useful for pre-checking capacity before attempting to join.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	if err := hub.CanJoin("chat-room"); err != nil {
//		// Room is full
//		return
//	}
//	// Room has capacity, proceed with join
func (h *Hub) CanJoin(room string) error {
	return h.canJoin(room, h.roomNameValidator())
}

func (h *Hub) canJoin(room string, validator RoomNameValidator) error {
	if h.stopped.Load() {
		return ErrHubStopped
	}
	if err := validateRoomName(room, validator); err != nil {
		return err
	}

	// Fast path: check atomic counter for room registrations
	if h.maxRoomRegistrations > 0 && int(h.roomRegistrations.Load()) >= h.maxRoomRegistrations {
		return ErrHubFull
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.maxRoomConns > 0 {
		if rs, ok := h.rooms[room]; ok && len(rs) >= h.maxRoomConns {
			return ErrRoomFull
		}
	}

	return nil
}

// Metrics returns a snapshot of hub metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	metrics := hub.Metrics()
//	fmt.Printf("Active connections: %d\n", metrics.ActiveConnections)
func (h *Hub) Metrics() HubMetrics {
	h.mu.RLock()
	rooms := len(h.rooms)
	seen := make(map[*Conn]struct{})
	for _, rs := range h.rooms {
		for c := range rs {
			if c != nil && !c.IsClosed() {
				seen[c] = struct{}{}
			}
		}
	}
	h.mu.RUnlock()

	return HubMetrics{
		ActiveConnections:    len(seen),
		RoomRegistrations:    int(h.roomRegistrations.Load()),
		Rooms:                rooms,
		AcceptedTotal:        h.accepted.Load(),
		RejectedTotal:        h.rejected.Load(),
		MaxRoomRegistrations: h.maxRoomRegistrations,
		MaxRoomConnections:   h.maxRoomConns,
		BroadcastAttempted:   h.broadcastAttempted.Load(),
		BroadcastEnqueued:    h.broadcastEnqueued.Load(),
		BroadcastSkipped:     h.broadcastSkipped.Load(),
		BroadcastDropped:     h.broadcastDropped.Load(),
		SecurityRejections:   h.securityRejections.Load(),
		InvalidWSKeys:        h.invalidWSKeys.Load(),
		SuccessfulAuths:      h.successfulAuths.Load(),
	}
}

// Leave room.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	conn, err := websocket.NewConnE(...)
//	hub.TryJoin("chat-room", conn)
//	// ... handle connection ...
//	hub.Leave("chat-room", conn)
func (h *Hub) Leave(room string, c *Conn) {
	if c == nil {
		return
	}
	if err := validateRoomName(room, h.roomNameValidator()); err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if rs, ok := h.rooms[room]; ok {
		if _, exists := rs[c]; exists {
			delete(rs, c)
			h.roomRegistrations.Add(^uint64(0)) // Decrement (add -1 in two's complement)
			if len(rs) == 0 {
				delete(h.rooms, room)
			}
		}
	}
}

// RemoveConn from all rooms.
//
// This is useful when a connection is closed and needs to be cleaned up from all rooms.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	conn, err := websocket.NewConnE(...)
//	hub.TryJoin("chat-room", conn)
//	hub.TryJoin("notifications-room", conn)
//	// Connection closed, remove from all rooms
//	hub.RemoveConn(conn)
func (h *Hub) RemoveConn(c *Conn) {
	if c == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	removedCount := 0
	for room, rs := range h.rooms {
		if _, ok := rs[c]; ok {
			delete(rs, c)
			removedCount++
			if len(rs) == 0 {
				delete(h.rooms, room)
			}
		}
	}
	// Decrement roomRegistrations by the number of rooms the connection was in
	if removedCount > 0 {
		h.roomRegistrations.Add(^uint64(removedCount - 1)) // Subtract removedCount
	}
}

// dispatchJobs enqueues send jobs for each connection in conns and tracks drops.
// label is used only for log/metric messages to identify the broadcast target.
func (h *Hub) dispatchJobs(conns []*Conn, op byte, data []byte, label string) BroadcastResult {
	result := BroadcastResult{Attempted: len(conns)}
	h.lifecycleMu.RLock()
	defer h.lifecycleMu.RUnlock()
	if h.stopped.Load() {
		result.Dropped = len(conns)
		result.Stopped = true
		h.broadcastAttempted.Add(uint64(result.Attempted))
		h.broadcastDropped.Add(uint64(result.Dropped))
		return result
	}
	for _, c := range conns {
		if c == nil || c.IsClosed() {
			result.Skipped++
			continue
		}
		select {
		case h.jobQueue <- hubJob{conn: c, op: op, data: data}:
			result.Enqueued++
		default:
			result.Dropped++
			if h.config.RejectOnQueueFull {
				if h.config.EnableDebugLogging {
					h.logger.Printf("broadcast queue full: dropped message to %s", label)
				}
				if h.config.EnableSecurityMetrics {
					h.recordSecurityEvent("broadcast_queue_full", map[string]any{
						"target":  label,
						"dropped": result.Dropped,
						"sent":    result.Enqueued,
					}, "error")
				}
			}
		}
	}
	h.broadcastAttempted.Add(uint64(result.Attempted))
	h.broadcastEnqueued.Add(uint64(result.Enqueued))
	h.broadcastSkipped.Add(uint64(result.Skipped))
	h.broadcastDropped.Add(uint64(result.Dropped))
	return result
}

func (h *Hub) recordSkippedBroadcast(skipped int) BroadcastResult {
	result := BroadcastResult{Skipped: skipped}
	if skipped > 0 {
		h.broadcastSkipped.Add(uint64(skipped))
	}
	return result
}

// BroadcastRoom enqueues jobs to jobQueue for workers to send.
//
// This method is thread-safe and can be called concurrently from multiple goroutines.
// Messages are delivered asynchronously via the worker pool.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	// Broadcast text message to all users in chat-room
//	hub.BroadcastRoom("chat-room", websocket.OpcodeText, []byte("Hello everyone!"))
//	// Broadcast binary data
//	hub.BroadcastRoom("chat-room", websocket.OpcodeBinary, []byte{0x01, 0x02, 0x03})
func (h *Hub) BroadcastRoom(room string, op byte, data []byte) {
	_ = h.TryBroadcastRoom(room, op, data)
}

// TryBroadcastRoom enqueues jobs for room members and returns the fanout result.
func (h *Hub) TryBroadcastRoom(room string, op byte, data []byte) BroadcastResult {
	return h.tryBroadcastRoom(room, op, data, h.roomNameValidator())
}

func (h *Hub) tryBroadcastRoom(room string, op byte, data []byte, validator RoomNameValidator) BroadcastResult {
	if h.stopped.Load() {
		return BroadcastResult{Stopped: true}
	}
	if err := validateRoomName(room, validator); err != nil {
		return BroadcastResult{Invalid: true}
	}

	connsList := h.getConnList()
	defer h.putConnList(connsList)

	h.mu.RLock()
	rs, ok := h.rooms[room]
	if !ok || len(rs) == 0 {
		h.mu.RUnlock()
		return BroadcastResult{}
	}
	if cap(*connsList) < len(rs) {
		*connsList = make([]*Conn, 0, len(rs))
	}
	for c := range rs {
		if c != nil && !c.IsClosed() {
			*connsList = append(*connsList, c)
		}
	}
	skipped := len(rs) - len(*connsList)
	h.mu.RUnlock()

	if len(*connsList) == 0 {
		return h.recordSkippedBroadcast(skipped)
	}
	result := h.dispatchJobs(*connsList, op, data, "room:"+room)
	result.Skipped += skipped
	if skipped > 0 {
		h.broadcastSkipped.Add(uint64(skipped))
	}
	return result
}

// BroadcastAll broadcasts to all clients.
//
// This is useful for system-wide announcements or notifications.
// Use with caution as it can generate significant network traffic.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	// Send system-wide notification
//	hub.BroadcastAll(websocket.OpcodeText, []byte("System maintenance in 5 minutes"))
func (h *Hub) BroadcastAll(op byte, data []byte) {
	_ = h.TryBroadcastAll(op, data)
}

// TryBroadcastAll enqueues jobs for all unique open connections and returns the fanout result.
func (h *Hub) TryBroadcastAll(op byte, data []byte) BroadcastResult {
	if h.stopped.Load() {
		return BroadcastResult{Stopped: true}
	}

	connsList := h.getConnList()
	defer h.putConnList(connsList)

	h.mu.RLock()
	// Fast path: single room needs no deduplication map.
	if len(h.rooms) == 1 {
		total := 0
		for _, rs := range h.rooms {
			total += len(rs)
			if cap(*connsList) < len(rs) {
				*connsList = make([]*Conn, 0, len(rs))
			}
			for c := range rs {
				if c != nil && !c.IsClosed() {
					*connsList = append(*connsList, c)
				}
			}
		}
		skipped := total - len(*connsList)
		h.mu.RUnlock()
		if len(*connsList) > 0 {
			result := h.dispatchJobs(*connsList, op, data, "all")
			result.Skipped += skipped
			if skipped > 0 {
				h.broadcastSkipped.Add(uint64(skipped))
			}
			return result
		}
		return h.recordSkippedBroadcast(skipped)
	}

	// Multi-room path: deduplicate connections that span multiple rooms.
	estimatedSize := 0
	for _, rs := range h.rooms {
		estimatedSize += len(rs)
	}
	if cap(*connsList) < estimatedSize {
		*connsList = make([]*Conn, 0, estimatedSize)
	}
	skipped := 0
	seen := make(map[*Conn]struct{}, estimatedSize)
	for _, rs := range h.rooms {
		for c := range rs {
			if c == nil || c.IsClosed() {
				skipped++
				continue
			}
			if _, dup := seen[c]; !dup {
				seen[c] = struct{}{}
				*connsList = append(*connsList, c)
			}
		}
	}
	h.mu.RUnlock()

	if len(*connsList) == 0 {
		return h.recordSkippedBroadcast(skipped)
	}
	result := h.dispatchJobs(*connsList, op, data, "all")
	result.Skipped += skipped
	if skipped > 0 {
		h.broadcastSkipped.Add(uint64(skipped))
	}
	return result
}

// GetRoomCount returns the number of connections in a room.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	count := hub.GetRoomCount("chat-room")
//	fmt.Printf("Chat room has %d connections\n", count)
func (h *Hub) GetRoomCount(room string) int {
	if err := validateRoomName(room, h.roomNameValidator()); err != nil {
		return 0
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if rs, ok := h.rooms[room]; ok {
		return len(rs)
	}
	return 0
}

// GetRoomRegistrationCount returns the number of connection-room registrations.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	total := hub.GetRoomRegistrationCount()
//	fmt.Printf("Room registrations: %d\n", total)
func (h *Hub) GetRoomRegistrationCount() int {
	return int(h.roomRegistrations.Load())
}

// GetActiveConnectionCount returns the number of unique open connections
// registered in at least one room.
func (h *Hub) GetActiveConnectionCount() int {
	return h.Metrics().ActiveConnections
}

// GetRooms returns a list of all room names.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	hub, err := websocket.NewHubWithConfigE(websocket.HubConfig{WorkerCount: 4, JobQueueSize: 1024})
//	rooms := hub.GetRooms()
//	for _, room := range rooms {
//		fmt.Printf("Room: %s, Connections: %d\n", room, hub.GetRoomCount(room))
//	}
func (h *Hub) GetRooms() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rooms := make([]string, 0, len(h.rooms))
	for room := range h.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
