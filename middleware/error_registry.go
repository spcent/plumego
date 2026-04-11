package middleware

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

// Middleware transport error codes. Keep values stable for client integrations.
// Shared error codes live in contract; only middleware-specific codes remain here.
const (
	CodeTenantRequired      = "tenant_required"
	CodeTenantInvalidID     = "tenant_invalid_id"
	CodeTenantPolicyDenied  = "tenant_policy_denied"
	CodeTenantQuotaExceeded = "tenant_quota_exceeded"
	CodeTenantRateLimited   = "tenant_rate_limited"
	CodeServerBusy          = "server_busy"
	CodeServerQueueTimeout  = "server_queue_timeout"
	CodeUpstreamFailed      = "upstream_failed"
)

// WriteTransportError writes middleware transport errors using the canonical contract helper.
func WriteTransportError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	code string,
	message string,
	category contract.ErrorCategory,
	details map[string]any,
) {
	contract.WriteError(w, r, contract.NewErrorBuilder().
		Status(status).
		Code(code).
		Message(message).
		Category(category).
		Details(details).
		Build())
}
