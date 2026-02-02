package mq

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// deadLetterEntry represents a message in the dead letter queue.
type deadLetterEntry struct {
	originalTopic string
	message       Message
	reason        string
	timestamp     time.Time
}

// deadLetterManager manages the dead letter queue.
type deadLetterManager struct {
	mu              sync.RWMutex
	entries         []deadLetterEntry
	totalCount      uint64
	lastMessageTime time.Time
	broker          *InProcBroker
	closed          bool
}

// newDeadLetterManager creates a new dead letter manager.
func newDeadLetterManager(broker *InProcBroker) *deadLetterManager {
	return &deadLetterManager{
		entries: make([]deadLetterEntry, 0),
		broker:  broker,
	}
}

// add adds a message to the dead letter queue.
func (dlm *deadLetterManager) add(originalTopic string, msg Message, reason string) error {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()

	if dlm.closed {
		return ErrBrokerClosed
	}

	entry := deadLetterEntry{
		originalTopic: originalTopic,
		message:       msg,
		reason:        reason,
		timestamp:     time.Now(),
	}

	dlm.entries = append(dlm.entries, entry)
	dlm.totalCount++
	dlm.lastMessageTime = time.Now()

	return nil
}

// stats returns statistics about the dead letter queue.
func (dlm *deadLetterManager) stats() DeadLetterStats {
	dlm.mu.RLock()
	defer dlm.mu.RUnlock()

	return DeadLetterStats{
		Enabled:         true,
		Topic:           dlm.broker.config.DeadLetterTopic,
		TotalMessages:   dlm.totalCount,
		CurrentCount:    len(dlm.entries),
		LastMessageTime: dlm.lastMessageTime,
	}
}

// close shuts down the dead letter manager.
func (dlm *deadLetterManager) close() {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()

	if dlm.closed {
		return
	}

	dlm.closed = true
	dlm.entries = nil
}

// PublishToDeadLetter publishes a message to the dead letter queue.
func (b *InProcBroker) PublishToDeadLetter(ctx context.Context, originalTopic string, msg Message, reason string) error {
	return b.executeWithObservability(ctx, OpPublish, originalTopic, func() error {
		// Validate context
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}

		// Validate broker initialization
		if b == nil || b.ps == nil {
			return ErrNotInitialized
		}

		// Check if dead letter queue is enabled
		if !b.config.EnableDeadLetterQueue {
			return fmt.Errorf("%w: dead letter queue is disabled", ErrDeadLetterNotSupported)
		}

		// Validate deadLetterManager
		if b.deadLetterManager == nil {
			return fmt.Errorf("%w: dead letter manager not initialized", ErrNotInitialized)
		}

		// Validate message
		if err := validateMessage(msg); err != nil {
			return err
		}

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
			return err
		}

		// Add to dead letter manager
		if err := b.deadLetterManager.add(originalTopic, msg, reason); err != nil {
			return err
		}

		// Determine dead letter topic
		topic := b.config.DeadLetterTopic
		if topic == "" {
			topic = "dead-letter"
		}

		// Publish to dead letter topic
		return b.ps.Publish(topic, msg)
	})
}

// GetDeadLetterStats returns statistics about the dead letter queue.
func (b *InProcBroker) GetDeadLetterStats() DeadLetterStats {
	if b == nil || !b.config.EnableDeadLetterQueue {
		return DeadLetterStats{
			Enabled: false,
		}
	}

	if b.deadLetterManager == nil {
		return DeadLetterStats{
			Enabled: true,
			Topic:   b.config.DeadLetterTopic,
		}
	}

	return b.deadLetterManager.stats()
}
