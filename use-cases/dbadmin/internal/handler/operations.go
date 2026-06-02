package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// OperationsHandler exposes active non-SQL operations.
type OperationsHandler struct {
	Registry *OperationRegistry
	Logger   plumelog.StructuredLogger
}

// ListActive returns currently running non-SQL operations.
func (h OperationsHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	if h.Registry == nil {
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, []any{}, nil))
		return
	}

	ops := h.Registry.ListActive()
	result := make([]map[string]any, 0, len(ops))
	for _, op := range ops {
		result = append(result, map[string]any{
			"operationId": op.OperationID,
			"driver":      op.Driver,
			"kind":        op.Kind,
			"connId":      op.ConnID,
			"resource":    op.Resource,
			"summary":     op.Summary,
			"startTime":   op.StartTime,
			"duration":    time.Since(op.StartTime).String(),
		})
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, result, map[string]any{"count": len(result)}))
}

// Cancel cancels an active non-SQL operation.
func (h OperationsHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	if h.Registry == nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("operation registry not available").Build()))
		return
	}

	var req struct {
		OperationID string `json:"operationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}
	if req.OperationID == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("operationId is required").Build()))
		return
	}

	if h.Registry.Cancel(req.OperationID) {
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
			"status":      "cancelled",
			"operationId": req.OperationID,
		}, nil))
		return
	}

	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeNotFound).Message("operation not found or already completed").Build()))
}
