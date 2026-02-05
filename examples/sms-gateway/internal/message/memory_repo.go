package message

import (
	"context"
	"sync"
	"time"
)

// MemoryRepository is an in-memory implementation of Store for demos/tests.
type MemoryRepository struct {
	mu       sync.RWMutex
	messages map[string]Message
}

// NewMemoryRepository creates an in-memory repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		messages: make(map[string]Message),
	}
}

func (r *MemoryRepository) Insert(ctx context.Context, msg Message) error {
	if r == nil {
		return ErrMessageNotFound
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.messages[msg.ID]; exists {
		return ErrMessageExists
	}
	r.messages[msg.ID] = msg
	return nil
}

func (r *MemoryRepository) Get(ctx context.Context, id string) (Message, bool, error) {
	if r == nil {
		return Message{}, false, ErrMessageNotFound
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg, ok := r.messages[id]
	if !ok {
		return Message{}, false, nil
	}
	return msg, true, nil
}

func (r *MemoryRepository) UpdateStatus(ctx context.Context, id string, from Status, to Status, reason Reason) error {
	if r == nil {
		return ErrMessageNotFound
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	msg, ok := r.messages[id]
	if !ok {
		return ErrMessageNotFound
	}
	if msg.Status != from {
		return ErrMessageStateConflict
	}
	msg.Status = to
	msg.ReasonCode = reason.Code
	msg.ReasonDetail = reason.Detail
	if to == StatusSent && msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}
	msg.UpdatedAt = time.Now().UTC()
	msg.Version++
	r.messages[id] = msg
	return nil
}

func (r *MemoryRepository) UpdateProvider(ctx context.Context, id string, provider string) error {
	if r == nil {
		return ErrMessageNotFound
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	msg, ok := r.messages[id]
	if !ok {
		return ErrMessageNotFound
	}
	msg.Provider = provider
	msg.UpdatedAt = time.Now().UTC()
	msg.Version++
	r.messages[id] = msg
	return nil
}

func (r *MemoryRepository) RecordAttempt(ctx context.Context, id string, attempts int) error {
	if r == nil {
		return ErrMessageNotFound
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	msg, ok := r.messages[id]
	if !ok {
		return ErrMessageNotFound
	}
	msg.Attempts = attempts
	msg.UpdatedAt = time.Now().UTC()
	msg.Version++
	r.messages[id] = msg
	return nil
}

func (r *MemoryRepository) RecordDLQ(ctx context.Context, id string, reason Reason) error {
	if r == nil {
		return ErrMessageNotFound
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	msg, ok := r.messages[id]
	if !ok {
		return ErrMessageNotFound
	}
	msg.Status = StatusFailed
	msg.ReasonCode = reason.Code
	msg.ReasonDetail = reason.Detail
	msg.UpdatedAt = time.Now().UTC()
	msg.Version++
	r.messages[id] = msg
	return nil
}
