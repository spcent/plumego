package healthhttp

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/health"
)

func TestHealthHistoryExportHandler(t *testing.T) {
	manager, checker := newHistoryManager(t)

	manager.CheckAllComponents(context.Background())

	req := httptest.NewRequest("GET", "/health/history/export?format=json", nil)
	rr := httptest.NewRecorder()
	HealthHistoryExportHandler(manager).ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected content type application/json, got %s", contentType)
	}

	var result HealthHistoryQueryResult
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}

	checker.healthy = false
	manager.CheckAllComponents(context.Background())

	req = httptest.NewRequest("GET", "/health/history/export?format=csv", nil)
	rr = httptest.NewRecorder()
	HealthHistoryExportHandler(manager).ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/csv" {
		t.Fatalf("expected content type text/csv, got %s", contentType)
	}

	reader := csv.NewReader(strings.NewReader(rr.Body.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 CSV records (header + 2 data rows), got %d", len(records))
	}
	if len(records[0]) != 5 {
		t.Fatalf("expected 5 CSV columns, got %d", len(records[0]))
	}

	req = httptest.NewRequest("GET", "/health/history/export?format=invalid", nil)
	rr = httptest.NewRecorder()
	HealthHistoryExportHandler(manager).ServeHTTP(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected status 400 for invalid format, got %d", rr.Code)
	}
}

func TestHealthHistoryStatsHandler(t *testing.T) {
	manager, checker := newHistoryManager(t)

	manager.CheckAllComponents(context.Background())
	manager.CheckAllComponents(context.Background())
	checker.healthy = false
	manager.CheckAllComponents(context.Background())

	req := httptest.NewRequest("GET", "/health/history/stats", nil)
	rr := httptest.NewRecorder()
	HealthHistoryStatsHandler(manager).ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected content type application/json, got %s", contentType)
	}

	var stats map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &stats); err != nil {
		t.Fatalf("failed to unmarshal stats response: %v", err)
	}

	if total, ok := stats["total_entries"]; !ok || total != float64(3) {
		t.Fatalf("expected total_entries to be 3, got %v", total)
	}
}

type mutableChecker struct {
	name    string
	healthy bool
}

func (c *mutableChecker) Name() string { return c.name }

func (c *mutableChecker) Check(context.Context) error {
	if c.healthy {
		return nil
	}
	return errUnhealthy
}

var errUnhealthy = errors.New("unhealthy")

func newHistoryManager(t *testing.T) (*Tracker, *mutableChecker) {
	t.Helper()

	coreManager, err := health.NewHealthManager(health.HealthCheckConfig{})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	manager := NewTracker(coreManager)

	checker := &mutableChecker{name: "history", healthy: true}
	if err := manager.RegisterComponent(checker); err != nil {
		t.Fatalf("register component: %v", err)
	}

	return manager, checker
}
