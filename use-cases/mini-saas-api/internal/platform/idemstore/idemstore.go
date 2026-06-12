// Package idemstore is an in-memory implementation of the stable
// store/idempotency contract (Store + HashAwareStore). Replace with a
// database-backed implementation for multi-instance deployments.
package idemstore

import (
	"context"
	"sync"
	"time"

	"github.com/spcent/plumego/store/idempotency"
)

// Memory is a mutex-guarded in-memory idempotency store.
type Memory struct {
	mu      sync.Mutex
	records map[string]idempotency.Record
	now     func() time.Time
}

var (
	_ idempotency.Store          = (*Memory)(nil)
	_ idempotency.HashAwareStore = (*Memory)(nil)
)

// NewMemory returns an empty in-memory idempotency store.
func NewMemory() *Memory {
	return &Memory{
		records: make(map[string]idempotency.Record),
		now:     time.Now,
	}
}

// Get returns a usable record; missing and expired keys report found=false.
func (m *Memory) Get(_ context.Context, key string) (idempotency.Record, bool, error) {
	if err := idempotency.ValidateKey(key); err != nil {
		return idempotency.Record{}, false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.records[key]
	if !ok {
		return idempotency.Record{}, false, nil
	}
	if m.expired(rec) {
		delete(m.records, key)
		return idempotency.Record{}, false, nil
	}
	return rec.Clone(), true, nil
}

// PutIfAbsent claims the key when no usable record holds it.
func (m *Memory) PutIfAbsent(_ context.Context, record idempotency.Record) (bool, error) {
	if err := idempotency.ValidateRecord(record); err != nil {
		return false, err
	}
	if m.expired(record) {
		return false, idempotency.ErrExpired
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.records[record.Key]; ok && !m.expired(existing) {
		return false, nil
	}
	m.records[record.Key] = record.Clone()
	return true, nil
}

// Complete marks an existing usable record complete with a replayable response.
func (m *Memory) Complete(_ context.Context, key string, response []byte) error {
	return m.complete(key, "", response)
}

// CompleteWithRequestHash completes only when the stored hash matches.
func (m *Memory) CompleteWithRequestHash(_ context.Context, key, requestHash string, response []byte) error {
	return m.complete(key, requestHash, response)
}

func (m *Memory) complete(key, requestHash string, response []byte) error {
	if err := idempotency.ValidateKey(key); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.records[key]
	if !ok || m.expired(rec) {
		delete(m.records, key)
		return idempotency.ErrNotFound
	}
	if requestHash != "" && rec.RequestHash != requestHash {
		return idempotency.ErrRequestMismatch
	}
	rec.Status = idempotency.StatusCompleted
	rec.Response = append([]byte(nil), response...)
	rec.UpdatedAt = m.now().UTC()
	m.records[key] = rec
	return nil
}

// Delete removes a record; missing keys return ErrNotFound.
func (m *Memory) Delete(_ context.Context, key string) error {
	if err := idempotency.ValidateKey(key); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.records[key]; !ok {
		return idempotency.ErrNotFound
	}
	delete(m.records, key)
	return nil
}

func (m *Memory) expired(rec idempotency.Record) bool {
	return !rec.ExpiresAt.IsZero() && !rec.ExpiresAt.After(m.now().UTC())
}
