package ops

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

type queueStatsResponse struct {
	Data struct {
		Queues []QueueStats `json:"queues"`
	} `json:"data"`
}

func TestOpsQueueStatsSingle(t *testing.T) {
	r := router.NewRouter()
	handler := New(Options{
		Enabled: true,
		Auth:    AuthConfig{AllowInsecure: true},
		Hooks: Hooks{
			QueueStats: func(ctx context.Context, queue string) (QueueStats, error) {
				return QueueStats{Queue: queue, Queued: 3}, nil
			},
		},
	})

	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/ops/queue?queue=primary", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp queueStatsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Data.Queues) != 1 {
		t.Fatalf("expected 1 queue, got %d", len(resp.Data.Queues))
	}
	if resp.Data.Queues[0].Queue != "primary" {
		t.Fatalf("expected queue primary, got %q", resp.Data.Queues[0].Queue)
	}
}

func TestOpsAuthRequired(t *testing.T) {
	r := router.NewRouter()
	handler := New(Options{
		Enabled: true,
		Hooks: Hooks{
			QueueStats: func(ctx context.Context, queue string) (QueueStats, error) {
				return QueueStats{Queue: queue}, nil
			},
		},
	})

	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/ops/queue?queue=primary", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestOpsNotConfiguredErrorsUseStableCodes(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		wantCode string
	}{
		{
			name:     "queue stats",
			method:   http.MethodGet,
			path:     "/ops/queue",
			wantCode: codeQueueStatsNotConfigured,
		},
		{
			name:     "queue replay",
			method:   http.MethodPost,
			path:     "/ops/queue/replay",
			wantCode: codeQueueReplayNotConfigured,
		},
		{
			name:     "receipt lookup",
			method:   http.MethodGet,
			path:     "/ops/receipts",
			wantCode: codeReceiptLookupNotConfigured,
		},
		{
			name:     "channel health",
			method:   http.MethodGet,
			path:     "/ops/channels",
			wantCode: codeChannelHealthNotConfigured,
		},
		{
			name:     "tenant quota",
			method:   http.MethodGet,
			path:     "/ops/tenants/quota",
			wantCode: codeTenantQuotaNotConfigured,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := router.NewRouter()
			handler := New(Options{
				Enabled: true,
				Auth:    AuthConfig{AllowInsecure: true},
			})
			handler.RegisterRoutes(r)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotImplemented {
				t.Fatalf("expected status 501, got %d: %s", rec.Code, rec.Body.String())
			}

			var resp contract.ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if resp.Error.Code != tt.wantCode {
				t.Fatalf("code=%s, want %s", resp.Error.Code, tt.wantCode)
			}
		})
	}
}
