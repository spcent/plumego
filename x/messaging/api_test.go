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
	}{
		{
			name:       "provider error",
			err:        ErrProviderFailure,
			wantStatus: http.StatusBadGateway,
			wantCode:   "PROVIDER_ERROR",
		},
		{
			name:       "quota exceeded",
			err:        ErrQuotaExceeded,
			wantStatus: http.StatusTooManyRequests,
			wantCode:   "QUOTA_EXCEEDED",
		},
		{
			name:       "duplicate task",
			err:        mq.ErrDuplicateTask,
			wantStatus: http.StatusConflict,
			wantCode:   "DUPLICATE_MESSAGE",
		},
		{
			name:       "task expired",
			err:        mq.ErrTaskExpired,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "TASK_EXPIRED",
		},
		{
			name:       "validation",
			err:        ErrMissingID,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "mq invalid config",
			err:        mq.ErrInvalidConfig,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "not initialized",
			err:        mq.ErrNotInitialized,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "SERVICE_UNAVAILABLE",
		},
		{
			name:       "context timeout",
			err:        context.DeadlineExceeded,
			wantStatus: http.StatusGatewayTimeout,
			wantCode:   "TIMEOUT",
		},
		{
			name:       "unknown",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "SEND_ERROR",
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
		})
	}
}
