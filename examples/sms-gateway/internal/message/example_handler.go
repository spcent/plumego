package message

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/examples/sms-gateway/internal/routing"
	"github.com/spcent/plumego/examples/sms-gateway/internal/tasks"
	"github.com/spcent/plumego/net/mq"
	"github.com/spcent/plumego/store/idempotency"
	"github.com/spcent/plumego/tenant"
)

// ExampleSendHandler demonstrates wiring idempotency + state transitions in a handler.
// This is a reference implementation and should be adapted to your persistence layer.
func ExampleSendHandler(idem idempotency.Store, repo Repository, router *routing.PolicyRouter, enqueuer TaskEnqueuer) contract.CtxHandlerFunc {
	return func(ctx *contract.Ctx) {
		tenantID := tenant.TenantIDFromContext(ctx.R.Context())
		if tenantID == "" {
			contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("tenant_id", "missing tenant id"))
			return
		}

		provider := ""
		if router != nil {
			selected, _, err := router.SelectProvider(ctx.R.Context(), tenantID)
			if err != nil {
				contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("route policy unavailable"))
				return
			}
			provider = selected
		}

		idempotencyKey := ctx.R.Header.Get("Idempotency-Key")
		if idempotencyKey == "" {
			contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("Idempotency-Key", "missing Idempotency-Key"))
			return
		}

		payload, err := io.ReadAll(ctx.R.Body)
		if err != nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("body", "invalid body"))
			return
		}

		var req SendRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("body", "invalid json payload"))
			return
		}
		if req.To == "" {
			contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("to", "missing recipient"))
			return
		}
		if req.Body == "" {
			contract.WriteError(ctx.W, ctx.R, contract.NewValidationError("body", "missing body"))
			return
		}

		maxAttempts := req.MaxAttempts
		if maxAttempts <= 0 {
			maxAttempts = mq.DefaultTaskMaxAttempts
		}

		hash := sha256.Sum256(payload)
		requestHash := hex.EncodeToString(hash[:])

		rec, found, err := idem.Get(ctx.R.Context(), idempotencyKey)
		if err != nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("idempotency store error"))
			return
		}
		if found {
			if rec.RequestHash != requestHash {
				contract.WriteError(ctx.W, ctx.R, idempotencyConflict())
				return
			}
			if rec.Status == idempotency.StatusCompleted {
				ctx.W.Header().Set("Content-Type", "application/json")
				ctx.W.WriteHeader(http.StatusOK)
				_, _ = ctx.W.Write(rec.Response)
				return
			}
		}

		created, err := idem.PutIfAbsent(ctx.R.Context(), idempotency.Record{
			Key:         idempotencyKey,
			RequestHash: requestHash,
			Status:      idempotency.StatusInProgress,
			ExpiresAt:   time.Now().Add(10 * time.Minute),
		})
		if err != nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("idempotency store error"))
			return
		}
		if !created {
			contract.WriteError(ctx.W, ctx.R, idempotencyConflict())
			return
		}

		msg := Message{
			ID:             fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			TenantID:       tenantID,
			Provider:       provider,
			Status:         StatusAccepted,
			MaxAttempts:    maxAttempts,
			IdempotencyKey: idempotencyKey,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		hooks := []TransitionHook{&AuditHook{repo: repo}}

		if err := repo.Insert(ctx.R.Context(), msg); err != nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("persist failed"))
			return
		}

		if err := ApplyTransition(ctx.R.Context(), &msg, StatusQueued, hooks...); err != nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("transition failed"))
			return
		}

		if provider != "" {
			if updater, ok := repo.(ProviderUpdater); ok {
				_ = updater.UpdateProvider(ctx.R.Context(), msg.ID, provider)
			}
		}

		if enqueuer != nil {
			taskPayload, err := tasks.EncodeSendTask(tasks.SendTaskPayload{
				MessageID: msg.ID,
				TenantID:  tenantID,
				To:        req.To,
				Body:      req.Body,
				Provider:  provider,
			})
			if err != nil {
				contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("queue payload error"))
				return
			}

			task := mq.Task{
				ID:        fmt.Sprintf("task-%d", time.Now().UnixNano()),
				Topic:     tasks.SendTopic,
				TenantID:  tenantID,
				Payload:   taskPayload,
				DedupeKey: msg.ID,
			}

			tasks.AttachTrace(ctx.R.Context(), &task)

			opts := mq.EnqueueOptions{}
			if req.MaxAttempts > 0 {
				opts.MaxAttempts = req.MaxAttempts
			}
			if err := enqueuer.Enqueue(ctx.R.Context(), task, opts); err != nil {
				contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("enqueue failed"))
				return
			}
		}

		response := []byte(`{"message_id":"` + msg.ID + `"}`)
		if err := idem.Complete(ctx.R.Context(), idempotencyKey, response); err != nil {
			// Idempotency completion failure should not block response.
			_ = err
		}

		ctx.W.Header().Set("Content-Type", "application/json")
		ctx.W.WriteHeader(http.StatusOK)
		_, _ = ctx.W.Write(response)
	}
}

// AuditHook is a simple example hook that persists transition events.
type AuditHook struct {
	repo Repository
}

func (h *AuditHook) Before(ctx context.Context, from Status, to Status, msg Message) error {
	return nil
}

func (h *AuditHook) After(ctx context.Context, from Status, to Status, msg Message) error {
	if h.repo == nil {
		return nil
	}
	return h.repo.UpdateStatus(ctx, msg.ID, from, to, Reason{Code: msg.ReasonCode, Detail: msg.ReasonDetail})
}

func (h *AuditHook) OnInvalid(ctx context.Context, from Status, to Status, msg Message, err error) {
	_ = err
}

func idempotencyConflict() contract.APIError {
	return contract.NewErrorBuilder().
		Status(http.StatusConflict).
		Category(contract.CategoryClient).
		Type(contract.ErrTypeConflict).
		Code("IDEMPOTENCY_CONFLICT").
		Message("idempotency key conflict").
		Build()
}

// SendRequest is the minimal payload expected by the handler example.
type SendRequest struct {
	To          string `json:"to"`
	Body        string `json:"body"`
	MaxAttempts int    `json:"max_attempts,omitempty"`
}
