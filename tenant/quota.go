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
	// Limits defines quota windows (minute/hour/day/month).
	// When non-empty, Limits take precedence over RequestsPerMinute/TokensPerMinute.
	// InMemoryQuotaManager enforces only the first valid limit window.
	// Use WindowQuotaManager for full multi-window enforcement.
	Limits []QuotaLimit
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
	windowEnd   time.Time
	requests    int
	tokens      int
}

// InMemoryQuotaManager is a single-window in-memory quota manager.
// It respects QuotaConfig.Limits by using the first valid limit entry.
// For multi-window enforcement (hour + day + month simultaneously),
// use WindowQuotaManager with an InMemoryQuotaStore instead.
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
// It uses normalizeQuotaLimits to resolve the effective limit, honouring
// QuotaConfig.Limits when set and falling back to RequestsPerMinute/TokensPerMinute.
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

	// Resolve effective limits: Limits array takes precedence.
	limits := normalizeQuotaLimits(cfg)
	if len(limits) == 0 {
		return QuotaResult{Allowed: true}, nil
	}

	// Use the first limit entry as the active window.
	// Users needing simultaneous multi-window checks should use WindowQuotaManager.
	limit := limits[0]
	limitRequests := limit.Requests
	limitTokens := limit.Tokens

	if limitRequests <= 0 && limitTokens <= 0 {
		return QuotaResult{Allowed: true}, nil
	}

	windowStart := quotaWindowStart(req.Now, limit.Window)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Evict stale counters periodically to prevent unbounded memory growth.
	m.evictStaleLocked(req.Now)

	counter := m.counters[tenantID]
	if counter == nil || !counter.windowStart.Equal(windowStart) {
		counter = &quotaCounter{
			windowStart: windowStart,
			windowEnd:   quotaWindowEnd(windowStart, limit.Window),
		}
		m.counters[tenantID] = counter
	}

	nextRequests := counter.requests + req.Requests
	nextTokens := counter.tokens + req.Tokens

	if (limitRequests > 0 && nextRequests > limitRequests) || (limitTokens > 0 && nextTokens > limitTokens) {
		retryAfter := time.Until(counter.windowEnd)
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

// evictStaleLocked removes counters whose window has expired.
// Must be called with m.mu held.
func (m *InMemoryQuotaManager) evictStaleLocked(now time.Time) {
	for tenantID, counter := range m.counters {
		if now.After(counter.windowEnd) {
			delete(m.counters, tenantID)
		}
	}
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
