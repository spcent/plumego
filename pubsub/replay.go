package pubsub

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Replay errors
var (
	ErrReplayClosed      = errors.New("replay store is closed")
	ErrMessageNotFound   = errors.New("message not found in replay store")
	ErrInvalidTimeRange  = errors.New("invalid time range")
	ErrArchiveNotEnabled = errors.New("archiving is not enabled")
)

// ReplayConfig configures the message replay store
type ReplayConfig struct {
	// Enabled enables message replay
	Enabled bool

	// MaxMessages is the maximum number of messages to retain in memory
	MaxMessages int

	// RetentionPeriod is how long to keep messages before archiving
	RetentionPeriod time.Duration

	// ArchiveEnabled enables archiving old messages to disk
	ArchiveEnabled bool

	// ArchiveDir is the directory for archived messages
	ArchiveDir string

	// ArchiveCompression enables gzip compression for archives
	ArchiveCompression bool

	// ArchiveInterval is how often to run archiving
	ArchiveInterval time.Duration

	// TopicFilter restricts replay to specific topics (empty = all topics)
	TopicFilter []string

	// IncludeData whether to store full message data (may consume memory)
	IncludeData bool
}

// DefaultReplayConfig returns default replay configuration
func DefaultReplayConfig() ReplayConfig {
	return ReplayConfig{
		Enabled:            true,
		MaxMessages:        100000,
		RetentionPeriod:    24 * time.Hour,
		ArchiveEnabled:     true,
		ArchiveDir:         "./replay-archives",
		ArchiveCompression: true,
		ArchiveInterval:    1 * time.Hour,
		TopicFilter:        nil,
		IncludeData:        true,
	}
}

// ReplayMessage represents a stored message for replay
type ReplayMessage struct {
	ID        string
	Topic     string
	Data      any
	Timestamp time.Time
	Size      int
	Metadata  map[string]string

	// Replay metadata
	ReplayCount int
	LastReplay  time.Time
	IsArchived  bool
}

// ReplayQuery specifies query filters for message replay
type ReplayQuery struct {
	// Time range
	StartTime time.Time
	EndTime   time.Time

	// Topic filters
	Topics []string

	// Message IDs
	MessageIDs []string

	// Limit and offset for pagination
	Limit  int
	Offset int

	// Include archived messages
	IncludeArchived bool

	// Sort order (timestamp, topic, size)
	SortBy    string
	Ascending bool
}

// ReplayStats provides statistics about the replay store
type ReplayStats struct {
	TotalMessages    atomic.Uint64
	ArchivedMessages atomic.Uint64
	ReplayedMessages atomic.Uint64
	TotalSize        atomic.Uint64
	OldestMessage    time.Time
	NewestMessage    time.Time
	TopicCount       int
}

// ReplayStore manages message storage and replay
type ReplayStore struct {
	ps     *InProcPubSub
	config ReplayConfig

	// In-memory storage
	messages   map[string]*ReplayMessage
	messagesMu sync.RWMutex

	// Topic index
	topicIndex   map[string][]string // topic -> message IDs
	topicIndexMu sync.RWMutex

	// Time index (sorted by timestamp)
	timeIndex   []*ReplayMessage
	timeIndexMu sync.RWMutex

	// Archived messages
	archived   []*ReplayMessage
	archivedMu sync.RWMutex

	// Statistics
	stats ReplayStats

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed atomic.Bool
}

// NewReplayStore creates a new message replay store
func NewReplayStore(ps *InProcPubSub, config ReplayConfig) (*ReplayStore, error) {
	if !config.Enabled {
		return nil, errors.New("replay is not enabled")
	}

	ctx, cancel := context.WithCancel(context.Background())

	rs := &ReplayStore{
		ps:         ps,
		config:     config,
		messages:   make(map[string]*ReplayMessage),
		topicIndex: make(map[string][]string),
		timeIndex:  make([]*ReplayMessage, 0),
		archived:   make([]*ReplayMessage, 0),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Create archive directory if needed
	if config.ArchiveEnabled {
		if err := os.MkdirAll(config.ArchiveDir, 0755); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create archive directory: %w", err)
		}
	}

	// Start background workers
	rs.wg.Add(1)
	go rs.archiveWorker()

	// Subscribe to all topics for capturing
	if err := rs.startCapturing(); err != nil {
		cancel()
		return nil, err
	}

	return rs, nil
}

// startCapturing starts capturing published messages
func (rs *ReplayStore) startCapturing() error {
	// Subscribe to topics based on filter
	topics := rs.config.TopicFilter
	usePattern := false

	if len(topics) == 0 {
		// Subscribe to all topics using wildcard
		topics = []string{"*"}
		usePattern = true
	}

	for _, topic := range topics {
		var sub Subscription
		var err error

		if usePattern {
			sub, err = rs.ps.SubscribePattern(topic, SubOptions{
				BufferSize: 1000,
			})
		} else {
			sub, err = rs.ps.Subscribe(topic, SubOptions{
				BufferSize: 1000,
			})
		}

		if err != nil {
			return err
		}

		// Start consumer
		rs.wg.Add(1)
		go func(s Subscription) {
			defer rs.wg.Done()
			for {
				select {
				case msg := <-s.C():
					rs.storeMessage(msg)
				case <-s.Done():
					return
				case <-rs.ctx.Done():
					s.Cancel()
					return
				}
			}
		}(sub)
	}

	return nil
}

// storeMessage stores a message for replay
func (rs *ReplayStore) storeMessage(msg Message) {
	if rs.closed.Load() {
		return
	}

	// Generate ID if missing
	msgID := msg.ID
	if msgID == "" {
		msgID = fmt.Sprintf("%s-%d-%d", msg.Topic, msg.Time.UnixNano(), rs.stats.TotalMessages.Load())
	}

	// Check size limit
	rs.messagesMu.Lock()
	if len(rs.messages) >= rs.config.MaxMessages {
		// Remove oldest message
		rs.removeOldest()
	}

	// Create replay message
	replayMsg := &ReplayMessage{
		ID:        msgID,
		Topic:     msg.Topic,
		Timestamp: msg.Time,
		Metadata:  make(map[string]string),
	}

	if rs.config.IncludeData {
		replayMsg.Data = msg.Data
	}

	// Estimate size
	if data, err := json.Marshal(msg.Data); err == nil {
		replayMsg.Size = len(data)
	}

	// Store message
	rs.messages[msgID] = replayMsg
	rs.messagesMu.Unlock()

	// Update topic index
	rs.topicIndexMu.Lock()
	rs.topicIndex[msg.Topic] = append(rs.topicIndex[msg.Topic], msgID)
	rs.topicIndexMu.Unlock()

	// Update time index
	rs.timeIndexMu.Lock()
	rs.timeIndex = append(rs.timeIndex, replayMsg)
	// Keep sorted by timestamp
	sort.Slice(rs.timeIndex, func(i, j int) bool {
		return rs.timeIndex[i].Timestamp.Before(rs.timeIndex[j].Timestamp)
	})
	rs.timeIndexMu.Unlock()

	// Update stats
	rs.stats.TotalMessages.Add(1)
	rs.stats.TotalSize.Add(uint64(replayMsg.Size))
}

// removeOldest removes the oldest message from storage
func (rs *ReplayStore) removeOldest() {
	if len(rs.timeIndex) == 0 {
		return
	}

	// Get oldest from time index
	oldest := rs.timeIndex[0]

	// Remove from messages map
	delete(rs.messages, oldest.ID)

	// Remove from topic index
	rs.topicIndexMu.Lock()
	if ids, ok := rs.topicIndex[oldest.Topic]; ok {
		for i, id := range ids {
			if id == oldest.ID {
				rs.topicIndex[oldest.Topic] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
	}
	rs.topicIndexMu.Unlock()

	// Remove from time index
	rs.timeIndexMu.Lock()
	rs.timeIndex = rs.timeIndex[1:]
	rs.timeIndexMu.Unlock()
}

// Query searches for messages matching the query
func (rs *ReplayStore) Query(query ReplayQuery) ([]*ReplayMessage, error) {
	if rs.closed.Load() {
		return nil, ErrReplayClosed
	}

	// Validate time range
	if !query.StartTime.IsZero() && !query.EndTime.IsZero() {
		if query.EndTime.Before(query.StartTime) {
			return nil, ErrInvalidTimeRange
		}
	}

	results := make([]*ReplayMessage, 0)

	// Query in-memory messages
	rs.messagesMu.RLock()
	for _, msg := range rs.messages {
		if rs.matchesQuery(msg, query) {
			msgCopy := *msg
			results = append(results, &msgCopy)
		}
	}
	rs.messagesMu.RUnlock()

	// Include archived if requested
	if query.IncludeArchived {
		rs.archivedMu.RLock()
		for _, msg := range rs.archived {
			if rs.matchesQuery(msg, query) {
				msgCopy := *msg
				results = append(results, &msgCopy)
			}
		}
		rs.archivedMu.RUnlock()
	}

	// Sort results
	rs.sortMessages(results, query.SortBy, query.Ascending)

	// Apply offset and limit
	if query.Offset > 0 {
		if query.Offset >= len(results) {
			return []*ReplayMessage{}, nil
		}
		results = results[query.Offset:]
	}

	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results, nil
}

// matchesQuery checks if a message matches the query filters
func (rs *ReplayStore) matchesQuery(msg *ReplayMessage, query ReplayQuery) bool {
	// Time range filter
	if !query.StartTime.IsZero() && msg.Timestamp.Before(query.StartTime) {
		return false
	}
	if !query.EndTime.IsZero() && msg.Timestamp.After(query.EndTime) {
		return false
	}

	// Topic filter
	if len(query.Topics) > 0 {
		found := false
		for _, topic := range query.Topics {
			if msg.Topic == topic {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Message ID filter
	if len(query.MessageIDs) > 0 {
		found := false
		for _, id := range query.MessageIDs {
			if msg.ID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// sortMessages sorts messages by the specified field
func (rs *ReplayStore) sortMessages(messages []*ReplayMessage, sortBy string, ascending bool) {
	switch sortBy {
	case "topic":
		sort.Slice(messages, func(i, j int) bool {
			if ascending {
				return messages[i].Topic < messages[j].Topic
			}
			return messages[i].Topic > messages[j].Topic
		})
	case "size":
		sort.Slice(messages, func(i, j int) bool {
			if ascending {
				return messages[i].Size < messages[j].Size
			}
			return messages[i].Size > messages[j].Size
		})
	default: // timestamp
		sort.Slice(messages, func(i, j int) bool {
			if ascending {
				return messages[i].Timestamp.Before(messages[j].Timestamp)
			}
			return messages[i].Timestamp.After(messages[j].Timestamp)
		})
	}
}

// GetByID retrieves a message by ID
func (rs *ReplayStore) GetByID(messageID string) (*ReplayMessage, error) {
	if rs.closed.Load() {
		return nil, ErrReplayClosed
	}

	rs.messagesMu.RLock()
	msg, exists := rs.messages[messageID]
	rs.messagesMu.RUnlock()

	if !exists {
		return nil, ErrMessageNotFound
	}

	msgCopy := *msg
	return &msgCopy, nil
}

// Replay replays messages matching the query
func (rs *ReplayStore) Replay(query ReplayQuery) (int, error) {
	messages, err := rs.Query(query)
	if err != nil {
		return 0, err
	}

	replayed := 0
	for _, msg := range messages {
		// Reconstruct original message
		original := Message{
			ID:    msg.ID,
			Topic: msg.Topic,
			Data:  msg.Data,
			Time:  msg.Timestamp,
		}

		// Publish to topic
		if err := rs.ps.Publish(msg.Topic, original); err == nil {
			// Update replay metadata
			rs.messagesMu.Lock()
			if stored, ok := rs.messages[msg.ID]; ok {
				stored.ReplayCount++
				stored.LastReplay = time.Now()
			}
			rs.messagesMu.Unlock()

			replayed++
			rs.stats.ReplayedMessages.Add(1)
		}
	}

	return replayed, nil
}

// archiveWorker periodically archives old messages
func (rs *ReplayStore) archiveWorker() {
	defer rs.wg.Done()

	if !rs.config.ArchiveEnabled {
		return
	}

	ticker := time.NewTicker(rs.config.ArchiveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rs.archiveOldMessages()
		case <-rs.ctx.Done():
			return
		}
	}
}

// archiveOldMessages moves old messages to archive
func (rs *ReplayStore) archiveOldMessages() {
	cutoff := time.Now().Add(-rs.config.RetentionPeriod)

	toArchive := make([]*ReplayMessage, 0)

	rs.messagesMu.Lock()
	rs.timeIndexMu.Lock()

	for i, msg := range rs.timeIndex {
		if msg.Timestamp.Before(cutoff) {
			toArchive = append(toArchive, msg)
			delete(rs.messages, msg.ID)

			// Remove from topic index
			rs.topicIndexMu.Lock()
			if ids, ok := rs.topicIndex[msg.Topic]; ok {
				for j, id := range ids {
					if id == msg.ID {
						rs.topicIndex[msg.Topic] = append(ids[:j], ids[j+1:]...)
						break
					}
				}
			}
			rs.topicIndexMu.Unlock()
		} else {
			// Time index is sorted, so we can stop here
			rs.timeIndex = rs.timeIndex[i:]
			break
		}
	}

	rs.timeIndexMu.Unlock()
	rs.messagesMu.Unlock()

	if len(toArchive) == 0 {
		return
	}

	// Write to archive file
	if err := rs.writeArchive(toArchive); err != nil {
		// Log error but continue
		return
	}

	// Move to archived list
	rs.archivedMu.Lock()
	for _, msg := range toArchive {
		msg.IsArchived = true
		rs.archived = append(rs.archived, msg)
		rs.stats.ArchivedMessages.Add(1)
	}
	rs.archivedMu.Unlock()
}

// writeArchive writes messages to archive file
func (rs *ReplayStore) writeArchive(messages []*ReplayMessage) error {
	if len(messages) == 0 {
		return nil
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("archive-%s.json", time.Now().Format("2006-01-02-150405"))
	if rs.config.ArchiveCompression {
		filename += ".gz"
	}

	filepath := filepath.Join(rs.config.ArchiveDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var writer io.Writer = file

	// Add gzip compression if enabled
	if rs.config.ArchiveCompression {
		gzWriter := gzip.NewWriter(file)
		defer gzWriter.Close()
		writer = gzWriter
	}

	// Write messages as JSON lines
	encoder := json.NewEncoder(writer)
	for _, msg := range messages {
		if err := encoder.Encode(msg); err != nil {
			return err
		}
	}

	return nil
}

// LoadArchive loads messages from an archive file
func (rs *ReplayStore) LoadArchive(filename string) error {
	if !rs.config.ArchiveEnabled {
		return ErrArchiveNotEnabled
	}

	filepath := filepath.Join(rs.config.ArchiveDir, filename)

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var reader io.Reader = file

	// Detect gzip compression
	if filepath[len(filepath)-3:] == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Read messages
	decoder := json.NewDecoder(reader)
	loaded := 0

	rs.archivedMu.Lock()
	defer rs.archivedMu.Unlock()

	for {
		var msg ReplayMessage
		if err := decoder.Decode(&msg); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		msg.IsArchived = true
		rs.archived = append(rs.archived, &msg)
		loaded++
	}

	return nil
}

// Stats returns replay store statistics.
func (rs *ReplayStore) Stats() *ReplayStats {
	stats := &ReplayStats{}
	stats.TotalMessages.Store(rs.stats.TotalMessages.Load())
	stats.ArchivedMessages.Store(rs.stats.ArchivedMessages.Load())
	stats.ReplayedMessages.Store(rs.stats.ReplayedMessages.Load())
	stats.TotalSize.Store(rs.stats.TotalSize.Load())

	// Calculate oldest and newest
	rs.timeIndexMu.RLock()
	if len(rs.timeIndex) > 0 {
		stats.OldestMessage = rs.timeIndex[0].Timestamp
		stats.NewestMessage = rs.timeIndex[len(rs.timeIndex)-1].Timestamp
	}
	rs.timeIndexMu.RUnlock()

	// Count topics
	rs.topicIndexMu.RLock()
	stats.TopicCount = len(rs.topicIndex)
	rs.topicIndexMu.RUnlock()

	return stats
}

// Close stops the replay store
func (rs *ReplayStore) Close() error {
	if !rs.closed.CompareAndSwap(false, true) {
		return ErrReplayClosed
	}

	rs.cancel()
	rs.wg.Wait()

	return nil
}
