package usage

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound     = errors.New("usage records not found")
	ErrInvalidUsage = errors.New("invalid usage record")
)

type UsageRecord struct {
	TenantID   string    `json:"tenant_id"`
	Resource   string    `json:"resource"`
	Count      int64     `json:"count"`
	RecordedAt time.Time `json:"recorded_at"`
}

type InMemoryUsageStore struct {
	mu      sync.RWMutex
	records map[string][]UsageRecord
}

func NewInMemoryUsageStore() *InMemoryUsageStore {
	return &InMemoryUsageStore{records: make(map[string][]UsageRecord)}
}

func (s *InMemoryUsageStore) Record(_ context.Context, tenantID string, resource string, count int64) error {
	if s == nil {
		return ErrInvalidUsage
	}
	tenantID = strings.TrimSpace(tenantID)
	resource = strings.TrimSpace(resource)
	if tenantID == "" || resource == "" || count <= 0 {
		return ErrInvalidUsage
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[tenantID] = append(s.records[tenantID], UsageRecord{
		TenantID:   tenantID,
		Resource:   resource,
		Count:      count,
		RecordedAt: time.Now().UTC(),
	})
	return nil
}

func (s *InMemoryUsageStore) Report(_ context.Context, tenantID string) ([]UsageRecord, error) {
	if s == nil {
		return nil, ErrNotFound
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	records := s.records[tenantID]
	if len(records) == 0 {
		return nil, ErrNotFound
	}
	out := make([]UsageRecord, len(records))
	copy(out, records)
	return out, nil
}
