package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/pubsub"
)

// TestHandler serves demo and integration-test endpoints.
type TestHandler struct {
	bus *pubsub.InProcPubSub
}

// NewTestHandler creates a TestHandler backed by the given pub/sub bus.
func NewTestHandler(bus *pubsub.InProcPubSub) *TestHandler {
	return &TestHandler{bus: bus}
}

// PubSub publishes a test message to the requested topic and confirms delivery.
func (h *TestHandler) PubSub(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if topic == "" {
		topic = "test.default"
	}

	msg := pubsub.Message{
		Topic: topic,
		Type:  "test",
		Data:  fmt.Sprintf("Test message published at %s", time.Now().Format(time.RFC3339)),
		Time:  time.Now(),
	}

	if err := h.bus.Publish(topic, msg); err != nil {
		log.Printf("pubsub publish error: %v", err)
		contract.WriteError(w, r, contract.NewInternalError("publish failed"))
		return
	}

	if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"status":    "success",
		"topic":     topic,
		"message":   fmt.Sprint(msg.Data),
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// Webhook echoes metadata about the incoming request body.
func (h *TestHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	const maxBodyBytes = int64(1 << 20) // 1 MiB

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		log.Printf("webhook read error: %v", err)
		contract.WriteError(w, r, contract.APIError{
			Status:   http.StatusBadRequest,
			Code:     "read_failed",
			Message:  "failed to read request body",
			Category: contract.CategoryClient,
		})
		return
	}
	if int64(len(body)) > maxBodyBytes {
		contract.WriteError(w, r, contract.APIError{
			Status:   http.StatusRequestEntityTooLarge,
			Code:     "body_too_large",
			Message:  "request body exceeds limit",
			Category: contract.CategoryClient,
		})
		return
	}

	if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"status":         "webhook_received",
		"content_length": len(body),
		"timestamp":      time.Now().Format(time.RFC3339),
		"headers": map[string]string{
			"user_agent":   r.UserAgent(),
			"content_type": r.Header.Get("Content-Type"),
		},
	}, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
