package messaging

import (
	"net/http"

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
		status := http.StatusUnprocessableEntity
		code := "VALIDATION_ERROR"
		if isProviderError(err) {
			status = http.StatusBadGateway
			code = "PROVIDER_ERROR"
		}
		ctx.ErrorJSON(status, code, err.Error(), nil)
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

func isProviderError(err error) bool {
	if err == nil {
		return false
	}
	return errorIs(err, ErrProviderFailure)
}

func errorIs(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
