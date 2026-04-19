package middleware

import (
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
var WriteTransportError = transport.WriteTransportError
