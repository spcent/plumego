package tenantmw

import (
	"net/http"
	"strconv"
	"time"

	"github.com/spcent/plumego/contract"
	mw "github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/tenant"
)

// TenantQuotaOptions configures quota enforcement.
type TenantQuotaOptions struct {
	Manager      tenant.QuotaManager
	TokensHeader string
	Estimator    func(*http.Request) int
	Hooks        tenant.Hooks
	OnRejected   func(http.ResponseWriter, *http.Request, tenant.QuotaResult)
}

// TenantQuota enforces tenant quota limits.
// On every request (allowed or denied) it sets X-Quota-Remaining-Requests and
// X-Quota-Remaining-Tokens headers so clients can monitor their quota consumption.
func TenantQuota(options TenantQuotaOptions) mw.Middleware {
	tokensHeader := headerOrDefault(options.TokensHeader, defaultTokensHeader)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if options.Manager == nil {
				next.ServeHTTP(w, r)
				return
			}

			tenantID := tenant.TenantIDFromContext(r.Context())
			tokens := 0
			if options.Estimator != nil {
				tokens = options.Estimator(r)
			} else if header := r.Header.Get(tokensHeader); header != "" {
				if value, err := strconv.Atoi(header); err == nil {
					tokens = value
				}
			}

			result, err := options.Manager.Allow(r.Context(), tenantID, tenant.QuotaRequest{
				Requests: 1,
				Tokens:   tokens,
				Now:      time.Now().UTC(),
			})
			allowed := err == nil && result.Allowed

			// Status reflects the actual outcome: 200 when allowed, 429 when denied.
			status := http.StatusOK
			if !allowed {
				status = http.StatusTooManyRequests
			}

			options.Hooks.Quota(r.Context(), tenant.QuotaDecision{
				TenantID:          tenantID,
				Allowed:           allowed,
				Tokens:            tokens,
				Requests:          1,
				RemainingTokens:   result.RemainingTokens,
				RemainingRequests: result.RemainingRequests,
				RetryAfter:        result.RetryAfter,
				Status:            status,
				Err:               err,
			})

			// Always surface remaining quota to the client.
			setQuotaHeaders(w, result.RemainingRequests, result.RemainingTokens)

			if allowed {
				next.ServeHTTP(w, r)
				return
			}

			setRetryAfterHeader(w, result.RetryAfter)

			if options.OnRejected != nil {
				options.OnRejected(w, r, result)
				return
			}

			writeTenantError(w, r, status, tenantCodeQuotaExceeded, "tenant quota exceeded", contract.CategoryRateLimit)
		})
	}
}
