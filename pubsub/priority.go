package pubsub

import (
	"container/heap"
	"sync"
)

// Priority levels (lower number = higher priority)
const (
	PriorityHighest = 0
	PriorityHigh    = 1
	PriorityNormal  = 2
	PriorityLow     = 3
	PriorityLowest  = 4
	PriorityDefault = PriorityNormal
)

// PriorityMessage wraps a message with priority for queue ordering.
type PriorityMessage struct {
	Message  Message
	Priority int
	Sequence uint64 // for FIFO within same priority
	index    int    // heap index
}

// priorityQueue implements heap.Interface for priority-based message delivery.
type priorityQueue []*PriorityMessage

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// Lower priority number = higher priority
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority < pq[j].Priority
	}
	// Same priority: FIFO by sequence
	return pq[i].Sequence < pq[j].Sequence
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*PriorityMessage)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// PriorityBuffer is a thread-safe priority queue for messages.
type PriorityBuffer struct {
	mu       sync.Mutex
	pq       priorityQueue
	capacity int
	sequence uint64
	notify   chan struct{}
	closed   bool
}

// NewPriorityBuffer creates a new priority buffer.
func NewPriorityBuffer(capacity int) *PriorityBuffer {
	if capacity <= 0 {
		capacity = 16
	}
	pb := &PriorityBuffer{
		pq:       make(priorityQueue, 0, capacity),
		capacity: capacity,
		notify:   make(chan struct{}, 1),
	}
	heap.Init(&pb.pq)
	return pb
}

// Push adds a message to the buffer with the given priority.
// Returns (dropped message, true) if capacity was exceeded.
func (pb *PriorityBuffer) Push(msg Message, priority int) (*Message, bool) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.closed {
		return nil, false
	}

	pb.sequence++
	pm := &PriorityMessage{
		Message:  msg,
		Priority: priority,
		Sequence: pb.sequence,
	}

	var dropped *Message

	if len(pb.pq) >= pb.capacity {
		// Drop lowest priority message
		// Find the lowest priority (highest number)
		lowestIdx := -1
		lowestPri := -1
		lowestSeq := uint64(0)

		for i, m := range pb.pq {
			if m.Priority > lowestPri || (m.Priority == lowestPri && m.Sequence < lowestSeq) {
				lowestIdx = i
				lowestPri = m.Priority
				lowestSeq = m.Sequence
			}
		}

		// Only drop if new message has higher or equal priority
		if lowestIdx >= 0 && priority <= lowestPri {
			droppedMsg := pb.pq[lowestIdx].Message
			dropped = &droppedMsg
			heap.Remove(&pb.pq, lowestIdx)
		} else if lowestIdx >= 0 {
			// New message has lower priority, drop it instead
			return &msg, true
		}
	}

	heap.Push(&pb.pq, pm)

	// Notify waiting consumers
	select {
	case pb.notify <- struct{}{}:
	default:
	}

	return dropped, dropped != nil
}

// Pop removes and returns the highest priority message.
func (pb *PriorityBuffer) Pop() (Message, int, bool) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if len(pb.pq) == 0 || pb.closed {
		return Message{}, 0, false
	}

	pm := heap.Pop(&pb.pq).(*PriorityMessage)
	return pm.Message, pm.Priority, true
}

// Peek returns the highest priority message without removing it.
func (pb *PriorityBuffer) Peek() (Message, int, bool) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if len(pb.pq) == 0 || pb.closed {
		return Message{}, 0, false
	}

	pm := pb.pq[0]
	return pm.Message, pm.Priority, true
}

// Len returns the number of messages in the buffer.
func (pb *PriorityBuffer) Len() int {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	return len(pb.pq)
}

// Cap returns the capacity of the buffer.
func (pb *PriorityBuffer) Cap() int {
	return pb.capacity
}

// Notify returns a channel that signals when messages are available.
func (pb *PriorityBuffer) Notify() <-chan struct{} {
	return pb.notify
}

// Close closes the buffer.
func (pb *PriorityBuffer) Close() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.closed {
		return
	}
	pb.closed = true
	close(pb.notify)
}

// IsClosed returns whether the buffer is closed.
func (pb *PriorityBuffer) IsClosed() bool {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	return pb.closed
}

// Drain removes all messages and returns them in priority order.
func (pb *PriorityBuffer) Drain() []Message {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	msgs := make([]Message, len(pb.pq))
	for i := 0; len(pb.pq) > 0; i++ {
		pm := heap.Pop(&pb.pq).(*PriorityMessage)
		msgs[i] = pm.Message
	}
	return msgs
}

// Stats returns buffer statistics.
func (pb *PriorityBuffer) Stats() PriorityBufferStats {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	stats := PriorityBufferStats{
		Size:       len(pb.pq),
		Capacity:   pb.capacity,
		ByPriority: make(map[int]int),
	}

	for _, pm := range pb.pq {
		stats.ByPriority[pm.Priority]++
	}

	return stats
}

// PriorityBufferStats contains buffer statistics.
type PriorityBufferStats struct {
	Size       int
	Capacity   int
	ByPriority map[int]int
}

// GetMessagePriority extracts priority from message metadata.
// Returns PriorityDefault if not set.
func GetMessagePriority(msg Message) int {
	if msg.Meta == nil {
		return PriorityDefault
	}

	priStr, ok := msg.Meta["X-Priority"]
	if !ok {
		return PriorityDefault
	}

	switch priStr {
	case "highest", "0":
		return PriorityHighest
	case "high", "1":
		return PriorityHigh
	case "normal", "2":
		return PriorityNormal
	case "low", "3":
		return PriorityLow
	case "lowest", "4":
		return PriorityLowest
	default:
		return PriorityDefault
	}
}

// SetMessagePriority sets priority in message metadata.
func SetMessagePriority(msg *Message, priority int) {
	if msg.Meta == nil {
		msg.Meta = make(map[string]string)
	}

	var priStr string
	switch priority {
	case PriorityHighest:
		priStr = "highest"
	case PriorityHigh:
		priStr = "high"
	case PriorityNormal:
		priStr = "normal"
	case PriorityLow:
		priStr = "low"
	case PriorityLowest:
		priStr = "lowest"
	default:
		priStr = "normal"
	}

	msg.Meta["X-Priority"] = priStr
}
