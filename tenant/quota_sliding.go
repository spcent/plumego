package tenant

import (
	"context"
	"sort"
	"sync"
	"time"
)

// SlidingWindowQuotaManager implements quota enforcement with a sliding window algorithm.
// This is more accurate than fixed windows and prevents burst traffic at window boundaries.
type SlidingWindowQuotaManager struct {
	provider QuotaConfigProvider
	windows  sync.Map // map[string]*slidingWindow
}

// slidingWindow tracks requests in a rolling time window.
type slidingWindow struct {
	mu             sync.RWMutex
	requestTimes   []time.Time // sorted list of request timestamps
	tokenUsage     []tokenUsage
	lastCleanup    time.Time
	requestsPerMin int
	tokensPerMin   int
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

// Allow checks if the request is within quota using sliding window algorithm.
func (m *SlidingWindowQuotaManager) Allow(ctx context.Context, tenantID string, req QuotaRequest) (QuotaResult, error) {
	// Get quota configuration
	cfg, err := m.provider.QuotaConfig(ctx, tenantID)
	if err != nil {
		return QuotaResult{Allowed: false}, err
	}

	// Get or create window
	windowInterface, _ := m.windows.LoadOrStore(tenantID, &slidingWindow{
		requestTimes:   make([]time.Time, 0),
		tokenUsage:     make([]tokenUsage, 0),
		lastCleanup:    time.Now(),
		requestsPerMin: cfg.RequestsPerMinute,
		tokensPerMin:   cfg.TokensPerMinute,
	})
	window := windowInterface.(*slidingWindow)

	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}

	window.mu.Lock()
	defer window.mu.Unlock()

	// Update quota limits if changed
	window.requestsPerMin = cfg.RequestsPerMinute
	window.tokensPerMin = cfg.TokensPerMinute

	// Clean up old entries (older than 1 minute)
	cutoff := now.Add(-time.Minute)
	window.cleanup(cutoff)

	// Count requests and tokens in current window
	currentRequests := len(window.requestTimes)
	currentTokens := 0
	for _, usage := range window.tokenUsage {
		currentTokens += usage.tokens
	}

	// Check request quota
	if cfg.RequestsPerMinute > 0 && currentRequests+req.Requests > cfg.RequestsPerMinute {
		retryAfter := window.calculateRetryAfter(now, cfg.RequestsPerMinute, true)
		return QuotaResult{
			Allowed:           false,
			RemainingRequests: remaining(cfg.RequestsPerMinute, currentRequests),
			RemainingTokens:   remaining(cfg.TokensPerMinute, currentTokens),
			RetryAfter:        retryAfter,
		}, nil
	}

	// Check token quota
	if cfg.TokensPerMinute > 0 && currentTokens+req.Tokens > cfg.TokensPerMinute {
		retryAfter := window.calculateRetryAfter(now, cfg.TokensPerMinute, false)
		return QuotaResult{
			Allowed:           false,
			RemainingRequests: remaining(cfg.RequestsPerMinute, currentRequests),
			RemainingTokens:   remaining(cfg.TokensPerMinute, currentTokens),
			RetryAfter:        retryAfter,
		}, nil
	}

	// Record the request
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
		RemainingRequests: remaining(cfg.RequestsPerMinute, currentRequests+req.Requests),
		RemainingTokens:   remaining(cfg.TokensPerMinute, currentTokens+req.Tokens),
		RetryAfter:        0,
	}, nil
}

// cleanup removes entries older than the cutoff time.
func (w *slidingWindow) cleanup(cutoff time.Time) {
	// Clean up request times using binary search
	idx := sort.Search(len(w.requestTimes), func(i int) bool {
		return w.requestTimes[i].After(cutoff)
	})
	if idx > 0 {
		w.requestTimes = w.requestTimes[idx:]
	}

	// Clean up token usage - find first entry after cutoff
	tokenIdx := -1
	for i, usage := range w.tokenUsage {
		if usage.time.After(cutoff) {
			tokenIdx = i
			break
		}
	}
	if tokenIdx > 0 {
		w.tokenUsage = w.tokenUsage[tokenIdx:]
	} else if tokenIdx < 0 && len(w.tokenUsage) > 0 {
		// All entries are expired, clear the slice
		w.tokenUsage = w.tokenUsage[:0]
	}

	w.lastCleanup = time.Now()
}

// calculateRetryAfter calculates how long to wait before the next request would be allowed.
func (w *slidingWindow) calculateRetryAfter(now time.Time, limit int, isRequest bool) time.Duration {
	var times []time.Time

	if isRequest {
		times = w.requestTimes
	} else {
		times = make([]time.Time, len(w.tokenUsage))
		for i, usage := range w.tokenUsage {
			times[i] = usage.time
		}
	}

	if len(times) == 0 {
		return 0
	}

	// Find the oldest entry that would need to expire
	if len(times) >= limit {
		oldestToExpire := times[len(times)-limit]
		expiresAt := oldestToExpire.Add(time.Minute)
		retryAfter := expiresAt.Sub(now)
		if retryAfter < 0 {
			return 0
		}
		return retryAfter
	}

	return 0
}
