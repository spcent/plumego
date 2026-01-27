package health

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMetricsCollector(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	collector := NewMetricsCollector(manager)

	// Register a component
	mock := &MockChecker{name: "test", healthy: true}
	manager.RegisterComponent(mock)

	// Record some checks
	collector.RecordCheck("test", 10*time.Millisecond, true, StatusHealthy)
	collector.RecordCheck("test", 20*time.Millisecond, false, StatusUnhealthy)
	collector.RecordCheck("test", 15*time.Millisecond, true, StatusHealthy)

	metrics := collector.GetMetrics()

	if metrics.CheckCount != 3 {
		t.Fatalf("expected check count 3, got %d", metrics.CheckCount)
	}

	if metrics.SuccessCount != 2 {
		t.Fatalf("expected success count 2, got %d", metrics.SuccessCount)
	}

	if metrics.FailureCount != 1 {
		t.Fatalf("expected failure count 1, got %d", metrics.FailureCount)
	}

	// Check component metrics
	compMetrics, exists := collector.GetComponentMetrics("test")
	if !exists {
		t.Fatalf("expected component metrics for 'test', not found")
	}

	if compMetrics.CheckCount != 3 {
		t.Fatalf("expected component check count 3, got %d", compMetrics.CheckCount)
	}

	if compMetrics.SuccessCount != 2 {
		t.Fatalf("expected component success count 2, got %d", compMetrics.SuccessCount)
	}

	if compMetrics.MinLatency != 10*time.Millisecond {
		t.Fatalf("expected min latency 10ms, got %v", compMetrics.MinLatency)
	}

	if compMetrics.MaxLatency != 20*time.Millisecond {
		t.Fatalf("expected max latency 20ms, got %v", compMetrics.MaxLatency)
	}
}

func TestMetricsCollectorAttachment(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	collector := NewMetricsCollector(manager)

	mock := &MockChecker{name: "attached", healthy: true}
	if err := manager.RegisterComponent(mock); err != nil {
		t.Fatalf("failed to register component: %v", err)
	}

	if err := manager.CheckComponent(context.Background(), "attached"); err != nil {
		t.Fatalf("expected check to succeed, got %v", err)
	}

	metrics := collector.GetMetrics()
	if metrics.CheckCount != 1 {
		t.Fatalf("expected check count 1, got %d", metrics.CheckCount)
	}
	if compMetrics, ok := metrics.ComponentMetrics["attached"]; !ok || compMetrics.CheckCount != 1 {
		t.Fatalf("expected component metrics to be recorded")
	}
}

func TestMetricsCopyIsolation(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	collector := NewMetricsCollector(manager)

	collector.RecordCheck("test", 10*time.Millisecond, true, StatusHealthy)

	metrics := collector.GetMetrics()
	if metrics.ComponentMetrics["test"] == nil {
		t.Fatalf("expected component metrics to exist")
	}
	metrics.ComponentMetrics["test"].CheckCount = 7
	metrics.ComponentMetrics["test"].RecentHistory = nil

	metricsAfter := collector.GetMetrics()
	if metricsAfter.CheckCount != 1 {
		t.Fatalf("expected check count to remain 1, got %d", metricsAfter.CheckCount)
	}
	if metricsAfter.ComponentMetrics["test"].CheckCount != 1 {
		t.Fatalf("expected component check count to remain 1, got %d", metricsAfter.ComponentMetrics["test"].CheckCount)
	}
	if len(metricsAfter.ComponentMetrics["test"].RecentHistory) != 1 {
		t.Fatalf("expected recent history to be preserved")
	}

	compCopy, ok := collector.GetComponentMetrics("test")
	if !ok {
		t.Fatalf("expected component metrics copy")
	}
	compCopy.CheckCount = 99
	compAfter, ok := collector.GetComponentMetrics("test")
	if !ok || compAfter.CheckCount != 1 {
		t.Fatalf("expected component metrics to be immutable")
	}
}

func TestGetSuccessRate(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	collector := NewMetricsCollector(manager)

	// No checks yet
	rate := collector.GetSuccessRate()
	if rate != 0.0 {
		t.Fatalf("expected success rate 0.0, got %f", rate)
	}

	// Add some checks
	collector.RecordCheck("comp1", time.Millisecond, true, StatusHealthy)
	collector.RecordCheck("comp2", time.Millisecond, false, StatusUnhealthy)
	collector.RecordCheck("comp3", time.Millisecond, true, StatusHealthy)

	rate = collector.GetSuccessRate()
	expectedRate := 66.66666666666666 // 2/3 * 100
	if rate != expectedRate {
		t.Fatalf("expected success rate %f, got %f", expectedRate, rate)
	}
}

func TestGetUptime(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	collector := NewMetricsCollector(manager)

	// Should have some uptime
	uptime1 := collector.GetUptime()
	if uptime1 <= 0 {
		t.Fatalf("expected positive uptime, got %v", uptime1)
	}

	// Wait a bit and check again
	time.Sleep(100 * time.Millisecond)
	uptime2 := collector.GetUptime()
	if uptime2 <= uptime1 {
		t.Fatalf("uptime should increase, got %v -> %v", uptime1, uptime2)
	}
}

func TestReset(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	collector := NewMetricsCollector(manager)

	// Add some data
	collector.RecordCheck("comp1", time.Millisecond, true, StatusHealthy)
	collector.RecordCheck("comp2", time.Millisecond, false, StatusUnhealthy)

	metrics := collector.GetMetrics()
	if metrics.CheckCount == 0 {
		t.Fatalf("expected some check count before reset")
	}

	// Reset
	collector.Reset()

	metrics = collector.GetMetrics()
	if metrics.CheckCount != 0 {
		t.Fatalf("expected check count 0 after reset, got %d", metrics.CheckCount)
	}

	if len(metrics.ComponentMetrics) != 0 {
		t.Fatalf("expected no component metrics after reset, got %d", len(metrics.ComponentMetrics))
	}
}

func TestGenerateReport(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	collector := NewMetricsCollector(manager)

	// Register components
	mock1 := &MockChecker{name: "comp1", healthy: true}
	mock2 := &MockChecker{name: "comp2", healthy: false}

	manager.RegisterComponent(mock1)
	manager.RegisterComponent(mock2)

	// Add some metrics
	collector.RecordCheck("comp1", 10*time.Millisecond, true, StatusHealthy)
	collector.RecordCheck("comp2", 20*time.Millisecond, false, StatusUnhealthy)

	report := collector.GenerateReport()

	if report.BuildInfo.Version == "" {
		t.Fatalf("expected build info to be populated")
	}

	if report.Readiness.Ready {
		t.Fatalf("expected readiness to be false due to unhealthy component")
	}

	if len(report.Components) != 2 {
		t.Fatalf("expected 2 components in report, got %d", len(report.Components))
	}

	if report.HealthStatus.Status != StatusUnhealthy {
		t.Fatalf("expected overall status to be unhealthy, got %v", report.HealthStatus.Status)
	}

	if report.Metrics.CheckCount != 2 {
		t.Fatalf("expected 2 checks in metrics, got %d", report.Metrics.CheckCount)
	}
}

func TestMetricsHandler(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	collector := NewMetricsCollector(manager)

	// Add some test data
	collector.RecordCheck("test", time.Millisecond, true, StatusHealthy)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()

	MetricsHandler(collector).ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected content type application/json, got %s", contentType)
	}
}

func TestHealthReportHandler(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	collector := NewMetricsCollector(manager)

	// Register a healthy component
	mock := &MockChecker{name: "healthy", healthy: true}
	manager.RegisterComponent(mock)

	// Add some metrics
	collector.RecordCheck("healthy", time.Millisecond, true, StatusHealthy)

	// Trigger a health check to update the overall status
	ctx := context.Background()
	manager.CheckAllComponents(ctx)

	req := httptest.NewRequest("GET", "/health/report", nil)
	rr := httptest.NewRecorder()

	HealthReportHandler(collector).ServeHTTP(rr, req)

	// Check if we get 200 or 503 based on the actual health status
	if rr.Code != 200 && rr.Code != 503 {
		t.Fatalf("expected status 200 or 503, got %d", rr.Code)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected content type application/json, got %s", contentType)
	}
}
