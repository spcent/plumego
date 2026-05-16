package pubsub

import (
	"fmt"
	"time"
)

// retryMessage attempts to reprocess a message.
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

// scheduleRetry schedules a message for retry.
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

// calculateRetryDelay calculates retry delay based on strategy.
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

// evictOldest removes oldest messages to stay within size limit.
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

// checkAlertThreshold checks if alert should be triggered.
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
