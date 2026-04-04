// Package handler contains the HTTP handlers for the with-messaging demo.
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/x/messaging"
)

// MessagingHandler handles pub/sub event endpoints.
// The Broker dependency is injected explicitly via the struct field.
type MessagingHandler struct {
	Broker *messaging.Broker
}

// Publish accepts a topic and payload and publishes it to the in-process broker.
func (h MessagingHandler) Publish(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Topic   string `json:"topic"`
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		_ = contract.WriteError(w, r, contract.NewBadRequestError("invalid JSON body"))
		return
	}
	if body.Topic == "" {
		_ = contract.WriteError(w, r, contract.NewBadRequestError("topic is required"))
		return
	}

	msg := messaging.BrokerMessage{
		Topic: body.Topic,
		Type:  "manual.publish",
		Time:  time.Now(),
		Data:  body.Payload,
	}
	if err := h.Broker.Publish(body.Topic, msg); err != nil {
		_ = contract.WriteError(w, r, contract.NewInternalError("failed to publish event"))
		return
	}

	if err := contract.WriteResponse(w, r, http.StatusAccepted, map[string]any{
		"ok":        true,
		"topic":     body.Topic,
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
