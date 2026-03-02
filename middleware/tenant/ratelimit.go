package tenant

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/tenant"
)

// TenantRateLimitOptions configures rate limit enforcement.
type TenantRateLimitOptions struct {
	Limiter    tenant.RateLimiter
	Estimator  func(*http.Request) int64
	Hooks      tenant.Hooks
	OnRejected func(http.ResponseWriter, *http.Request, tenant.RateLimitResult)
}

// TenantRateLimit enforces per-tenant rate limits.
func TenantRateLimit(options TenantRateLimitOptions) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if options.Limiter == nil {
				next.ServeHTTP(w, r)
				return
			}

			tenantID := tenant.TenantIDFromContext(r.Context())
			if tenantID == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokens := int64(1)
			if options.Estimator != nil {
				if t := options.Estimator(r); t > 0 {
					tokens = t
				}
			}

			result, err := options.Limiter.Allow(r.Context(), tenantID, tenant.RateLimitRequest{
				Tokens: tokens,
				Now:    time.Now().UTC(),
			})
			allowed := err == nil && result.Allowed

			// Status reflects the actual outcome: 200 when allowed, 429 when denied.
			status := http.StatusOK
			if !allowed {
				status = http.StatusTooManyRequests
			}

			options.Hooks.RateLimit(r.Context(), tenant.RateLimitDecision{
				TenantID:   tenantID,
				Allowed:    allowed,
				Tokens:     tokens,
				Remaining:  result.Remaining,
				Limit:      result.Limit,
				Burst:      result.Burst,
				RetryAfter: result.RetryAfter,
				Status:     status,
			})

			if allowed {
				next.ServeHTTP(w, r)
				return
			}

			setRetryAfterHeader(w, result.RetryAfter)
			setRateLimitHeaders(w, result.Limit, result.Remaining)

			if options.OnRejected != nil {
				options.OnRejected(w, r, result)
				return
			}

			writeTenantError(w, r, status, "tenant_rate_limited", "tenant rate limit exceeded", contract.CategoryRateLimit)
		})
	}
}
