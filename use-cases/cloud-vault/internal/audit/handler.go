package audit

import (
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// Handler exposes the audit log via HTTP.
type Handler struct {
	logger *Logger
	log    plumelog.StructuredLogger
}

// NewHandler creates an audit Handler.
func NewHandler(logger *Logger, log plumelog.StructuredLogger) *Handler {
	return &Handler{logger: logger, log: log}
}

// List handles GET /api/v1/audit
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resourceType := q.Get("resource_type")
	resourceID := q.Get("resource_id")
	limit := parseIntParam(q.Get("limit"), 50, 200)
	offset := parseIntParam(q.Get("offset"), 0, -1)

	events, total, err := h.logger.List(r.Context(), resourceType, resourceID, limit, offset)
	if err != nil {
		h.log.Error("audit list", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.log, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("internal server error").
			Build()))
		return
	}

	// Return empty slice rather than null when no events exist.
	if events == nil {
		events = []Event{}
	}

	logWriteErr(h.log, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"items":  events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}, nil))
}

func parseIntParam(s string, def, max int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	if max > 0 && n > max {
		return max
	}
	return n
}

func logWriteErr(log plumelog.StructuredLogger, err error) {
	if err != nil && log != nil {
		log.Error("write response", plumelog.Fields{"error": err.Error()})
	}
}
