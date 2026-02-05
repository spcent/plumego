package ops

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/router"
)

type queueStatsResponse struct {
	Data struct {
		Queues []QueueStats `json:"queues"`
	} `json:"data"`
}

func TestOpsQueueStatsSingle(t *testing.T) {
	r := router.NewRouter()
	comp := NewComponent(Options{
		Enabled: true,
		Auth:    AuthConfig{AllowInsecure: true},
		Hooks: Hooks{
			QueueStats: func(ctx context.Context, queue string) (QueueStats, error) {
				return QueueStats{Queue: queue, Queued: 3}, nil
			},
		},
	})

	comp.RegisterRoutes(r)

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
	comp := NewComponent(Options{
		Enabled: true,
		Hooks: Hooks{
			QueueStats: func(ctx context.Context, queue string) (QueueStats, error) {
				return QueueStats{Queue: queue}, nil
			},
		},
	})

	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/ops/queue?queue=primary", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
