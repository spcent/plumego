package message

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/store/idempotency"
)

// ExampleSendHandler demonstrates wiring idempotency + state transitions in a handler.
// This is a reference implementation and should be adapted to your persistence layer.
func ExampleSendHandler(idem idempotency.Store, repo Repository) contract.CtxHandlerFunc {
	return func(ctx *contract.Ctx) {
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
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Status:    StatusAccepted,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		hooks := []TransitionHook{&AuditHook{repo: repo}}

		if err := ApplyTransition(ctx.R.Context(), &msg, StatusQueued, hooks...); err != nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("transition failed"))
			return
		}

		if err := repo.Insert(ctx.R.Context(), msg); err != nil {
			contract.WriteError(ctx.W, ctx.R, contract.NewInternalError("persist failed"))
			return
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

// Repository is the minimal persistence interface needed by the handler example.
type Repository interface {
	Insert(ctx context.Context, msg Message) error
	UpdateStatus(ctx context.Context, id string, from Status, to Status, reason Reason) error
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
