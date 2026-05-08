package middleware

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware/internal/transport"
)

// Middleware transport error codes. Re-exported from middleware/internal/transport.
// Keep values stable for client integrations.
const (
	CodeServerBusy         = transport.CodeServerBusy
	CodeServerQueueTimeout = transport.CodeServerQueueTimeout
	CodeUpstreamFailed     = transport.CodeUpstreamFailed
)

// WriteTransportError writes middleware transport errors using the canonical contract helper.
// Re-exported from middleware/internal/transport for use by callers of the parent package.
func WriteTransportError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	code string,
	message string,
	category contract.ErrorCategory,
	details map[string]any,
) {
	transport.WriteTransportError(w, r, status, code, message, category, details)
}
