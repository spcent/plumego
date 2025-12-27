package webhookout

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type MemStore struct {
	mu         sync.RWMutex
	targets    map[string]Target
	deliveries map[string]Delivery
}

func NewMemStore() *MemStore {
	return &MemStore{
		targets:    make(map[string]Target),
		deliveries: make(map[string]Delivery),
	}
}

func (s *MemStore) CreateTarget(ctx context.Context, t Target) (Target, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now
	s.targets[t.ID] = t
	return t, nil
}

func (s *MemStore) UpdateTarget(ctx context.Context, id string, patch TargetPatch) (Target, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.targets[id]
	if !ok {
		return Target{}, ErrNotFound
	}
	if patch.Name != nil {
		t.Name = *patch.Name
	}
	if patch.URL != nil {
		t.URL = *patch.URL
	}
	if patch.Secret != nil {
		t.Secret = *patch.Secret
	}
	if patch.Events != nil {
		t.Events = *patch.Events
	}
	if patch.Enabled != nil {
		t.Enabled = *patch.Enabled
	}
	if patch.Headers != nil {
		t.Headers = *patch.Headers
	}
	if patch.TimeoutMs != nil {
		t.TimeoutMs = *patch.TimeoutMs
	}
	if patch.MaxRetries != nil {
		t.MaxRetries = *patch.MaxRetries
	}
	if patch.BackoffBaseMs != nil {
		t.BackoffBaseMs = *patch.BackoffBaseMs
	}
	if patch.BackoffMaxMs != nil {
		t.BackoffMaxMs = *patch.BackoffMaxMs
	}
	if patch.RetryOn429 != nil {
		t.RetryOn429 = patch.RetryOn429
	}

	t.UpdatedAt = time.Now().UTC()
	s.targets[id] = t
	return t, nil
}

func (s *MemStore) GetTarget(ctx context.Context, id string) (Target, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.targets[id]
	return t, ok
}

func (s *MemStore) ListTargets(ctx context.Context, filter TargetFilter) ([]Target, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Target, 0, len(s.targets))
	for _, t := range s.targets {
		if filter.Enabled != nil && t.Enabled != *filter.Enabled {
			continue
		}
		if filter.Event != "" && !eventMatch(t.Events, filter.Event) {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}

func (s *MemStore) CreateDelivery(ctx context.Context, d Delivery) (Delivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	d.CreatedAt = now
	d.UpdatedAt = now

	if d.PayloadJSON != nil {
		b := make([]byte, len(d.PayloadJSON))
		copy(b, d.PayloadJSON)
		d.PayloadJSON = b
	}

	s.deliveries[d.ID] = d
	return d, nil
}

func (s *MemStore) UpdateDelivery(ctx context.Context, id string, patch DeliveryPatch) (Delivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.deliveries[id]
	if !ok {
		return Delivery{}, ErrNotFound
	}

	if patch.Attempt != nil {
		d.Attempt = *patch.Attempt
	}
	if patch.Status != nil {
		d.Status = *patch.Status
	}
	if patch.NextAt != nil {
		d.NextAt = patch.NextAt
	}
	if patch.LastHTTPStatus != nil {
		d.LastHTTPStatus = *patch.LastHTTPStatus
	}
	if patch.LastError != nil {
		d.LastError = *patch.LastError
	}
	if patch.LastDurationMs != nil {
		d.LastDurationMs = *patch.LastDurationMs
	}
	if patch.LastRespSnippet != nil {
		d.LastRespSnippet = *patch.LastRespSnippet
	}
	if patch.PayloadJSON != nil {
		// store a copy to avoid caller mutation
		b := make([]byte, len(*patch.PayloadJSON))
		copy(b, *patch.PayloadJSON)
		d.PayloadJSON = b
	}

	d.UpdatedAt = time.Now().UTC()
	s.deliveries[id] = d
	return d, nil
}

func (s *MemStore) GetDelivery(ctx context.Context, id string) (Delivery, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.deliveries[id]
	return d, ok
}

func (s *MemStore) ListDeliveries(ctx context.Context, filter DeliveryFilter) ([]Delivery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	// cursor is naive: treat it as "created_at after id lexicographically" not perfect.
	// For MemStore demo, we sort by CreatedAt then ID.
	all := make([]Delivery, 0, len(s.deliveries))
	for _, d := range s.deliveries {
		if filter.TargetID != nil && d.TargetID != *filter.TargetID {
			continue
		}
		if filter.Event != nil && d.EventType != *filter.Event {
			continue
		}
		if filter.Status != nil && d.Status != *filter.Status {
			continue
		}
		if filter.From != nil && d.CreatedAt.Before(*filter.From) {
			continue
		}
		if filter.To != nil && d.CreatedAt.After(*filter.To) {
			continue
		}

		all = append(all, d)
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].CreatedAt.Equal(all[j].CreatedAt) {
			return all[i].ID < all[j].ID
		}
		return all[i].CreatedAt.Before(all[j].CreatedAt)
	})

	start := 0
	if filter.Cursor != "" {
		for i := range all {
			if strings.EqualFold(all[i].ID, filter.Cursor) {
				start = i + 1
				break
			}
		}
	}

	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], nil
}

func eventMatch(events []string, event string) bool {
	if event == "" {
		return true
	}
	for _, e := range events {
		if e == "*" || strings.EqualFold(e, event) {
			return true
		}
	}
	return false
}
