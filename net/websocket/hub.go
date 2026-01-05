package websocket

import (
	"sync"
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
