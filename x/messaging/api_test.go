package messaging

import (
	"context"
	"errors"
	"net/http"
	"testing"

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
