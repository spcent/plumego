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
		_ = contract.WriteError(w, r, contract.APIError{Status: http.StatusBadRequest, Code: "invalid_body", Message: "invalid JSON body", Category: contract.CategoryForStatus(http.StatusBadRequest)})
		return
	}
	if body.Topic == "" {
		_ = contract.WriteError(w, r, contract.APIError{Status: http.StatusBadRequest, Code: "missing_topic", Message: "topic is required", Category: contract.CategoryForStatus(http.StatusBadRequest)})
		return
	}

	msg := messaging.BrokerMessage{
		Topic: body.Topic,
		Type:  "manual.publish",
		Time:  time.Now(),
		Data:  body.Payload,
	}
	if err := h.Broker.Publish(body.Topic, msg); err != nil {
		_ = contract.WriteError(w, r, contract.APIError{Status: http.StatusInternalServerError, Code: "publish_failed", Message: "failed to publish event", Category: contract.CategoryForStatus(http.StatusInternalServerError)})
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
