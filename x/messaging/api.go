package messaging

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/x/mq"
)

// HandleSend is the HTTP handler for POST /messages/send.
func (s *Service) HandleSend(ctx *contract.Ctx) {
	var req SendRequest
	if err := ctx.BindJSON(&req); err != nil {
		_ = contract.WriteBindError(ctx.W, ctx.R, err)
		return
	}
	if err := contract.ValidateStruct(&req); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.ErrTypeValidation).
			Code("INVALID_REQUEST").
			Message(err.Error()).
			Build())
		return
	}
	if err := s.Send(ctx.R.Context(), req); err != nil {
		writeServiceError(ctx, err)
		return
	}
	_ = ctx.Response(http.StatusAccepted, map[string]string{
		"id":     req.ID,
		"status": "queued",
	}, nil)
}

// HandleBatchSend is the HTTP handler for POST /messages/batch.
func (s *Service) HandleBatchSend(ctx *contract.Ctx) {
	var batch BatchRequest
	if err := ctx.BindJSON(&batch); err != nil {
		_ = contract.WriteBindError(ctx.W, ctx.R, err)
		return
	}
	if err := contract.ValidateStruct(&batch); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.ErrTypeValidation).
			Code("INVALID_REQUEST").
			Message(err.Error()).
			Build())
		return
	}
	if len(batch.Requests) == 0 {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.ErrTypeValidation).
			Code("EMPTY_BATCH").
			Message("requests array is empty").
			Build())
		return
	}
	result := s.SendBatch(ctx.R.Context(), batch)
	_ = ctx.Response(http.StatusOK, result, nil)
}

// HandleStats is the HTTP handler for GET /messages/stats.
func (s *Service) HandleStats(ctx *contract.Ctx) {
	stats, err := s.Stats(ctx.R.Context())
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.ErrTypeInternal).
			Code("STATS_ERROR").
			Message(err.Error()).
			Build())
		return
	}
	_ = ctx.Response(http.StatusOK, stats, nil)
}

// HandleGetReceipt is the HTTP handler for GET /messages/:id/receipt.
func (s *Service) HandleGetReceipt(ctx *contract.Ctx) {
	id, ok := ctx.Param("id")
	if !ok || id == "" {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.ErrTypeRequired).
			Code("MISSING_ID").
			Message("message id is required").
			Build())
		return
	}
	receipt, found := s.receipts.Get(id)
	if !found {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.ErrTypeNotFound).
			Code("NOT_FOUND").
			Message("receipt not found").
			Build())
		return
	}
	_ = ctx.Response(http.StatusOK, receipt, nil)
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
	_ = ctx.Response(http.StatusOK, map[string]any{
		"receipts": receipts,
		"count":    len(receipts),
	}, nil)
}

// HandleChannelHealth is the HTTP handler for GET /messages/channels.
func (s *Service) HandleChannelHealth(ctx *contract.Ctx) {
	statuses := s.monitor.Status()
	_ = ctx.Response(http.StatusOK, map[string]any{
		"channels": statuses,
	}, nil)
}

func writeServiceError(ctx *contract.Ctx, err error) {
	_ = contract.WriteError(ctx.W, ctx.R, classifyServiceError(err))
}

func classifyServiceError(err error) contract.APIError {
	switch {
	case errors.Is(err, ErrProviderFailure):
		return contract.NewErrorBuilder().
			Status(http.StatusBadGateway).
			Category(contract.CategoryServer).
			Code("PROVIDER_ERROR").
			Message(err.Error()).
			Build()
	case errors.Is(err, ErrQuotaExceeded):
		return contract.NewErrorBuilder().
			Type(contract.ErrTypeRateLimited).
			Code("QUOTA_EXCEEDED").
			Message(err.Error()).
			Build()
	case errors.Is(err, mq.ErrDuplicateTask):
		return contract.NewErrorBuilder().
			Type(contract.ErrTypeConflict).
			Code("DUPLICATE_MESSAGE").
			Message(err.Error()).
			Build()
	case errors.Is(err, mq.ErrTaskExpired):
		return contract.NewErrorBuilder().
			Status(http.StatusUnprocessableEntity).
			Category(contract.CategoryValidation).
			Code("TASK_EXPIRED").
			Message(err.Error()).
			Build()
	case errors.Is(err, mq.ErrNotInitialized):
		return contract.NewErrorBuilder().
			Type(contract.ErrTypeInternal).
			Code("SERVICE_UNAVAILABLE").
			Message(err.Error()).
			Build()
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return contract.NewErrorBuilder().
			Type(contract.ErrTypeTimeout).
			Status(http.StatusGatewayTimeout).
			Code("REQUEST_TIMEOUT").
			Message(err.Error()).
			Build()
	case isValidationError(err), errors.Is(err, mq.ErrInvalidConfig):
		return contract.NewErrorBuilder().
			Status(http.StatusUnprocessableEntity).
			Category(contract.CategoryValidation).
			Code("VALIDATION_ERROR").
			Message(err.Error()).
			Build()
	default:
		return contract.NewErrorBuilder().
			Type(contract.ErrTypeInternal).
			Code("SEND_ERROR").
			Message(err.Error()).
			Build()
	}
}

func isValidationError(err error) bool {
	return errors.Is(err, ErrMissingID) ||
		errors.Is(err, ErrInvalidChannel) ||
		errors.Is(err, ErrInvalidEmail) ||
		errors.Is(err, ErrInvalidPhone) ||
		errors.Is(err, ErrMissingBody) ||
		errors.Is(err, ErrMissingRecipient) ||
		errors.Is(err, ErrMissingSubject) ||
		errors.Is(err, ErrTemplateRender)
}
