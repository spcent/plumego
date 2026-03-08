package tenant

import (
	"context"
	"sort"
	"sync"
	"time"
)

// SlidingWindowQuotaManager implements quota enforcement with a sliding window algorithm.
// This is more accurate than fixed windows and prevents burst traffic at window boundaries.
//
// SlidingWindowQuotaManager only enforces the per-minute window (RequestsPerMinute /
// TokensPerMinute, or the QuotaWindowMinute entry from Limits). For hour/day/month
// enforcement use WindowQuotaManager with InMemoryQuotaStore instead.
type SlidingWindowQuotaManager struct {
	provider QuotaConfigProvider
	windows  sync.Map // map[string]*slidingWindow
}

// slidingWindow tracks requests in a rolling time window.
type slidingWindow struct {
	mu           sync.RWMutex
	requestTimes []time.Time // sorted list of request timestamps (oldest first)
	tokenUsage   []tokenUsage
}

// tokenUsage tracks token consumption with timestamp.
type tokenUsage struct {
	time   time.Time
	tokens int
}

// NewSlidingWindowQuotaManager creates a new sliding window quota manager.
func NewSlidingWindowQuotaManager(provider QuotaConfigProvider) *SlidingWindowQuotaManager {
	return &SlidingWindowQuotaManager{
		provider: provider,
	}
}

// DeleteTenant removes the sliding window state for a tenant, reclaiming memory.
// Call this when a tenant is deleted or permanently deactivated.
func (m *SlidingWindowQuotaManager) DeleteTenant(tenantID string) {
	m.windows.Delete(tenantID)
}

// Allow checks if the request is within quota using sliding window algorithm.
func (m *SlidingWindowQuotaManager) Allow(ctx context.Context, tenantID string, req QuotaRequest) (QuotaResult, error) {
	cfg, err := m.provider.QuotaConfig(ctx, tenantID)
	if err != nil {
		return QuotaResult{Allowed: false}, err
	}

	reqLimit, tokLimit := slidingWindowEffectiveLimits(cfg)
	if reqLimit <= 0 && tokLimit <= 0 {
		return QuotaResult{Allowed: true}, nil
	}

	windowInterface, _ := m.windows.LoadOrStore(tenantID, &slidingWindow{
		requestTimes: make([]time.Time, 0),
		tokenUsage:   make([]tokenUsage, 0),
	})
	window := windowInterface.(*slidingWindow)

	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}
	if req.Requests <= 0 {
		req.Requests = 1
	}

	window.mu.Lock()
	defer window.mu.Unlock()

	cutoff := now.Add(-time.Minute)
	window.cleanup(cutoff)

	currentRequests := len(window.requestTimes)
	currentTokens := 0
	for _, u := range window.tokenUsage {
		currentTokens += u.tokens
	}

	if reqLimit > 0 && currentRequests+req.Requests > reqLimit {
		needed := currentRequests + req.Requests - reqLimit
		retryAfter := window.calculateRequestRetryAfter(now, needed)
		return QuotaResult{
			Allowed:           false,
			RemainingRequests: remaining(reqLimit, currentRequests),
			RemainingTokens:   remaining(tokLimit, currentTokens),
			RetryAfter:        retryAfter,
		}, ErrQuotaExceeded
	}

	if tokLimit > 0 && currentTokens+req.Tokens > tokLimit {
		needed := currentTokens + req.Tokens - tokLimit
		retryAfter := window.calculateTokenRetryAfter(now, needed)
		return QuotaResult{
			Allowed:           false,
			RemainingRequests: remaining(reqLimit, currentRequests),
			RemainingTokens:   remaining(tokLimit, currentTokens),
			RetryAfter:        retryAfter,
		}, ErrQuotaExceeded
	}

	for i := 0; i < req.Requests; i++ {
		window.requestTimes = append(window.requestTimes, now)
	}
	if req.Tokens > 0 {
		window.tokenUsage = append(window.tokenUsage, tokenUsage{
			time:   now,
			tokens: req.Tokens,
		})
	}

	return QuotaResult{
		Allowed:           true,
		RemainingRequests: remaining(reqLimit, currentRequests+req.Requests),
		RemainingTokens:   remaining(tokLimit, currentTokens+req.Tokens),
	}, nil
}

// cleanup removes entries older than the cutoff time using binary search.
func (w *slidingWindow) cleanup(cutoff time.Time) {
	// requestTimes: binary search for first entry after cutoff
	idx := sort.Search(len(w.requestTimes), func(i int) bool {
		return w.requestTimes[i].After(cutoff)
	})
	if idx > 0 {
		w.requestTimes = w.requestTimes[idx:]
	}

	// tokenUsage: binary search for first entry after cutoff
	tidx := sort.Search(len(w.tokenUsage), func(i int) bool {
		return w.tokenUsage[i].time.After(cutoff)
	})
	if tidx > 0 {
		w.tokenUsage = w.tokenUsage[tidx:]
	}
}

// calculateRequestRetryAfter returns how long until `needed` request slots free up.
// `needed` is the number of existing request entries that must expire to allow the request.
func (w *slidingWindow) calculateRequestRetryAfter(now time.Time, needed int) time.Duration {
	if len(w.requestTimes) == 0 || needed <= 0 {
		return 0
	}
	if needed > len(w.requestTimes) {
		needed = len(w.requestTimes)
	}
	// The oldest `needed` entries must expire. The last one (index needed-1) is the
	// latest-expiring among those. It expires 1 minute after its timestamp.
	expireAt := w.requestTimes[needed-1].Add(time.Minute)
	d := expireAt.Sub(now)
	if d < 0 {
		return 0
	}
	return d
}

// calculateTokenRetryAfter returns how long until enough tokens free up.
// `needed` is the total token count that must expire to allow the request.
func (w *slidingWindow) calculateTokenRetryAfter(now time.Time, needed int) time.Duration {
	if len(w.tokenUsage) == 0 || needed <= 0 {
		return 0
	}
	// Scan from oldest to newest, accumulating freed tokens, until we've covered `needed`.
	accumulated := 0
	var expireEntry time.Time
	for _, u := range w.tokenUsage {
		accumulated += u.tokens
		expireEntry = u.time
		if accumulated >= needed {
			break
		}
	}
	// expireEntry is the timestamp of the last tokenUsage entry that must expire.
	expireAt := expireEntry.Add(time.Minute)
	d := expireAt.Sub(now)
	if d < 0 {
		return 0
	}
	return d
}

// slidingWindowEffectiveLimits returns the per-minute request and token limits for
// the sliding window manager. When cfg.Limits is set, only the QuotaWindowMinute
// entry is used; all other windows require WindowQuotaManager.
func slidingWindowEffectiveLimits(cfg QuotaConfig) (reqLimit, tokLimit int) {
	if len(cfg.Limits) > 0 {
		for _, l := range cfg.Limits {
			if l.Window == QuotaWindowMinute {
				return l.Requests, l.Tokens
			}
		}
		// Limits defined but no minute window — this manager cannot enforce
		// hour/day/month limits; callers should use WindowQuotaManager.
		return 0, 0
	}
	return cfg.RequestsPerMinute, cfg.TokensPerMinute
}
