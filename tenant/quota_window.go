package tenant

import (
	"context"
	"time"
)

// WindowQuotaManager enforces per-tenant quota limits for multiple windows.
// It uses a QuotaStore to atomically reserve usage per window.
type WindowQuotaManager struct {
	provider QuotaConfigProvider
	store    QuotaStore
}

// NewWindowQuotaManager creates a quota manager backed by a QuotaStore.
func NewWindowQuotaManager(provider QuotaConfigProvider, store QuotaStore) *WindowQuotaManager {
	return &WindowQuotaManager{
		provider: provider,
		store:    store,
	}
}

// Allow checks quota usage for a tenant.
func (m *WindowQuotaManager) Allow(ctx context.Context, tenantID string, req QuotaRequest) (QuotaResult, error) {
	if m == nil || m.provider == nil || m.store == nil {
		return QuotaResult{Allowed: true}, nil
	}

	cfg, err := m.provider.QuotaConfig(ctx, tenantID)
	if err != nil {
		return QuotaResult{Allowed: false}, err
	}

	limits := normalizeQuotaLimits(cfg)
	if len(limits) == 0 {
		return QuotaResult{Allowed: true}, nil
	}

	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	}
	if req.Requests <= 0 {
		req.Requests = 1
	}

	type reservation struct {
		window      QuotaWindow
		windowStart time.Time
	}

	var (
		reserved []reservation
		minReq   = -1
		minTok   = -1
	)

	for _, limit := range limits {
		if limit.Requests <= 0 && limit.Tokens <= 0 {
			continue
		}

		windowStart := quotaWindowStart(req.Now, limit.Window)
		usage, allowed, err := m.store.Reserve(
			ctx,
			tenantID,
			limit.Window,
			windowStart,
			req.Requests,
			req.Tokens,
			limit.Requests,
			limit.Tokens,
		)
		if err != nil || !allowed {
			for _, item := range reserved {
				_ = m.store.Release(ctx, tenantID, item.window, item.windowStart, req.Requests, req.Tokens)
			}

			retryAfter := time.Until(quotaWindowEnd(windowStart, limit.Window))
			if retryAfter < 0 {
				retryAfter = 0
			}

			result := QuotaResult{
				Allowed:           false,
				RemainingRequests: remaining(limit.Requests, usage.Requests),
				RemainingTokens:   remaining(limit.Tokens, usage.Tokens),
				RetryAfter:        retryAfter,
			}
			if err != nil {
				return result, err
			}
			return result, ErrQuotaExceeded
		}

		reserved = append(reserved, reservation{window: limit.Window, windowStart: windowStart})
		minReq = minRemaining(minReq, remaining(limit.Requests, usage.Requests))
		minTok = minRemaining(minTok, remaining(limit.Tokens, usage.Tokens))
	}

	return QuotaResult{
		Allowed:           true,
		RemainingRequests: minReq,
		RemainingTokens:   minTok,
	}, nil
}

func minRemaining(current, candidate int) int {
	if candidate < 0 {
		return current
	}
	if current < 0 {
		return candidate
	}
	if candidate < current {
		return candidate
	}
	return current
}
