package transport

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

// Middleware-specific transport error codes. Keep values stable for client integrations.
const (
	CodeServerBusy         = "server_busy"
	CodeServerQueueTimeout = "server_queue_timeout"
	CodeUpstreamFailed     = "upstream_failed"
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
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Status(status).
		Code(code).
		Message(message).
		Category(category).
		Details(details).
		Build())
}
