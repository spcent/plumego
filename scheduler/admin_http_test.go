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

	block := make(chan struct{})
	_, err = s.Delay("admin-running", 0, func(ctx context.Context) error {
		<-block
		return nil
	})
	if err != nil {
		t.Fatalf("delay running: %v", err)
	}
	if !waitForAdminState(s, "admin-running", JobStateRunning, 500*time.Millisecond) {
		t.Fatalf("expected admin-running to reach running state")
	}

	req = httptest.NewRequest(http.MethodGet, "/scheduler/jobs?state=running", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("jobs state status: %d", rec.Code)
	}
	var jobs []JobStatus
	if err := json.NewDecoder(rec.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode jobs: %v", err)
	}
	found := false
	for _, job := range jobs {
		if job.ID == "admin-running" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected admin-running to be in running jobs list")
	}
	close(block)
}

func waitForAdminState(s *Scheduler, id JobID, state JobState, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, ok := s.Status(id)
		if ok && status.State == state {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}
