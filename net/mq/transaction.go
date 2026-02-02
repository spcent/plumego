package mq

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TransactionState represents the state of a transaction.
type TransactionState int

const (
	// TxStateActive - Transaction is active and accepting messages
	TxStateActive TransactionState = iota
	// TxStateCommitted - Transaction has been committed
	TxStateCommitted
	// TxStateRolledBack - Transaction has been rolled back
	TxStateRolledBack
	// TxStateTimeout - Transaction has timed out
	TxStateTimeout
)

// transaction represents a message queue transaction.
type transaction struct {
	id        string
	state     TransactionState
	messages  []messageEnvelope // Buffered messages
	createdAt time.Time
	timeout   time.Duration
	timer     *time.Timer
	mu        sync.RWMutex
}

// messageEnvelope wraps a message with its topic for transaction buffering.
type messageEnvelope struct {
	topic string
	msg   Message
}

// transactionManager manages all active transactions.
type transactionManager struct {
	mu           sync.RWMutex
	transactions map[string]*transaction
	broker       *InProcBroker
	closed       bool
}

// newTransactionManager creates a new transaction manager.
func newTransactionManager(broker *InProcBroker) *transactionManager {
	return &transactionManager{
		transactions: make(map[string]*transaction),
		broker:       broker,
	}
}

// begin starts a new transaction with the given ID and timeout.
func (tm *transactionManager) begin(txID string, timeout time.Duration) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.closed {
		return ErrBrokerClosed
	}

	if _, exists := tm.transactions[txID]; exists {
		return fmt.Errorf("transaction %s already exists", txID)
	}

	tx := &transaction{
		id:        txID,
		state:     TxStateActive,
		messages:  make([]messageEnvelope, 0),
		createdAt: time.Now(),
		timeout:   timeout,
	}

	// Set up timeout timer
	if timeout > 0 {
		tx.timer = time.AfterFunc(timeout, func() {
			tm.handleTimeout(txID)
		})
	}

	tm.transactions[txID] = tx
	return nil
}

// addMessage adds a message to the transaction buffer.
func (tm *transactionManager) addMessage(txID string, topic string, msg Message) error {
	tm.mu.RLock()
	tx, exists := tm.transactions[txID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: %s", ErrTransactionNotFound, txID)
	}

	tx.mu.Lock()
	defer tx.mu.Unlock()

	switch tx.state {
	case TxStateCommitted:
		return ErrTransactionCommitted
	case TxStateRolledBack:
		return ErrTransactionRolledBack
	case TxStateTimeout:
		return ErrTransactionTimeout
	}

	tx.messages = append(tx.messages, messageEnvelope{
		topic: topic,
		msg:   msg,
	})

	return nil
}

// commit commits the transaction, publishing all buffered messages.
func (tm *transactionManager) commit(txID string) error {
	tm.mu.Lock()
	tx, exists := tm.transactions[txID]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrTransactionNotFound, txID)
	}
	tm.mu.Unlock()

	tx.mu.Lock()
	defer tx.mu.Unlock()

	switch tx.state {
	case TxStateCommitted:
		return ErrTransactionCommitted
	case TxStateRolledBack:
		return ErrTransactionRolledBack
	case TxStateTimeout:
		return ErrTransactionTimeout
	}

	// Stop the timeout timer
	if tx.timer != nil {
		tx.timer.Stop()
	}

	// Publish all buffered messages
	for _, env := range tx.messages {
		if err := tm.broker.ps.Publish(env.topic, env.msg); err != nil {
			// If publish fails, mark as rolled back
			tx.state = TxStateRolledBack
			return fmt.Errorf("transaction commit failed: %w", err)
		}
	}

	// Mark as committed
	tx.state = TxStateCommitted

	// Remove from active transactions
	tm.mu.Lock()
	delete(tm.transactions, txID)
	tm.mu.Unlock()

	return nil
}

// rollback rolls back the transaction, discarding all buffered messages.
func (tm *transactionManager) rollback(txID string) error {
	tm.mu.Lock()
	tx, exists := tm.transactions[txID]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrTransactionNotFound, txID)
	}
	tm.mu.Unlock()

	tx.mu.Lock()
	defer tx.mu.Unlock()

	switch tx.state {
	case TxStateCommitted:
		return ErrTransactionCommitted
	case TxStateRolledBack:
		return ErrTransactionRolledBack
	case TxStateTimeout:
		return ErrTransactionTimeout
	}

	// Stop the timeout timer
	if tx.timer != nil {
		tx.timer.Stop()
	}

	// Mark as rolled back and discard messages
	tx.state = TxStateRolledBack
	tx.messages = nil

	// Remove from active transactions
	tm.mu.Lock()
	delete(tm.transactions, txID)
	tm.mu.Unlock()

	return nil
}

// handleTimeout processes a transaction timeout.
func (tm *transactionManager) handleTimeout(txID string) {
	tm.mu.Lock()
	tx, exists := tm.transactions[txID]
	if !exists {
		tm.mu.Unlock()
		return
	}
	tm.mu.Unlock()

	tx.mu.Lock()
	if tx.state != TxStateActive {
		tx.mu.Unlock()
		return
	}

	// Mark as timed out and discard messages
	tx.state = TxStateTimeout
	tx.messages = nil
	tx.mu.Unlock()

	// Remove from active transactions
	tm.mu.Lock()
	delete(tm.transactions, txID)
	tm.mu.Unlock()
}

// close shuts down the transaction manager and aborts all active transactions.
func (tm *transactionManager) close() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.closed {
		return
	}

	tm.closed = true

	// Abort all active transactions
	for _, tx := range tm.transactions {
		tx.mu.Lock()
		if tx.timer != nil {
			tx.timer.Stop()
		}
		tx.state = TxStateRolledBack
		tx.messages = nil
		tx.mu.Unlock()
	}

	// Clear all transactions
	tm.transactions = nil
}

// stats returns statistics about active transactions.
func (tm *transactionManager) stats() (active int, totalMessages int) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	active = len(tm.transactions)
	for _, tx := range tm.transactions {
		tx.mu.RLock()
		totalMessages += len(tx.messages)
		tx.mu.RUnlock()
	}

	return
}

// PublishWithTransaction publishes a message within a transaction.
func (b *InProcBroker) PublishWithTransaction(ctx context.Context, topic string, msg Message, txID string) error {
	return b.executeWithObservability(ctx, OpPublish, topic, func() error {
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

		// Check if transaction support is enabled
		if !b.config.EnableTransactions {
			return fmt.Errorf("%w: transaction support is disabled", ErrTransactionNotSupported)
		}

		// Validate txManager
		if b.txManager == nil {
			return fmt.Errorf("%w: transaction manager not initialized", ErrNotInitialized)
		}

		// Validate topic
		if err := validateTopic(topic); err != nil {
			return err
		}

		// Validate message
		if err := validateMessage(msg); err != nil {
			return err
		}

		// Check memory limit
		if err := b.checkMemoryLimit(); err != nil {
			return err
		}

		// Begin transaction if it doesn't exist yet
		// Note: begin will return error if transaction already exists
		timeout := b.config.TransactionTimeout
		if timeout == 0 {
			timeout = DefaultTransactionTimeoutDuration
		}
		_ = b.txManager.begin(txID, timeout) // Ignore "already exists" error

		// Add message to transaction buffer
		return b.txManager.addMessage(txID, topic, msg)
	})
}

// CommitTransaction commits a transaction.
func (b *InProcBroker) CommitTransaction(ctx context.Context, txID string) error {
	return b.executeWithObservability(ctx, OpPublish, "", func() error {
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

		// Check if transaction support is enabled
		if !b.config.EnableTransactions {
			return fmt.Errorf("%w: transaction support is disabled", ErrTransactionNotSupported)
		}

		// Validate txManager
		if b.txManager == nil {
			return fmt.Errorf("%w: transaction manager not initialized", ErrNotInitialized)
		}

		// Commit the transaction
		return b.txManager.commit(txID)
	})
}

// RollbackTransaction rolls back a transaction.
func (b *InProcBroker) RollbackTransaction(ctx context.Context, txID string) error {
	return b.executeWithObservability(ctx, OpPublish, "", func() error {
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

		// Check if transaction support is enabled
		if !b.config.EnableTransactions {
			return fmt.Errorf("%w: transaction support is disabled", ErrTransactionNotSupported)
		}

		// Validate txManager
		if b.txManager == nil {
			return fmt.Errorf("%w: transaction manager not initialized", ErrNotInitialized)
		}

		// Rollback the transaction
		return b.txManager.rollback(txID)
	})
}
