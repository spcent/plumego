package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/x/mq"
)

func TestClassifyServiceError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantMsg    string
	}{
		{
			name:       "provider error",
			err:        ErrProviderFailure,
			wantStatus: http.StatusBadGateway,
			wantCode:   "PROVIDER_ERROR",
			wantMsg:    "provider error",
		},
		{
			name:       "quota exceeded",
			err:        ErrQuotaExceeded,
			wantStatus: http.StatusTooManyRequests,
			wantCode:   "QUOTA_EXCEEDED",
			wantMsg:    "quota exceeded",
		},
		{
			name:       "duplicate task",
			err:        mq.ErrDuplicateTask,
			wantStatus: http.StatusConflict,
			wantCode:   "DUPLICATE_MESSAGE",
			wantMsg:    "duplicate message",
		},
		{
			name:       "task expired",
			err:        mq.ErrTaskExpired,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "TASK_EXPIRED",
			wantMsg:    "task expired",
		},
		{
			name:       "validation",
			err:        ErrMissingID,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
			wantMsg:    "message validation failed",
		},
		{
			name:       "mq invalid config",
			err:        mq.ErrInvalidConfig,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
			wantMsg:    "message validation failed",
		},
		{
			name:       "not initialized",
			err:        mq.ErrNotInitialized,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "SERVICE_UNAVAILABLE",
			wantMsg:    "messaging service unavailable",
		},
		{
			name:       "context timeout",
			err:        context.DeadlineExceeded,
			wantStatus: http.StatusGatewayTimeout,
			wantCode:   "TIMEOUT",
			wantMsg:    "request timed out",
		},
		{
			name:       "unknown",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "SEND_ERROR",
			wantMsg:    "send failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := classifyServiceError(tt.err)
			if apiErr.Status != tt.wantStatus {
				t.Fatalf("status=%d, want %d", apiErr.Status, tt.wantStatus)
			}
			if apiErr.Code != tt.wantCode {
				t.Fatalf("code=%s, want %s", apiErr.Code, tt.wantCode)
			}
			if apiErr.Message != tt.wantMsg {
				t.Fatalf("message=%q, want %q", apiErr.Message, tt.wantMsg)
			}
		})
	}
}

func TestHandleSendMissingRequiredFieldsUsesSafeError(t *testing.T) {
	svc := New(Config{})
	req := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	svc.HandleSend(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}

	var resp contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != contract.CodeValidationError {
		t.Fatalf("code=%s, want %s", resp.Error.Code, contract.CodeValidationError)
	}
	if resp.Error.Message != "message validation failed" {
		t.Fatalf("message=%q, want %q", resp.Error.Message, "message validation failed")
	}
	if strings.Contains(resp.Error.Message, "messaging:") {
		t.Fatalf("message exposes raw error text: %q", resp.Error.Message)
	}
}

func TestHandleBatchSendEmptyRequestsUsesSafeError(t *testing.T) {
	svc := New(Config{})
	req := httptest.NewRequest(http.MethodPost, "/messages/batch", bytes.NewBufferString(`{"requests":[]}`))
	rec := httptest.NewRecorder()

	svc.HandleBatchSend(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != contract.CodeEmptyBatch {
		t.Fatalf("code=%s, want %s", resp.Error.Code, contract.CodeEmptyBatch)
	}
	if resp.Error.Message != "requests array is empty" {
		t.Fatalf("message=%q, want %q", resp.Error.Message, "requests array is empty")
	}
	if strings.Contains(resp.Error.Message, "messaging:") {
		t.Fatalf("message exposes raw error text: %q", resp.Error.Message)
	}
}
