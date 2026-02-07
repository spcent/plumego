package messaging

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
)

// HandleSend is the HTTP handler for POST /messages/send.
func (s *Service) HandleSend(ctx *contract.Ctx) {
	var req SendRequest
	if err := ctx.BindAndValidateJSON(&req); err != nil {
		ctx.ErrorJSON(http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if err := s.Send(ctx.R.Context(), req); err != nil {
		writeServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusAccepted, map[string]string{
		"id":     req.ID,
		"status": "queued",
	})
}

// HandleBatchSend is the HTTP handler for POST /messages/batch.
func (s *Service) HandleBatchSend(ctx *contract.Ctx) {
	var batch BatchRequest
	if err := ctx.BindAndValidateJSON(&batch); err != nil {
		ctx.ErrorJSON(http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if len(batch.Requests) == 0 {
		ctx.ErrorJSON(http.StatusBadRequest, "EMPTY_BATCH", "requests array is empty", nil)
		return
	}
	result := s.SendBatch(ctx.R.Context(), batch)
	ctx.JSON(http.StatusOK, result)
}

// HandleStats is the HTTP handler for GET /messages/stats.
func (s *Service) HandleStats(ctx *contract.Ctx) {
	stats, err := s.Stats(ctx.R.Context())
	if err != nil {
		ctx.ErrorJSON(http.StatusInternalServerError, "STATS_ERROR", err.Error(), nil)
		return
	}
	ctx.JSON(http.StatusOK, stats)
}

// HandleGetReceipt is the HTTP handler for GET /messages/:id/receipt.
func (s *Service) HandleGetReceipt(ctx *contract.Ctx) {
	id, ok := ctx.Param("id")
	if !ok || id == "" {
		ctx.ErrorJSON(http.StatusBadRequest, "MISSING_ID", "message id is required", nil)
		return
	}
	receipt, found := s.receipts.Get(id)
	if !found {
		ctx.ErrorJSON(http.StatusNotFound, "NOT_FOUND", "receipt not found", nil)
		return
	}
	ctx.JSON(http.StatusOK, receipt)
}

// HandleListReceipts is the HTTP handler for GET /messages/receipts.
func (s *Service) HandleListReceipts(ctx *contract.Ctx) {
	q := ctx.R.URL.Query()
	filter := ReceiptFilter{
		Channel:  Channel(q.Get("channel")),
		Status:   q.Get("status"),
		TenantID: q.Get("tenant_id"),
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			filter.Offset = n
		}
	}
	receipts := s.receipts.List(filter)
	ctx.JSON(http.StatusOK, map[string]any{
		"receipts": receipts,
		"count":    len(receipts),
	})
}

// HandleChannelHealth is the HTTP handler for GET /messages/channels.
func (s *Service) HandleChannelHealth(ctx *contract.Ctx) {
	statuses := s.monitor.Status()
	ctx.JSON(http.StatusOK, map[string]any{
		"channels": statuses,
	})
}

func writeServiceError(ctx *contract.Ctx, err error) {
	status := http.StatusUnprocessableEntity
	code := "VALIDATION_ERROR"
	if errors.Is(err, ErrProviderFailure) {
		status = http.StatusBadGateway
		code = "PROVIDER_ERROR"
	} else if errors.Is(err, ErrQuotaExceeded) {
		status = http.StatusTooManyRequests
		code = "QUOTA_EXCEEDED"
	}
	ctx.ErrorJSON(status, code, err.Error(), nil)
}
