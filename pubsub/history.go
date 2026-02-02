package pubsub

import (
	"sync"
	"time"
)

// messageHistory stores recent messages for replay functionality.
// It uses a circular buffer with optional TTL-based expiration.
type messageHistory struct {
	mu       sync.RWMutex
	messages []historicalMessage
	head     int // oldest message index
	tail     int // next write index
	count    int // current message count
	capacity int // maximum messages to retain
	sequence uint64
}

// historicalMessage wraps a message with metadata for history tracking.
type historicalMessage struct {
	msg      Message
	sequence uint64
	addedAt  time.Time
}

// newMessageHistory creates a new message history with the given capacity.
func newMessageHistory(capacity int) *messageHistory {
	if capacity <= 0 {
		return nil
	}
	return &messageHistory{
		messages: make([]historicalMessage, capacity),
		capacity: capacity,
	}
}

// Add adds a message to the history.
// If the history is full, the oldest message is evicted.
func (h *messageHistory) Add(msg Message) uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sequence++
	seq := h.sequence

	hm := historicalMessage{
		msg:      msg,
		sequence: seq,
		addedAt:  time.Now(),
	}

	if h.count == h.capacity {
		// Overwrite oldest
		h.head = (h.head + 1) % h.capacity
	} else {
		h.count++
	}

	h.messages[h.tail] = hm
	h.tail = (h.tail + 1) % h.capacity

	return seq
}

// GetAll returns all messages in the history (oldest first).
func (h *messageHistory) GetAll() []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.count == 0 {
		return nil
	}

	result := make([]Message, h.count)
	for i := 0; i < h.count; i++ {
		idx := (h.head + i) % h.capacity
		result[i] = h.messages[idx].msg
	}
	return result
}

// GetSince returns messages added after the given sequence number.
func (h *messageHistory) GetSince(sequence uint64) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.count == 0 {
		return nil
	}

	var result []Message
	for i := 0; i < h.count; i++ {
		idx := (h.head + i) % h.capacity
		if h.messages[idx].sequence > sequence {
			result = append(result, h.messages[idx].msg)
		}
	}
	return result
}

// GetLast returns the last N messages (newest last).
func (h *messageHistory) GetLast(n int) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.count == 0 || n <= 0 {
		return nil
	}

	if n > h.count {
		n = h.count
	}

	result := make([]Message, n)
	startIdx := h.count - n
	for i := 0; i < n; i++ {
		idx := (h.head + startIdx + i) % h.capacity
		result[i] = h.messages[idx].msg
	}
	return result
}

// GetWithTTL returns messages not older than the given TTL.
func (h *messageHistory) GetWithTTL(ttl time.Duration) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.count == 0 {
		return nil
	}

	cutoff := time.Now().Add(-ttl)
	var result []Message

	for i := 0; i < h.count; i++ {
		idx := (h.head + i) % h.capacity
		if h.messages[idx].addedAt.After(cutoff) {
			result = append(result, h.messages[idx].msg)
		}
	}
	return result
}

// Len returns the current number of messages in history.
func (h *messageHistory) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

// Cap returns the capacity of the history.
func (h *messageHistory) Cap() int {
	return h.capacity
}

// CurrentSequence returns the current sequence number.
func (h *messageHistory) CurrentSequence() uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sequence
}

// Clear removes all messages from history.
func (h *messageHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i := range h.messages {
		h.messages[i] = historicalMessage{}
	}
	h.head = 0
	h.tail = 0
	h.count = 0
}

// topicHistory manages message history per topic.
type topicHistory struct {
	mu        sync.RWMutex
	histories map[string]*messageHistory
	config    HistoryConfig
}

// HistoryConfig configures topic history behavior.
type HistoryConfig struct {
	// DefaultRetention is the default number of messages to retain per topic
	DefaultRetention int

	// MaxRetention is the maximum allowed retention (0 = no limit)
	MaxRetention int

	// DefaultTTL is the default message TTL (0 = no expiration)
	DefaultTTL time.Duration

	// CleanupInterval is how often to run TTL cleanup (0 = disabled)
	CleanupInterval time.Duration
}

// DefaultHistoryConfig returns default history configuration.
func DefaultHistoryConfig() HistoryConfig {
	return HistoryConfig{
		DefaultRetention: 100,
		MaxRetention:     10000,
		DefaultTTL:       0,
		CleanupInterval:  0,
	}
}

// newTopicHistory creates a new topic history manager.
func newTopicHistory(config HistoryConfig) *topicHistory {
	return &topicHistory{
		histories: make(map[string]*messageHistory),
		config:    config,
	}
}

// GetOrCreate gets or creates history for a topic.
func (th *topicHistory) GetOrCreate(topic string, retention int) *messageHistory {
	th.mu.Lock()
	defer th.mu.Unlock()

	if h, ok := th.histories[topic]; ok {
		return h
	}

	if retention <= 0 {
		retention = th.config.DefaultRetention
	}
	if th.config.MaxRetention > 0 && retention > th.config.MaxRetention {
		retention = th.config.MaxRetention
	}

	h := newMessageHistory(retention)
	th.histories[topic] = h
	return h
}

// Get gets history for a topic (returns nil if not exists).
func (th *topicHistory) Get(topic string) *messageHistory {
	th.mu.RLock()
	defer th.mu.RUnlock()
	return th.histories[topic]
}

// Delete removes history for a topic.
func (th *topicHistory) Delete(topic string) {
	th.mu.Lock()
	defer th.mu.Unlock()
	delete(th.histories, topic)
}

// Topics returns all topics with history.
func (th *topicHistory) Topics() []string {
	th.mu.RLock()
	defer th.mu.RUnlock()

	topics := make([]string, 0, len(th.histories))
	for topic := range th.histories {
		topics = append(topics, topic)
	}
	return topics
}

// Stats returns history statistics.
func (th *topicHistory) Stats() map[string]HistoryStats {
	th.mu.RLock()
	defer th.mu.RUnlock()

	stats := make(map[string]HistoryStats, len(th.histories))
	for topic, h := range th.histories {
		stats[topic] = HistoryStats{
			Count:    h.Len(),
			Capacity: h.Cap(),
			Sequence: h.CurrentSequence(),
		}
	}
	return stats
}

// HistoryStats contains statistics for a topic's history.
type HistoryStats struct {
	Count    int
	Capacity int
	Sequence uint64
}
