package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// discardLogger returns a StructuredLogger that silently discards all output.
// Use it in tests that construct handlers to satisfy the non-nil Logger requirement.
func discardLogger() plumelog.StructuredLogger {
	return plumelog.NewLogger(plumelog.LoggerConfig{Format: plumelog.LoggerFormatDiscard})
}

func TestHealthHandlerLive(t *testing.T) {
	h := NewHealthHandler(nil, discardLogger())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	h.Live(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestHealthHandlerReadyReportsUnavailable(t *testing.T) {
	h := NewHealthHandler(func(context.Context) error {
		return errors.New("store unavailable")
	}, discardLogger())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	h.Ready(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}

	var payload struct {
		Error struct {
			Code     string                 `json:"code"`
			Message  string                 `json:"message"`
			Category contract.ErrorCategory `json:"category"`
			Type     contract.ErrorType     `json:"type,omitempty"`
			Details  map[string]any         `json:"details,omitempty"`
		} `json:"error"`
		RequestID string `json:"request_id,omitempty"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if payload.Error.Code != contract.CodeUnavailable {
		t.Fatalf("error code = %q, want %q", payload.Error.Code, contract.CodeUnavailable)
	}
	if payload.Error.Message != "workerfleet readiness check failed" {
		t.Fatalf("error message = %q", payload.Error.Message)
	}
	if strings.Contains(payload.Error.Message, "store unavailable") {
		t.Fatalf("readiness response exposes raw error text: %q", payload.Error.Message)
	}
}
