package admin

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("tenant not found")

const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
)

type TenantRecord struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	SuspendedAt *time.Time `json:"suspended_at,omitempty"`
}

type InMemoryStore struct {
	mu      sync.RWMutex
	tenants map[string]TenantRecord
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{tenants: make(map[string]TenantRecord)}
}

func (s *InMemoryStore) Create(_ context.Context, record TenantRecord) (TenantRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if record.Status == "" {
		record.Status = StatusActive
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	s.tenants[record.ID] = record
	return record, nil
}

func (s *InMemoryStore) Get(_ context.Context, id string) (TenantRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.tenants[id]
	if !ok {
		return TenantRecord{}, ErrNotFound
	}
	return record, nil
}

func (s *InMemoryStore) Suspend(_ context.Context, id string) (TenantRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.tenants[id]
	if !ok {
		return TenantRecord{}, ErrNotFound
	}
	now := time.Now().UTC()
	record.Status = StatusSuspended
	record.SuspendedAt = &now
	s.tenants[id] = record
	return record, nil
}

func (s *InMemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tenants[id]; !ok {
		return ErrNotFound
	}
	delete(s.tenants, id)
	return nil
}
