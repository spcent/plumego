package message

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/metrics/smsgateway"
)

// ReceiptRequest represents a delivery receipt payload.
type ReceiptRequest struct {
	MessageID   string `json:"message_id" validate:"required"`
	DeliveredAt string `json:"delivered_at,omitempty"` // RFC3339, defaults to now if empty
}

// ExampleReceiptHandler demonstrates handling delivery receipts and emitting metrics.
func ExampleReceiptHandler(repo Store, reporter *smsgateway.Reporter) contract.CtxHandlerFunc {
	return func(ctx *contract.Ctx) {
		var req ReceiptRequest
		if err := ctx.BindAndValidateJSON(&req); err != nil {
			contract.WriteBindError(ctx.W, ctx.R, err)
			return
		}

		msg, found, err := repo.Get(ctx.R.Context(), req.MessageID)
		if err != nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("message lookup failed"))
			return
		}
		if !found {
			contract.WriteError(ctx.W, ctx.R, contract.NewNotFoundError("message"))
			return
		}

		deliveredAt := time.Now().UTC()
		if req.DeliveredAt != "" {
			parsed, err := time.Parse(time.RFC3339, req.DeliveredAt)
			if err != nil {
				contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("delivered_at", "must be RFC3339"))
				return
			}
			deliveredAt = parsed.UTC()
		}

		if msg.Status == StatusDelivered {
			ctx.W.WriteHeader(http.StatusOK)
			return
		}

		if msg.Status != StatusSent {
			contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("status", "message not in sent state"))
			return
		}

		sentAt := msg.SentAt
		if sentAt.IsZero() {
			sentAt = msg.UpdatedAt
		}
		if sentAt.IsZero() {
			sentAt = msg.CreatedAt
		}
		delay := deliveredAt.Sub(sentAt)
		if delay < 0 {
			delay = 0
		}

		if err := ApplyAndPersist(ctx.R.Context(), repo, &msg, StatusDelivered, Reason{}); err != nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("update failed"))
			return
		}

		if reporter != nil {
			reporter.RecordReceiptDelay(ctx.R.Context(), msg.TenantID, msg.Provider, delay)
			reporter.RecordStatus(ctx.R.Context(), msg.TenantID, string(StatusDelivered))
		}

		ctx.W.WriteHeader(http.StatusOK)
	}
}
