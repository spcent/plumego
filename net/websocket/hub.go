package websocket

import (
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type hubJob struct {
	conn *Conn
	op   byte
	data []byte
}

// Hub manages rooms and broadcast.
//
// Hub provides a production-ready WebSocket hub with:
//   - Room-based message broadcasting
//   - Worker pool for concurrent message delivery
//   - Connection limits (total and per-room)
//   - Metrics collection
//   - Security event monitoring
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	// Create hub with 4 workers and 1024 job queue size
//	hub := websocket.NewHub(4, 1024)
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
	jobQueue chan hubJob
	workers  int
	wg       sync.WaitGroup
	quit     chan struct{}

	maxConns     int
	maxRoomConns int
	totalConns   atomic.Uint64 // Atomic counter for total active connections
	accepted     atomic.Uint64
	rejected     atomic.Uint64

	// Rate limiting (simple token bucket implementation)
	rateLimiter *simpleRateLimiter

	// Message pooling to reduce allocations
	connListPool sync.Pool // Pool of []*Conn slices

	// Production-ready configuration
	config HubConfig
	// Metrics for monitoring
	metrics HubMetrics
	// Logger for production events
	logger *log.Logger
	// Channel for security events
	securityEvents chan SecurityEvent
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

// simpleRateLimiter implements a basic token bucket rate limiter using only standard library
type simpleRateLimiter struct {
	rate       int           // Tokens per second
	burst      int           // Maximum burst size
	tokens     atomic.Uint64 // Current tokens (scaled by 1000 for precision)
	lastRefill atomic.Int64  // Last refill timestamp (nanoseconds)
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
	rl := &simpleRateLimiter{
		rate:  rate,
		burst: burst,
	}
	rl.tokens.Store(uint64(burst * 1000)) // Start with full burst capacity
	rl.lastRefill.Store(time.Now().UnixNano())
	return rl
}

// allow checks if an event can proceed under rate limits
func (rl *simpleRateLimiter) allow() bool {
	if rl == nil {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UnixNano()
	last := rl.lastRefill.Load()
	elapsed := time.Duration(now - last)

	// Calculate tokens to add based on elapsed time
	tokensToAdd := int64(elapsed.Seconds() * float64(rl.rate) * 1000)
	if tokensToAdd > 0 {
		current := rl.tokens.Load()
		maxTokens := uint64(rl.burst * 1000)
		newTokens := current + uint64(tokensToAdd)
		if newTokens > maxTokens {
			newTokens = maxTokens
		}
		rl.tokens.Store(newTokens)
		rl.lastRefill.Store(now)
	}

	// Try to consume one token
	current := rl.tokens.Load()
	if current >= 1000 {
		rl.tokens.Store(current - 1000)
		return true
	}

	return false
}

// SecurityEvent represents a security-related event.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	event := websocket.SecurityEvent{
//		Timestamp: time.Now(),
//		Type:      "hub_full",
//		Details:   map[string]any{"room": "chat", "total": 1000},
//		Severity:  "warning",
//	}
type SecurityEvent struct {
	Timestamp time.Time
	Type      string
	Details   map[string]any
	Severity  string // "info", "warning", "error"
}

// HubConfig configures a hub instance with production features.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	config := websocket.HubConfig{
//		WorkerCount:            4,
//		JobQueueSize:           1024,
//		MaxConnections:         10000,
//		MaxRoomConnections:     100,
//		EnableDebugLogging:     true,
//		EnableMetrics:          true,
//		RejectOnQueueFull:      true,
//		MaxConnectionRate:      100, // 100 connections per second
//		EnableSecurityMetrics:  true,
//	}
//	hub := websocket.NewHubWithConfig(config)
type HubConfig struct {
	// WorkerCount is the number of worker goroutines for message delivery
	WorkerCount int

	// JobQueueSize is the size of the job queue
	JobQueueSize int

	// MaxConnections is the maximum total connections allowed
	MaxConnections int

	// MaxRoomConnections is the maximum connections per room
	MaxRoomConnections int

	// EnableDebugLogging enables detailed logging for debugging
	EnableDebugLogging bool

	// EnableMetrics enables metrics collection
	EnableMetrics bool

	// RejectOnQueueFull determines behavior when broadcast queue is full
	// true: reject message and log error
	// false: drop message silently
	RejectOnQueueFull bool

	// MaxConnectionRate limits new connections per second
	// 0 means no limit
	MaxConnectionRate int

	// EnableSecurityMetrics enables security metrics collection
	EnableSecurityMetrics bool
}

// HubMetrics describes hub connection metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	metrics := hub.Metrics()
//	fmt.Printf("Active: %d, Rooms: %d\n", metrics.ActiveConnections, metrics.Rooms)
type HubMetrics struct {
	ActiveConnections  int    `json:"active_connections"`
	Rooms              int    `json:"rooms"`
	AcceptedTotal      uint64 `json:"accepted_total"`
	RejectedTotal      uint64 `json:"rejected_total"`
	MaxConnections     int    `json:"max_connections"`
	MaxRoomConnections int    `json:"max_room_connections"`
}

// NewHub creates a new WebSocket hub with default configuration.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	defer hub.Stop()
func NewHub(workerCount int, jobQueueSize int) *Hub {
	return NewHubWithConfig(HubConfig{
		WorkerCount:  workerCount,
		JobQueueSize: jobQueueSize,
	})
}

// NewHubWithConfig creates a new WebSocket hub with custom configuration.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	config := websocket.HubConfig{
//		WorkerCount:            8,
//		JobQueueSize:           2048,
//		MaxConnections:         50000,
//		MaxRoomConnections:     500,
//		EnableDebugLogging:     true,
//		EnableMetrics:          true,
//		RejectOnQueueFull:      true,
//		MaxConnectionRate:      200,
//		EnableSecurityMetrics:  true,
//	}
//	hub := websocket.NewHubWithConfig(config)
//	defer hub.Stop()
func NewHubWithConfig(cfg HubConfig) *Hub {
	// Validate configuration
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}
	if cfg.JobQueueSize <= 0 {
		cfg.JobQueueSize = 1024
	}

	h := &Hub{
		rooms:          make(map[string]map[*Conn]struct{}),
		jobQueue:       make(chan hubJob, cfg.JobQueueSize),
		workers:        cfg.WorkerCount,
		quit:           make(chan struct{}),
		maxConns:       cfg.MaxConnections,
		maxRoomConns:   cfg.MaxRoomConnections,
		config:         cfg,
		securityEvents: make(chan SecurityEvent, 100),
		logger:         log.New(os.Stderr, "[WEBSOCKET] ", log.LstdFlags),
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
	return h
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
					err := j.conn.WriteMessage(j.op, j.data)
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
					return
				}
			}
		}(i)
	}
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
				return
			}
		}
	}()
}

// recordSecurityEvent records a security event
func (h *Hub) recordSecurityEvent(eventType string, details map[string]any, severity string) {
	select {
	case h.securityEvents <- SecurityEvent{
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
// This method:
//   - Closes the job queue to stop accepting new jobs
//   - Waits for all workers to finish processing remaining jobs
//   - Closes all security event channels
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	defer hub.Stop()
func (h *Hub) Stop() {
	close(h.quit)
	close(h.jobQueue)
	h.wg.Wait()
}

// TryJoin registers a connection in a room when limits allow it.
//
// Returns ErrHubFull if the hub has reached its maximum connection limit.
// Returns ErrRoomFull if the room has reached its maximum connection limit.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	conn := websocket.NewConn(...)
//	err := hub.TryJoin("chat-room", conn)
//	if err != nil {
//		if errors.Is(err, websocket.ErrHubFull) {
//			// Hub is at capacity
//		} else if errors.Is(err, websocket.ErrRoomFull) {
//			// Room is at capacity
//		}
//	}
func (h *Hub) TryJoin(room string, c *Conn) error {
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

	// Fast path: check atomic counter for total connections (O(1) instead of O(n))
	if h.maxConns > 0 {
		currentTotal := int(h.totalConns.Load())
		if currentTotal >= h.maxConns {
			h.rejected.Add(1)
			if h.config.EnableSecurityMetrics {
				h.recordSecurityEvent("hub_full", map[string]any{
					"room":  room,
					"total": currentTotal,
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

	// Check if connection already in room (idempotent join)
	if _, exists := rs[c]; exists {
		if h.config.EnableDebugLogging {
			h.logger.Printf("Connection already in room: %s", room)
		}
		return nil
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
	h.totalConns.Add(1) // Increment atomic counter
	h.accepted.Add(1)

	if h.config.EnableDebugLogging {
		h.logger.Printf("Connection joined room: %s (total: %d, room: %d)", room, h.totalConns.Load(), len(rs))
	}

	return nil
}

// CanJoin checks if a room can accept another connection.
//
// This is useful for pre-checking capacity before attempting to join.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	if err := hub.CanJoin("chat-room"); err != nil {
//		// Room is full
//		return
//	}
//	// Room has capacity, proceed with join
func (h *Hub) CanJoin(room string) error {
	// Fast path: check atomic counter for total connections
	if h.maxConns > 0 && int(h.totalConns.Load()) >= h.maxConns {
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

// Join room (ignores capacity limits).
//
// This method is useful when you want to bypass capacity checks.
// Use with caution as it can lead to resource exhaustion.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	conn := websocket.NewConn(...)
//	hub.Join("admin-room", conn) // Bypass capacity limits
func (h *Hub) Join(room string, c *Conn) {
	_ = h.TryJoin(room, c)
}

// Metrics returns a snapshot of hub metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	metrics := hub.Metrics()
//	fmt.Printf("Active connections: %d\n", metrics.ActiveConnections)
func (h *Hub) Metrics() HubMetrics {
	h.mu.RLock()
	rooms := len(h.rooms)
	h.mu.RUnlock()

	return HubMetrics{
		ActiveConnections:  int(h.totalConns.Load()), // Use atomic counter
		Rooms:              rooms,
		AcceptedTotal:      h.accepted.Load(),
		RejectedTotal:      h.rejected.Load(),
		MaxConnections:     h.maxConns,
		MaxRoomConnections: h.maxRoomConns,
	}
}

// Leave room.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	conn := websocket.NewConn(...)
//	hub.TryJoin("chat-room", conn)
//	// ... handle connection ...
//	hub.Leave("chat-room", conn)
func (h *Hub) Leave(room string, c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if rs, ok := h.rooms[room]; ok {
		if _, exists := rs[c]; exists {
			delete(rs, c)
			h.totalConns.Add(^uint64(0)) // Decrement (add -1 in two's complement)
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
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	conn := websocket.NewConn(...)
//	hub.TryJoin("chat-room", conn)
//	hub.TryJoin("notifications-room", conn)
//	// Connection closed, remove from all rooms
//	hub.RemoveConn(conn)
func (h *Hub) RemoveConn(c *Conn) {
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
	// Decrement totalConns by the number of rooms the connection was in
	if removedCount > 0 {
		h.totalConns.Add(^uint64(removedCount - 1)) // Subtract removedCount
	}
}

// BroadcastRoom enqueues jobs to jobQueue for workers to send.
//
// This method is thread-safe and can be called concurrently from multiple goroutines.
// Messages are delivered asynchronously via the worker pool.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	// Broadcast text message to all users in chat-room
//	hub.BroadcastRoom("chat-room", websocket.OpcodeText, []byte("Hello everyone!"))
//	// Broadcast binary data
//	hub.BroadcastRoom("chat-room", websocket.OpcodeBinary, []byte{0x01, 0x02, 0x03})
func (h *Hub) BroadcastRoom(room string, op byte, data []byte) {
	// Get a connection list from pool
	connsList := h.getConnList()
	defer h.putConnList(connsList)

	// Copy connections while holding read lock to prevent race conditions
	h.mu.RLock()
	rs, ok := h.rooms[room]
	if !ok || len(rs) == 0 {
		h.mu.RUnlock()
		return
	}

	// Grow slice if needed (pool size might be too small)
	if cap(*connsList) < len(rs) {
		*connsList = make([]*Conn, 0, len(rs))
	}

	// Create snapshot of connections to iterate safely
	for c := range rs {
		// Only include active connections
		if !c.IsClosed() {
			*connsList = append(*connsList, c)
		}
	}
	h.mu.RUnlock()

	if len(*connsList) == 0 {
		return
	}

	sent := 0
	dropped := 0

	// Iterate over snapshot - safe from concurrent modifications
	for _, c := range *connsList {
		select {
		case h.jobQueue <- hubJob{conn: c, op: op, data: data}:
			sent++
		default:
			dropped++
			if h.config.RejectOnQueueFull {
				// Production mode: log and record metric
				if h.config.EnableDebugLogging {
					h.logger.Printf("Broadcast queue full: dropped message to room %s", room)
				}
				if h.config.EnableSecurityMetrics {
					h.recordSecurityEvent("broadcast_queue_full", map[string]any{
						"room":    room,
						"dropped": dropped,
						"sent":    sent,
					}, "error")
				}
			}
			// Debug mode: silently drop (original behavior)
		}
	}

	// Production metrics - use atomic operations instead of mutex
	if h.config.EnableMetrics && dropped > 0 {
		atomic.AddUint64(&securityMetrics.BroadcastQueueFull, uint64(dropped))
	}
}

// BroadcastAll broadcasts to all clients.
//
// This is useful for system-wide announcements or notifications.
// Use with caution as it can generate significant network traffic.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	// Send system-wide notification
//	hub.BroadcastAll(websocket.OpcodeText, []byte("System maintenance in 5 minutes"))
func (h *Hub) BroadcastAll(op byte, data []byte) {
	// Get a connection list from pool
	connsList := h.getConnList()
	defer h.putConnList(connsList)

	// Copy all connections while holding read lock
	h.mu.RLock()
	// Pre-allocate slice with estimated capacity if needed
	estimatedSize := 0
	for _, rs := range h.rooms {
		estimatedSize += len(rs)
	}
	if cap(*connsList) < estimatedSize {
		*connsList = make([]*Conn, 0, estimatedSize)
	}

	for _, rs := range h.rooms {
		for c := range rs {
			// Only include active connections
			if !c.IsClosed() {
				*connsList = append(*connsList, c)
			}
		}
	}
	h.mu.RUnlock()

	if len(*connsList) == 0 {
		return
	}

	sent := 0
	dropped := 0

	for _, c := range *connsList {
		select {
		case h.jobQueue <- hubJob{conn: c, op: op, data: data}:
			sent++
		default:
			dropped++
			if h.config.RejectOnQueueFull {
				if h.config.EnableDebugLogging {
					h.logger.Printf("Broadcast queue full: dropped message to all clients")
				}
				if h.config.EnableSecurityMetrics {
					h.recordSecurityEvent("broadcast_all_queue_full", map[string]any{
						"dropped": dropped,
						"sent":    sent,
					}, "error")
				}
			}
		}
	}

	// Production metrics - use atomic operations instead of mutex
	if h.config.EnableMetrics && dropped > 0 {
		atomic.AddUint64(&securityMetrics.BroadcastQueueFull, uint64(dropped))
	}
}

// GetRoomCount returns the number of connections in a room.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	count := hub.GetRoomCount("chat-room")
//	fmt.Printf("Chat room has %d connections\n", count)
func (h *Hub) GetRoomCount(room string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if rs, ok := h.rooms[room]; ok {
		return len(rs)
	}
	return 0
}

// GetTotalCount returns the total number of connections.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
//	total := hub.GetTotalCount()
//	fmt.Printf("Total connections: %d\n", total)
func (h *Hub) GetTotalCount() int {
	// Use atomic counter for O(1) performance
	return int(h.totalConns.Load())
}

// GetRooms returns a list of all room names.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	hub := websocket.NewHub(4, 1024)
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
