package tenant

import (
	"net/http"
	"strconv"
	"time"

	"github.com/spcent/plumego/contract"
)

const (
	defaultTenantHeader = "X-Tenant-ID"
	defaultModelHeader  = "X-Model"
	defaultToolHeader   = "X-Tool"
	defaultTokensHeader = "X-Token-Count"
)

func headerOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func writeTenantError(w http.ResponseWriter, r *http.Request, status int, code, message string, category contract.ErrorCategory) {
	contract.WriteError(w, r, contract.APIError{
		Status:   status,
		Code:     code,
		Message:  message,
		Category: category,
	})
}

func setRetryAfterHeader(w http.ResponseWriter, retry time.Duration) {
	if retry <= 0 {
		return
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
}

func setRateLimitHeaders(w http.ResponseWriter, limit, remaining int64) {
	if limit > 0 {
		w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
	}
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
}
