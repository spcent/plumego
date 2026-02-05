package tenant

import (
	"net/http"
	"strconv"
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

			if result.RetryAfter > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
			}
			if result.Limit > 0 {
				w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
			}
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))

			if options.OnRejected != nil {
				options.OnRejected(w, r, result)
				return
			}

			contract.WriteError(w, r, contract.APIError{
				Status:   status,
				Code:     "tenant_rate_limited",
				Message:  "tenant rate limit exceeded",
				Category: contract.CategoryRateLimit,
			})
		})
	}
}
