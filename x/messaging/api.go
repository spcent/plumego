package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/x/mq"
)

// HandleSend is the HTTP handler for POST /messages/send.
func (s *Service) HandleSend(w http.ResponseWriter, r *http.Request) {
	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Category(contract.CategoryValidation).
			Code(contract.CodeInvalidJSON).
			Message("invalid request body").
			Build())
		return
	}
	if err := contract.ValidateStruct(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).
			Code(contract.CodeInvalidRequest).
			Message(err.Error()).
			Build())
		return
	}
	if err := s.Send(r.Context(), req); err != nil {
		writeServiceError(w, r, err)
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusAccepted, map[string]string{
		"id":     req.ID,
		"status": "queued",
	}, nil)
}

// HandleBatchSend is the HTTP handler for POST /messages/batch.
func (s *Service) HandleBatchSend(w http.ResponseWriter, r *http.Request) {
	var batch BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Category(contract.CategoryValidation).
			Code(contract.CodeInvalidJSON).
			Message("invalid request body").
			Build())
		return
	}
	if err := contract.ValidateStruct(&batch); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).
			Code(contract.CodeInvalidRequest).
			Message(err.Error()).
			Build())
		return
	}
	if len(batch.Requests) == 0 {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).
			Code("EMPTY_BATCH").
			Message("requests array is empty").
			Build())
		return
	}
	result := s.SendBatch(r.Context(), batch)
	_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
}

// HandleStats is the HTTP handler for GET /messages/stats.
func (s *Service) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.Stats(r.Context())
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("STATS_ERROR").
			Message(err.Error()).
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, stats, nil)
}

// HandleGetReceipt is the HTTP handler for GET /messages/:id/receipt.
func (s *Service) HandleGetReceipt(w http.ResponseWriter, r *http.Request) {
	id := contract.RequestContextFromContext(r.Context()).Params["id"]
	if id == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Code("MISSING_ID").
			Message("message id is required").
			Build())
		return
	}
	receipt, found := s.receipts.Get(id)
	if !found {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code("NOT_FOUND").
			Message("receipt not found").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, receipt, nil)
}

// HandleListReceipts is the HTTP handler for GET /messages/receipts.
func (s *Service) HandleListReceipts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
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
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"receipts": receipts,
		"count":    len(receipts),
	}, nil)
}

// HandleChannelHealth is the HTTP handler for GET /messages/channels.
func (s *Service) HandleChannelHealth(w http.ResponseWriter, r *http.Request) {
	statuses := s.monitor.Status()
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"channels": statuses,
	}, nil)
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	_ = contract.WriteError(w, r, classifyServiceError(err))
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
			Type(contract.TypeRateLimited).
			Code("QUOTA_EXCEEDED").
			Message(err.Error()).
			Build()
	case errors.Is(err, mq.ErrDuplicateTask):
		return contract.NewErrorBuilder().
			Type(contract.TypeConflict).
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
			Type(contract.TypeInternal).
			Code(contract.CodeUnavailable).
			Message(err.Error()).
			Build()
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return contract.NewErrorBuilder().
			Type(contract.TypeTimeout).
			Status(http.StatusGatewayTimeout).
			Code("REQUEST_TIMEOUT").
			Message(err.Error()).
			Build()
	case isValidationError(err), errors.Is(err, mq.ErrInvalidConfig):
		return contract.NewErrorBuilder().
			Status(http.StatusUnprocessableEntity).
			Category(contract.CategoryValidation).
			Code(contract.CodeValidationError).
			Message(err.Error()).
			Build()
	default:
		return contract.NewErrorBuilder().
			Type(contract.TypeInternal).
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
