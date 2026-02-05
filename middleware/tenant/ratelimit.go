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
				tokens = options.Estimator(r)
				if tokens <= 0 {
					tokens = 1
				}
			}

			result, err := options.Limiter.Allow(r.Context(), tenantID, tenant.RateLimitRequest{
				Tokens: tokens,
				Now:    time.Now().UTC(),
			})
			allowed := err == nil && result.Allowed
			status := http.StatusTooManyRequests

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
