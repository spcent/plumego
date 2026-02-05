package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Enhanced DLQ errors
var (
	ErrDLQClosed    = errors.New("dead letter queue is closed")
	ErrDLQNotFound  = errors.New("dead letter message not found")
	ErrInvalidQuery = errors.New("invalid query parameters")
)

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

// EnhancedDLQConfig configures the enhanced dead letter queue
type EnhancedDLQConfig struct {
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

// DefaultEnhancedDLQConfig returns default configuration
func DefaultEnhancedDLQConfig(topic string) EnhancedDLQConfig {
	return EnhancedDLQConfig{
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

// EnhancedDLQ provides advanced dead letter queue functionality
type EnhancedDLQ struct {
	config EnhancedDLQConfig
	ps     *InProcPubSub

	// Message storage
	messages   map[string]*EnhancedDLQMessage
	messagesMu sync.RWMutex
	sequence   atomic.Uint64

	// Retry management
	retryQueue   chan *EnhancedDLQMessage
	retryTimers  map[string]*time.Timer
	retryTimerMu sync.Mutex

	// Archive
	archived   []*EnhancedDLQMessage
	archivedMu sync.RWMutex

	// Metrics
	stats DLQStats

	// Background workers
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed atomic.Bool
}

// EnhancedDLQMessage represents a message in the enhanced DLQ
type EnhancedDLQMessage struct {
	ID             string
	OriginalMsg    Message
	OriginalTopic  string
	Reason         string
	Timestamp      time.Time
	Sequence       uint64
	RetryCount     int
	LastRetry      time.Time
	NextRetry      time.Time
	IsArchived     bool
	ArchiveTime    time.Time
	Tags           map[string]string
	Priority       int
}

// DLQAlert represents an alert notification
type DLQAlert struct {
	Level      string
	Message    string
	Count      int
	Threshold  int
	Reason     string
	Timestamp  time.Time
	Affected   []string
}

// DLQStats holds DLQ statistics
type DLQStats struct {
	TotalMessages    atomic.Uint64
	ArchivedMessages atomic.Uint64
	RetriedMessages  atomic.Uint64
	SuccessfulRetries atomic.Uint64
	FailedRetries    atomic.Uint64
	AlertsTriggered  atomic.Uint64
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

// NewEnhancedDLQ creates a new enhanced dead letter queue
func NewEnhancedDLQ(ps *InProcPubSub, config EnhancedDLQConfig) *EnhancedDLQ {
	ctx, cancel := context.WithCancel(context.Background())

	dlq := &EnhancedDLQ{
		config:      config,
		ps:          ps,
		messages:    make(map[string]*EnhancedDLQMessage),
		retryQueue:  make(chan *EnhancedDLQMessage, 100),
		retryTimers: make(map[string]*time.Timer),
		archived:    make([]*EnhancedDLQMessage, 0),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start background workers
	dlq.startWorkers()

	return dlq
}

// Add adds a message to the DLQ
func (dlq *EnhancedDLQ) Add(originalMsg Message, originalTopic, reason string) error {
	if dlq.closed.Load() {
		return ErrDLQClosed
	}

	dlq.messagesMu.Lock()
	defer dlq.messagesMu.Unlock()

	// Generate unique ID
	msgID := fmt.Sprintf("dlq-%d-%s", time.Now().UnixNano(), generateCorrelationID())

	msg := &EnhancedDLQMessage{
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
func (dlq *EnhancedDLQ) Get(msgID string) (*EnhancedDLQMessage, error) {
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
func (dlq *EnhancedDLQ) Query(opts DLQQueryOptions) ([]*EnhancedDLQMessage, error) {
	dlq.messagesMu.RLock()
	defer dlq.messagesMu.RUnlock()

	results := make([]*EnhancedDLQMessage, 0)

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
func (dlq *EnhancedDLQ) matchesFilter(msg *EnhancedDLQMessage, opts DLQQueryOptions) bool {
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
func (dlq *EnhancedDLQ) sortMessages(messages []*EnhancedDLQMessage, sortBy string) {
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
func (dlq *EnhancedDLQ) Retry(msgID string) error {
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
func (dlq *EnhancedDLQ) RetryBatch(msgIDs []string) (succeeded, failed int) {
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
func (dlq *EnhancedDLQ) RetryAll(opts DLQQueryOptions) (succeeded, failed int) {
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
func (dlq *EnhancedDLQ) retryMessage(msg *EnhancedDLQMessage) error {
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
func (dlq *EnhancedDLQ) scheduleRetry(msg *EnhancedDLQMessage) {
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
func (dlq *EnhancedDLQ) calculateRetryDelay(retryCount int) time.Duration {
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

// archiveMessage moves a message to archive
func (dlq *EnhancedDLQ) archiveMessage(msg *EnhancedDLQMessage) {
	msg.IsArchived = true
	msg.ArchiveTime = time.Now()

	delete(dlq.messages, msg.ID)

	dlq.archivedMu.Lock()
	dlq.archived = append(dlq.archived, msg)
	dlq.archivedMu.Unlock()

	dlq.stats.ArchivedMessages.Add(1)
}

// evictOldest removes oldest messages to stay within size limit
func (dlq *EnhancedDLQ) evictOldest() {
	// Find oldest message
	var oldest *EnhancedDLQMessage
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
func (dlq *EnhancedDLQ) checkAlertThreshold() {
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
func (dlq *EnhancedDLQ) Delete(msgID string) error {
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
func (dlq *EnhancedDLQ) Clear() {
	dlq.messagesMu.Lock()
	defer dlq.messagesMu.Unlock()

	dlq.messages = make(map[string]*EnhancedDLQMessage)

	dlq.retryTimerMu.Lock()
	for _, timer := range dlq.retryTimers {
		timer.Stop()
	}
	dlq.retryTimers = make(map[string]*time.Timer)
	dlq.retryTimerMu.Unlock()
}

// Stats returns DLQ statistics
func (dlq *EnhancedDLQ) Stats() DLQStats {
	return dlq.stats
}

// Count returns current message count
func (dlq *EnhancedDLQ) Count() int {
	dlq.messagesMu.RLock()
	defer dlq.messagesMu.RUnlock()
	return len(dlq.messages)
}

// startWorkers starts background workers
func (dlq *EnhancedDLQ) startWorkers() {
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
func (dlq *EnhancedDLQ) cleanupExpired() {
	now := time.Now()

	dlq.messagesMu.Lock()
	defer dlq.messagesMu.Unlock()

	for id, msg := range dlq.messages {
		// TTL check
		if dlq.config.TTL > 0 && now.Sub(msg.Timestamp) > dlq.config.TTL {
			delete(dlq.messages, id)
			continue
		}

		// Archive old messages
		if dlq.config.ArchiveAfter > 0 && now.Sub(msg.Timestamp) > dlq.config.ArchiveAfter {
			dlq.archiveMessage(msg)
		}
	}
}

// Close closes the DLQ
func (dlq *EnhancedDLQ) Close() error {
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
