package tenantmw

import (
	"net/http"
	"strconv"
	"time"

	"github.com/spcent/plumego/contract"
	mw "github.com/spcent/plumego/middleware"
)

const (
	defaultModelHeader  = "X-Model"
	defaultToolHeader   = "X-Tool"
	defaultTokensHeader = "X-Token-Count"

	tenantCodePolicyDenied  = mw.CodeTenantPolicyDenied
	tenantCodeQuotaExceeded = mw.CodeTenantQuotaExceeded
)

func headerOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func writeTenantError(w http.ResponseWriter, r *http.Request, status int, code, message string, category contract.ErrorCategory) {
	mw.WriteTransportError(w, r, status, code, message, category, nil)
}

func setRetryAfterHeader(w http.ResponseWriter, retry time.Duration) {
	if retry <= 0 {
		return
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
}

// setQuotaHeaders sets X-Quota-Remaining-Requests and X-Quota-Remaining-Tokens headers.
// Values < 0 indicate unlimited and are omitted.
func setQuotaHeaders(w http.ResponseWriter, remainingRequests, remainingTokens int) {
	if remainingRequests >= 0 {
		w.Header().Set("X-Quota-Remaining-Requests", strconv.Itoa(remainingRequests))
	}
	if remainingTokens >= 0 {
		w.Header().Set("X-Quota-Remaining-Tokens", strconv.Itoa(remainingTokens))
	}
}
