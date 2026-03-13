package healthhttp

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/health"
)

func TestMetricsHandler(t *testing.T) {
	config := health.HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := health.NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	tracker := health.NewMetricsTracker(manager)
	tracker.RecordCheck("test", time.Millisecond, true, health.StatusHealthy)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	MetricsHandler(tracker).ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected content type application/json, got %s", contentType)
	}
}

func TestHealthReportHandler(t *testing.T) {
	config := health.HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := health.NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	tracker := health.NewMetricsTracker(manager)

	_ = manager.RegisterComponent(&mockChecker{name: "healthy", healthy: true})
	tracker.RecordCheck("healthy", time.Millisecond, true, health.StatusHealthy)
	manager.CheckAllComponents(context.Background())

	req := httptest.NewRequest("GET", "/health/report", nil)
	rr := httptest.NewRecorder()
	HealthReportHandler(tracker).ServeHTTP(rr, req)

	if rr.Code != 200 && rr.Code != 503 {
		t.Fatalf("expected status 200 or 503, got %d", rr.Code)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected content type application/json, got %s", contentType)
	}
}
