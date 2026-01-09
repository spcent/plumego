package websocket

import (
	"errors"
	"sync"
	"sync/atomic"
)

type hubJob struct {
	conn *Conn
	op   byte
	data []byte
}

// Hub manages rooms and broadcast
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
	accepted     atomic.Uint64
	rejected     atomic.Uint64
}

// HubConfig configures a hub instance.
type HubConfig struct {
	WorkerCount        int
	JobQueueSize       int
	MaxConnections     int
	MaxRoomConnections int
}

// HubMetrics describes hub connection metrics.
type HubMetrics struct {
	ActiveConnections  int    `json:"active_connections"`
	Rooms              int    `json:"rooms"`
	AcceptedTotal      uint64 `json:"accepted_total"`
	RejectedTotal      uint64 `json:"rejected_total"`
	MaxConnections     int    `json:"max_connections"`
	MaxRoomConnections int    `json:"max_room_connections"`
}

var (
	ErrHubFull  = errors.New("websocket hub at capacity")
	ErrRoomFull = errors.New("websocket room at capacity")
)

func NewHub(workerCount int, jobQueueSize int) *Hub {
	return NewHubWithConfig(HubConfig{
		WorkerCount:  workerCount,
		JobQueueSize: jobQueueSize,
	})
}

func NewHubWithConfig(cfg HubConfig) *Hub {
	h := &Hub{
		rooms:        make(map[string]map[*Conn]struct{}),
		jobQueue:     make(chan hubJob, cfg.JobQueueSize),
		workers:      cfg.WorkerCount,
		quit:         make(chan struct{}),
		maxConns:     cfg.MaxConnections,
		maxRoomConns: cfg.MaxRoomConnections,
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

// TryJoin registers a connection in a room when limits allow it.
func (h *Hub) TryJoin(room string, c *Conn) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.maxConns > 0 {
		total := 0
		for _, rs := range h.rooms {
			total += len(rs)
		}
		if total >= h.maxConns {
			h.rejected.Add(1)
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
		return ErrRoomFull
	}

	rs[c] = struct{}{}
	h.accepted.Add(1)
	return nil
}

// CanJoin checks if a room can accept another connection.
func (h *Hub) CanJoin(room string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.maxConns > 0 {
		total := 0
		for _, rs := range h.rooms {
			total += len(rs)
		}
		if total >= h.maxConns {
			return ErrHubFull
		}
	}

	if h.maxRoomConns > 0 {
		if rs, ok := h.rooms[room]; ok && len(rs) >= h.maxRoomConns {
			return ErrRoomFull
		}
	}

	return nil
}

// Join room (ignores capacity limits).
func (h *Hub) Join(room string, c *Conn) {
	_ = h.TryJoin(room, c)
}

// Metrics returns a snapshot of hub metrics.
func (h *Hub) Metrics() HubMetrics {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := len(h.rooms)
	active := 0
	for _, rs := range h.rooms {
		active += len(rs)
	}

	return HubMetrics{
		ActiveConnections:  active,
		Rooms:              rooms,
		AcceptedTotal:      h.accepted.Load(),
		RejectedTotal:      h.rejected.Load(),
		MaxConnections:     h.maxConns,
		MaxRoomConnections: h.maxRoomConns,
	}
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

// GetRoomCount returns the number of connections in a room
func (h *Hub) GetRoomCount(room string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if rs, ok := h.rooms[room]; ok {
		return len(rs)
	}
	return 0
}

// GetTotalCount returns the total number of connections
func (h *Hub) GetTotalCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, rs := range h.rooms {
		count += len(rs)
	}
	return count
}

// GetRooms returns a list of all room names
func (h *Hub) GetRooms() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rooms := make([]string, 0, len(h.rooms))
	for room := range h.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
