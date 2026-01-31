package scheduler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAdminHandlerEndpoints(t *testing.T) {
	s := New(WithWorkers(1))
	s.Start()
	defer func() { _ = s.Stop(context.Background()) }()

	_, err := s.Delay("admin-1", 10*time.Millisecond, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	h := NewAdminHandler(s)

	req := httptest.NewRequest(http.MethodGet, "/scheduler/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status: %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/scheduler/jobs/admin-1", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("job status: %d", rec.Code)
	}

	var status JobStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if status.ID != "admin-1" {
		t.Fatalf("expected job id admin-1, got %s", status.ID)
	}

	req = httptest.NewRequest(http.MethodPost, "/scheduler/jobs/admin-1/pause", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("pause status: %d", rec.Code)
	}
}
