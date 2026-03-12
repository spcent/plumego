package transport

import (
	"net/http"
	"strconv"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
)

const (
	DefaultTenantHeader = "X-Tenant-ID"
	DefaultModelHeader  = "X-Model"
	DefaultToolHeader   = "X-Tool"
	DefaultTokensHeader = "X-Token-Count"

	CodeRequired      = middleware.CodeTenantRequired
	CodeInvalidID     = middleware.CodeTenantInvalidID
	CodeRateLimited   = middleware.CodeTenantRateLimited
	CodePolicyDenied  = middleware.CodeTenantPolicyDenied
	CodeQuotaExceeded = middleware.CodeTenantQuotaExceeded
)

func HeaderOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string, category contract.ErrorCategory) {
	middleware.WriteTransportError(w, r, status, code, message, category, nil)
}

func SetRetryAfterHeader(w http.ResponseWriter, retry time.Duration) {
	if retry <= 0 {
		return
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
}

func SetRateLimitHeaders(w http.ResponseWriter, limit, remaining int64) {
	if limit > 0 {
		w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
	}
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
}

// SetQuotaHeaders sets X-Quota-Remaining-Requests and X-Quota-Remaining-Tokens headers.
// Values < 0 indicate unlimited and are omitted.
func SetQuotaHeaders(w http.ResponseWriter, remainingRequests, remainingTokens int) {
	if remainingRequests >= 0 {
		w.Header().Set("X-Quota-Remaining-Requests", strconv.Itoa(remainingRequests))
	}
	if remainingTokens >= 0 {
		w.Header().Set("X-Quota-Remaining-Tokens", strconv.Itoa(remainingTokens))
	}
}
