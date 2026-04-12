package pubsub

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Enhanced DLQ errors

// RetryStrategy defines retry behavior
type RetryStrategy int

const (
	// NoRetry - don't retry failed messages
	NoRetry RetryStrategy = iota

	// FixedDelay - retry with fixed delay
	FixedDelay

	// ExponentialBackoff - retry with exponential backoff
	ExponentialBackoff

	// LinearBackoff - retry with linear backoff
	LinearBackoff
)

// DLQConfig configures the enhanced dead letter queue
type DLQConfig struct {
	// Topic for dead letter messages
	Topic string

	// MaxSize limits the queue size (0 = unlimited)
	MaxSize int

	// TTL for messages in DLQ (0 = forever)
	TTL time.Duration

	// RetryStrategy for automatic reprocessing
	RetryStrategy RetryStrategy

	// InitialRetryDelay for first retry attempt
	InitialRetryDelay time.Duration

	// MaxRetryDelay caps the retry delay
	MaxRetryDelay time.Duration

	// MaxRetryAttempts before giving up (0 = infinite)
	MaxRetryAttempts int

	// EnableAlerts enables alert notifications
	EnableAlerts bool

	// AlertThreshold triggers alert after N messages
	AlertThreshold int

	// AlertCallback called when threshold reached
	AlertCallback func(alert DLQAlert)

	// EnableMetrics enables detailed metrics collection
	EnableMetrics bool

	// ArchiveAfter moves old messages to archive
	ArchiveAfter time.Duration
}

// DefaultDLQConfig returns default configuration
func DefaultDLQConfig(topic string) DLQConfig {
	return DLQConfig{
		Topic:             topic,
		MaxSize:           10000,
		TTL:               7 * 24 * time.Hour,
		RetryStrategy:     ExponentialBackoff,
		InitialRetryDelay: 1 * time.Second,
		MaxRetryDelay:     1 * time.Hour,
		MaxRetryAttempts:  5,
		EnableAlerts:      true,
		AlertThreshold:    100,
		EnableMetrics:     true,
		ArchiveAfter:      24 * time.Hour,
	}
}

// DLQ provides advanced dead letter queue functionality
type DLQ struct {
	config DLQConfig
	ps     *InProcBroker

	// Message storage
	messages   map[string]*DLQMessage
	messagesMu sync.RWMutex
	sequence   atomic.Uint64

	// Retry management
	retryQueue   chan *DLQMessage
	retryTimers  map[string]*time.Timer
	retryTimerMu sync.Mutex

	// Archive
	archived   []*DLQMessage
	archivedMu sync.RWMutex

	// Metrics
	stats DLQStats

	// Background workers
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed atomic.Bool
}

// DLQMessage represents a message in the DLQ
type DLQMessage struct {
	ID            string
	OriginalMsg   Message
	OriginalTopic string
	Reason        string
	Timestamp     time.Time
	Sequence      uint64
	RetryCount    int
	LastRetry     time.Time
	NextRetry     time.Time
	IsArchived    bool
	ArchiveTime   time.Time
	Tags          map[string]string
	Priority      int
}

// DLQAlert represents an alert notification
type DLQAlert struct {
	Level     string
	Message   string
	Count     int
	Threshold int
	Reason    string
	Timestamp time.Time
	Affected  []string
}

// DLQStats holds DLQ statistics
type DLQStats struct {
	TotalMessages     atomic.Uint64
	ArchivedMessages  atomic.Uint64
	RetriedMessages   atomic.Uint64
	SuccessfulRetries atomic.Uint64
	FailedRetries     atomic.Uint64
	AlertsTriggered   atomic.Uint64
}

// DLQQueryOptions defines query parameters
type DLQQueryOptions struct {
	// TimeRange filters by timestamp
	StartTime time.Time
	EndTime   time.Time

	// Reason filters by failure reason
	Reason string

	// OriginalTopic filters by original topic
	OriginalTopic string

	// MinRetryCount filters by retry attempts
	MinRetryCount int

	// IncludeArchived includes archived messages
	IncludeArchived bool

	// Limit limits results
	Limit int

	// SortBy sorts results
	SortBy string // "timestamp", "retries", "sequence"

	// Tags filters by tags
	Tags map[string]string
}

// NewDLQ creates a new enhanced dead letter queue
func NewDLQ(ps *InProcBroker, config DLQConfig) *DLQ {
	ctx, cancel := context.WithCancel(context.Background())

	dlq := &DLQ{
		config:      config,
		ps:          ps,
		messages:    make(map[string]*DLQMessage),
		retryQueue:  make(chan *DLQMessage, 100),
		retryTimers: make(map[string]*time.Timer),
		archived:    make([]*DLQMessage, 0),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start background workers
	dlq.startWorkers()

	return dlq
}

// Add adds a message to the DLQ
func (dlq *DLQ) Add(originalMsg Message, originalTopic, reason string) error {
	if dlq.closed.Load() {
		return ErrDLQClosed
	}

	dlq.messagesMu.Lock()
	defer dlq.messagesMu.Unlock()

	// Generate unique ID
	msgID := fmt.Sprintf("dlq-%d-%s", time.Now().UnixNano(), generateCorrelationID())

	msg := &DLQMessage{
		ID:            msgID,
		OriginalMsg:   originalMsg,
		OriginalTopic: originalTopic,
		Reason:        reason,
		Timestamp:     time.Now(),
		Sequence:      dlq.sequence.Add(1),
		RetryCount:    0,
		Tags:          make(map[string]string),
	}

	dlq.messages[msgID] = msg
	dlq.stats.TotalMessages.Add(1)

	// Publish to DLQ topic
	if dlq.config.Topic != "" {
		dlqMsg := Message{
			ID:   msgID,
			Data: originalMsg.Data,
			Meta: map[string]string{
				"X-DLQ-Reason":         reason,
				"X-DLQ-Original-Topic": originalTopic,
				"X-DLQ-Timestamp":      msg.Timestamp.Format(time.RFC3339),
			},
		}
		_ = dlq.ps.Publish(dlq.config.Topic, dlqMsg)
	}

	// Check size limit
	if dlq.config.MaxSize > 0 && len(dlq.messages) > dlq.config.MaxSize {
		dlq.evictOldest()
	}

	// Schedule retry if configured
	if dlq.config.RetryStrategy != NoRetry {
		dlq.scheduleRetry(msg)
	}

	// Check alert threshold
	if dlq.config.EnableAlerts {
		dlq.checkAlertThreshold()
	}

	return nil
}

// Get retrieves a message by ID
func (dlq *DLQ) Get(msgID string) (*DLQMessage, error) {
	dlq.messagesMu.RLock()
	defer dlq.messagesMu.RUnlock()

	msg, exists := dlq.messages[msgID]
	if !exists {
		return nil, ErrDLQNotFound
	}

	// Create copy
	msgCopy := *msg
	return &msgCopy, nil
}

// Query searches messages with filters
func (dlq *DLQ) Query(opts DLQQueryOptions) ([]*DLQMessage, error) {
	dlq.messagesMu.RLock()
	defer dlq.messagesMu.RUnlock()

	results := make([]*DLQMessage, 0)

	// Filter messages
	for _, msg := range dlq.messages {
		if !dlq.matchesFilter(msg, opts) {
			continue
		}
		msgCopy := *msg
		results = append(results, &msgCopy)
	}

	// Include archived if requested
	if opts.IncludeArchived {
		dlq.archivedMu.RLock()
		for _, msg := range dlq.archived {
			if dlq.matchesFilter(msg, opts) {
				msgCopy := *msg
				results = append(results, &msgCopy)
			}
		}
		dlq.archivedMu.RUnlock()
	}

	// Sort results
	dlq.sortMessages(results, opts.SortBy)

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// matchesFilter checks if message matches query filters
func (dlq *DLQ) matchesFilter(msg *DLQMessage, opts DLQQueryOptions) bool {
	// Time range
	if !opts.StartTime.IsZero() && msg.Timestamp.Before(opts.StartTime) {
		return false
	}
	if !opts.EndTime.IsZero() && msg.Timestamp.After(opts.EndTime) {
		return false
	}

	// Reason
	if opts.Reason != "" && msg.Reason != opts.Reason {
		return false
	}

	// Original topic
	if opts.OriginalTopic != "" && msg.OriginalTopic != opts.OriginalTopic {
		return false
	}

	// Retry count
	if msg.RetryCount < opts.MinRetryCount {
		return false
	}

	// Tags
	for key, value := range opts.Tags {
		if msg.Tags[key] != value {
			return false
		}
	}

	return true
}

// sortMessages sorts messages by specified field
func (dlq *DLQ) sortMessages(messages []*DLQMessage, sortBy string) {
	switch sortBy {
	case "timestamp":
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].Timestamp.Before(messages[j].Timestamp)
		})
	case "retries":
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].RetryCount > messages[j].RetryCount
		})
	case "sequence":
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].Sequence < messages[j].Sequence
		})
	default:
		// Default to timestamp
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].Timestamp.Before(messages[j].Timestamp)
		})
	}
}

// Retry manually retries a specific message
func (dlq *DLQ) Retry(msgID string) error {
	dlq.messagesMu.Lock()
	msg, exists := dlq.messages[msgID]
	if !exists {
		dlq.messagesMu.Unlock()
		return ErrDLQNotFound
	}
	dlq.messagesMu.Unlock()

	return dlq.retryMessage(msg)
}

// RetryBatch retries multiple messages
func (dlq *DLQ) RetryBatch(msgIDs []string) (succeeded, failed int) {
	for _, msgID := range msgIDs {
		if err := dlq.Retry(msgID); err == nil {
			succeeded++
		} else {
			failed++
		}
	}
	return
}

// RetryAll retries all messages matching filter
func (dlq *DLQ) RetryAll(opts DLQQueryOptions) (succeeded, failed int) {
	messages, err := dlq.Query(opts)
	if err != nil {
		return 0, 0
	}

	for _, msg := range messages {
		if err := dlq.retryMessage(msg); err == nil {
			succeeded++
		} else {
			failed++
		}
	}

	return
}

// retryMessage attempts to reprocess a message
func (dlq *DLQ) retryMessage(msg *DLQMessage) error {
	dlq.stats.RetriedMessages.Add(1)

	// Publish back to original topic
	err := dlq.ps.Publish(msg.OriginalTopic, msg.OriginalMsg)

	dlq.messagesMu.Lock()
	defer dlq.messagesMu.Unlock()

	msg.RetryCount++
	msg.LastRetry = time.Now()

	if err == nil {
		// Success - remove from DLQ
		delete(dlq.messages, msg.ID)
		dlq.stats.SuccessfulRetries.Add(1)
		return nil
	}

	// Failed - schedule next retry or give up
	dlq.stats.FailedRetries.Add(1)

	if dlq.config.MaxRetryAttempts > 0 && msg.RetryCount >= dlq.config.MaxRetryAttempts {
		// Max retries reached - archive
		dlq.archiveMessage(msg)
		return fmt.Errorf("max retries reached: %w", err)
	}

	// Schedule next retry
	dlq.scheduleRetry(msg)

	return err
}

// scheduleRetry schedules a message for retry
func (dlq *DLQ) scheduleRetry(msg *DLQMessage) {
	delay := dlq.calculateRetryDelay(msg.RetryCount)
	msg.NextRetry = time.Now().Add(delay)

	dlq.retryTimerMu.Lock()
	defer dlq.retryTimerMu.Unlock()

	// Cancel existing timer if any
	if timer, exists := dlq.retryTimers[msg.ID]; exists {
		timer.Stop()
	}

	// Create new timer
	timer := time.AfterFunc(delay, func() {
		select {
		case dlq.retryQueue <- msg:
		case <-dlq.ctx.Done():
		}
	})

	dlq.retryTimers[msg.ID] = timer
}

// calculateRetryDelay calculates retry delay based on strategy
func (dlq *DLQ) calculateRetryDelay(retryCount int) time.Duration {
	switch dlq.config.RetryStrategy {
	case FixedDelay:
		return dlq.config.InitialRetryDelay

	case ExponentialBackoff:
		// 2^n * initial delay
		delay := dlq.config.InitialRetryDelay * time.Duration(1<<uint(retryCount))
		if delay > dlq.config.MaxRetryDelay {
			delay = dlq.config.MaxRetryDelay
		}
		return delay

	case LinearBackoff:
		// n * initial delay
		delay := dlq.config.InitialRetryDelay * time.Duration(retryCount+1)
		if delay > dlq.config.MaxRetryDelay {
			delay = dlq.config.MaxRetryDelay
		}
		return delay

	default:
		return dlq.config.InitialRetryDelay
	}
}

// archiveMessage moves a message to the archive.
// MUST be called with dlq.messagesMu already held (write lock).
// It releases and re-acquires the lock around the archivedMu write to avoid
// nested lock ordering: messagesMu > archivedMu would deadlock against any
// reader that acquires archivedMu while messagesMu is read-locked.
func (dlq *DLQ) archiveMessage(msg *DLQMessage) {
	msg.IsArchived = true
	msg.ArchiveTime = time.Now()

	delete(dlq.messages, msg.ID)
	// Release messagesMu before acquiring archivedMu to preserve a consistent
	// lock ordering and avoid potential deadlocks.
	dlq.messagesMu.Unlock()

	dlq.archivedMu.Lock()
	dlq.archived = append(dlq.archived, msg)
	dlq.archivedMu.Unlock()

	dlq.stats.ArchivedMessages.Add(1)

	// Re-acquire messagesMu so the caller's defer/unlock is still valid.
	dlq.messagesMu.Lock()
}

// evictOldest removes oldest messages to stay within size limit
func (dlq *DLQ) evictOldest() {
	// Find oldest message
	var oldest *DLQMessage
	for _, msg := range dlq.messages {
		if oldest == nil || msg.Timestamp.Before(oldest.Timestamp) {
			oldest = msg
		}
	}

	if oldest != nil {
		dlq.archiveMessage(oldest)
	}
}

// checkAlertThreshold checks if alert should be triggered
func (dlq *DLQ) checkAlertThreshold() {
	count := len(dlq.messages)

	if count >= dlq.config.AlertThreshold && dlq.config.AlertCallback != nil {
		alert := DLQAlert{
			Level:     "warning",
			Message:   fmt.Sprintf("DLQ threshold reached: %d messages", count),
			Count:     count,
			Threshold: dlq.config.AlertThreshold,
			Timestamp: time.Now(),
		}

		dlq.stats.AlertsTriggered.Add(1)

		// Call callback in goroutine
		go dlq.config.AlertCallback(alert)
	}
}

// Delete removes a message from DLQ
func (dlq *DLQ) Delete(msgID string) error {
	dlq.messagesMu.Lock()
	defer dlq.messagesMu.Unlock()

	if _, exists := dlq.messages[msgID]; !exists {
		return ErrDLQNotFound
	}

	delete(dlq.messages, msgID)

	// Cancel retry timer
	dlq.retryTimerMu.Lock()
	if timer, exists := dlq.retryTimers[msgID]; exists {
		timer.Stop()
		delete(dlq.retryTimers, msgID)
	}
	dlq.retryTimerMu.Unlock()

	return nil
}

// Clear removes all messages
func (dlq *DLQ) Clear() {
	dlq.messagesMu.Lock()
	defer dlq.messagesMu.Unlock()

	dlq.messages = make(map[string]*DLQMessage)

	dlq.retryTimerMu.Lock()
	for _, timer := range dlq.retryTimers {
		timer.Stop()
	}
	dlq.retryTimers = make(map[string]*time.Timer)
	dlq.retryTimerMu.Unlock()
}

// Stats returns DLQ statistics.
func (dlq *DLQ) Stats() *DLQStats {
	stats := &DLQStats{}
	stats.TotalMessages.Store(dlq.stats.TotalMessages.Load())
	stats.ArchivedMessages.Store(dlq.stats.ArchivedMessages.Load())
	stats.RetriedMessages.Store(dlq.stats.RetriedMessages.Load())
	stats.SuccessfulRetries.Store(dlq.stats.SuccessfulRetries.Load())
	stats.FailedRetries.Store(dlq.stats.FailedRetries.Load())
	stats.AlertsTriggered.Store(dlq.stats.AlertsTriggered.Load())
	return stats
}

// Count returns current message count
func (dlq *DLQ) Count() int {
	dlq.messagesMu.RLock()
	defer dlq.messagesMu.RUnlock()
	return len(dlq.messages)
}

// startWorkers starts background workers
func (dlq *DLQ) startWorkers() {
	// Retry processor
	dlq.wg.Add(1)
	go func() {
		defer dlq.wg.Done()

		for {
			select {
			case msg := <-dlq.retryQueue:
				_ = dlq.retryMessage(msg)

			case <-dlq.ctx.Done():
				return
			}
		}
	}()

	// TTL and archive worker
	if dlq.config.TTL > 0 || dlq.config.ArchiveAfter > 0 {
		dlq.wg.Add(1)
		go func() {
			defer dlq.wg.Done()
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					dlq.cleanupExpired()

				case <-dlq.ctx.Done():
					return
				}
			}
		}()
	}
}

// cleanupExpired removes expired and old messages
func (dlq *DLQ) cleanupExpired() {
	now := time.Now()

	dlq.messagesMu.Lock()

	// Collect expired/archivable IDs first to avoid modifying the map while
	// ranging over it (which is undefined behaviour in Go).
	var toDelete []string
	var toArchive []*DLQMessage

	for id, msg := range dlq.messages {
		if dlq.config.TTL > 0 && now.Sub(msg.Timestamp) > dlq.config.TTL {
			toDelete = append(toDelete, id)
			continue
		}
		if dlq.config.ArchiveAfter > 0 && now.Sub(msg.Timestamp) > dlq.config.ArchiveAfter {
			toArchive = append(toArchive, msg)
		}
	}

	for _, id := range toDelete {
		delete(dlq.messages, id)
	}

	// archiveMessage releases and re-acquires messagesMu internally.
	for _, msg := range toArchive {
		dlq.archiveMessage(msg)
	}

	dlq.messagesMu.Unlock()
}

// Close closes the DLQ
func (dlq *DLQ) Close() error {
	if dlq.closed.Swap(true) {
		return nil
	}

	dlq.cancel()
	dlq.wg.Wait()

	// Stop all retry timers
	dlq.retryTimerMu.Lock()
	for _, timer := range dlq.retryTimers {
		timer.Stop()
	}
	dlq.retryTimerMu.Unlock()

	close(dlq.retryQueue)

	return nil
}
