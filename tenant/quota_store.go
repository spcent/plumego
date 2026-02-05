package tenant

import (
	"context"
	"sync"
	"time"
)

// QuotaUsage tracks quota usage for a window.
type QuotaUsage struct {
	WindowStart time.Time
	WindowEnd   time.Time
	Requests    int
	Tokens      int
}

// QuotaStore provides atomic quota reservation for a single window.
type QuotaStore interface {
	// Reserve attempts to add usage in a window. It should be atomic and only
	// update usage when within limits. The returned usage should reflect the
	// stored values after the call.
	Reserve(ctx context.Context, tenantID string, window QuotaWindow, windowStart time.Time, deltaRequests, deltaTokens, limitRequests, limitTokens int) (QuotaUsage, bool, error)
	// Release rolls back a previous reservation.
	Release(ctx context.Context, tenantID string, window QuotaWindow, windowStart time.Time, deltaRequests, deltaTokens int) error
}

// InMemoryQuotaStore is a simple in-memory quota store.
type InMemoryQuotaStore struct {
	mu              sync.Mutex
	entries         map[quotaKey]*QuotaUsage
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

// InMemoryQuotaStoreConfig configures the in-memory quota store.
type InMemoryQuotaStoreConfig struct {
	CleanupInterval time.Duration
}

// NewInMemoryQuotaStore creates a new in-memory quota store.
func NewInMemoryQuotaStore(opts ...InMemoryQuotaStoreConfig) *InMemoryQuotaStore {
	cfg := InMemoryQuotaStoreConfig{
		CleanupInterval: 5 * time.Minute,
	}
	if len(opts) > 0 {
		if opts[0].CleanupInterval > 0 {
			cfg.CleanupInterval = opts[0].CleanupInterval
		}
	}

	return &InMemoryQuotaStore{
		entries:         make(map[quotaKey]*QuotaUsage),
		cleanupInterval: cfg.CleanupInterval,
		lastCleanup:     time.Now().UTC(),
	}
}

// Reserve attempts to add usage for a window.
func (s *InMemoryQuotaStore) Reserve(ctx context.Context, tenantID string, window QuotaWindow, windowStart time.Time, deltaRequests, deltaTokens, limitRequests, limitTokens int) (QuotaUsage, bool, error) {
	if s == nil {
		return QuotaUsage{}, false, ErrQuotaExceeded
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.maybeCleanupLocked(time.Now().UTC())

	key := quotaKey{tenantID: tenantID, window: window, windowStart: windowStart}
	usage := s.entries[key]
	if usage == nil {
		usage = &QuotaUsage{
			WindowStart: windowStart,
			WindowEnd:   quotaWindowEnd(windowStart, window),
		}
		s.entries[key] = usage
	}

	nextRequests := usage.Requests + deltaRequests
	nextTokens := usage.Tokens + deltaTokens

	if (limitRequests > 0 && nextRequests > limitRequests) || (limitTokens > 0 && nextTokens > limitTokens) {
		return *usage, false, nil
	}

	usage.Requests = nextRequests
	usage.Tokens = nextTokens

	return *usage, true, nil
}

// Release rolls back a previous reservation.
func (s *InMemoryQuotaStore) Release(ctx context.Context, tenantID string, window QuotaWindow, windowStart time.Time, deltaRequests, deltaTokens int) error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := quotaKey{tenantID: tenantID, window: window, windowStart: windowStart}
	usage := s.entries[key]
	if usage == nil {
		return nil
	}

	usage.Requests -= deltaRequests
	if usage.Requests < 0 {
		usage.Requests = 0
	}
	usage.Tokens -= deltaTokens
	if usage.Tokens < 0 {
		usage.Tokens = 0
	}

	return nil
}

// Usage returns current usage for a window (best-effort, for inspection/testing).
func (s *InMemoryQuotaStore) Usage(tenantID string, window QuotaWindow, windowStart time.Time) (QuotaUsage, bool) {
	if s == nil {
		return QuotaUsage{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	key := quotaKey{tenantID: tenantID, window: window, windowStart: windowStart}
	usage := s.entries[key]
	if usage == nil {
		return QuotaUsage{}, false
	}
	return *usage, true
}

type quotaKey struct {
	tenantID    string
	window      QuotaWindow
	windowStart time.Time
}

func (s *InMemoryQuotaStore) maybeCleanupLocked(now time.Time) {
	if s.cleanupInterval <= 0 {
		return
	}
	if now.Sub(s.lastCleanup) < s.cleanupInterval {
		return
	}

	for key, usage := range s.entries {
		if usage.WindowEnd.Before(now) {
			delete(s.entries, key)
		}
	}

	s.lastCleanup = now
}
