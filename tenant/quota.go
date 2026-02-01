package tenant

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrQuotaExceeded = errors.New("quota exceeded")

// QuotaConfig defines per-tenant quota limits.
// Zero values mean unlimited.
type QuotaConfig struct {
	RequestsPerMinute int
	TokensPerMinute   int
}

// QuotaRequest is the input to quota checks.
type QuotaRequest struct {
	Requests int
	Tokens   int
	Now      time.Time
}

// QuotaResult describes a quota decision.
type QuotaResult struct {
	Allowed           bool
	RemainingRequests int
	RemainingTokens   int
	RetryAfter        time.Duration
}

// QuotaManager enforces tenant quota.
type QuotaManager interface {
	Allow(ctx context.Context, tenantID string, req QuotaRequest) (QuotaResult, error)
}

type quotaCounter struct {
	windowStart time.Time
	requests    int
	tokens      int
}

// InMemoryQuotaManager is a fixed-window, in-memory quota manager.
type InMemoryQuotaManager struct {
	mu       sync.Mutex
	provider QuotaConfigProvider
	counters map[string]*quotaCounter
}

// NewInMemoryQuotaManager builds a quota manager from a config provider.
func NewInMemoryQuotaManager(provider QuotaConfigProvider) *InMemoryQuotaManager {
	return &InMemoryQuotaManager{
		provider: provider,
		counters: make(map[string]*quotaCounter),
	}
}

// Allow checks quota usage for a tenant.
func (m *InMemoryQuotaManager) Allow(ctx context.Context, tenantID string, req QuotaRequest) (QuotaResult, error) {
	if m == nil || m.provider == nil {
		return QuotaResult{Allowed: true}, nil
	}

	cfg, err := m.provider.QuotaConfig(ctx, tenantID)
	if err != nil {
		return QuotaResult{Allowed: false}, err
	}

	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	}
	if req.Requests <= 0 {
		req.Requests = 1
	}

	limitRequests := cfg.RequestsPerMinute
	limitTokens := cfg.TokensPerMinute

	if limitRequests <= 0 && limitTokens <= 0 {
		return QuotaResult{Allowed: true}, nil
	}

	windowStart := req.Now.Truncate(time.Minute)

	m.mu.Lock()
	defer m.mu.Unlock()

	counter := m.counters[tenantID]
	if counter == nil || !counter.windowStart.Equal(windowStart) {
		counter = &quotaCounter{windowStart: windowStart}
		m.counters[tenantID] = counter
	}

	nextRequests := counter.requests + req.Requests
	nextTokens := counter.tokens + req.Tokens

	if (limitRequests > 0 && nextRequests > limitRequests) || (limitTokens > 0 && nextTokens > limitTokens) {
		retryAfter := time.Until(windowStart.Add(time.Minute))
		if retryAfter < 0 {
			retryAfter = 0
		}
		return QuotaResult{
			Allowed:           false,
			RemainingRequests: remaining(limitRequests, counter.requests),
			RemainingTokens:   remaining(limitTokens, counter.tokens),
			RetryAfter:        retryAfter,
		}, ErrQuotaExceeded
	}

	counter.requests = nextRequests
	counter.tokens = nextTokens

	return QuotaResult{
		Allowed:           true,
		RemainingRequests: remaining(limitRequests, counter.requests),
		RemainingTokens:   remaining(limitTokens, counter.tokens),
	}, nil
}

func remaining(limit, used int) int {
	if limit <= 0 {
		return -1
	}
	remain := limit - used
	if remain < 0 {
		return 0
	}
	return remain
}
