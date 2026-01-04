package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/middleware"
)

func TestPrometheusCollectorObserveAndHandler(t *testing.T) {
	collector := NewPrometheusCollector("plumego_test")

	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/test",
		Status:   http.StatusOK,
		Duration: 125 * time.Millisecond,
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	collector.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status from metrics handler: %d", rr.Code)
	}
	body := rr.Body.String()

	// Check for counter metric
	if !strings.Contains(body, "plumego_test_http_requests_total") {
		t.Fatalf("expected counter metric in response: %s", body)
	}

	// Check for duration sum metric
	if !strings.Contains(body, "plumego_test_http_request_duration_seconds_sum") {
		t.Fatalf("expected duration sum metric in response: %s", body)
	}

	// Check for duration count metric
	if !strings.Contains(body, "plumego_test_http_request_duration_seconds_count") {
		t.Fatalf("expected duration count metric in response: %s", body)
	}

	// Check for min and max metrics
	if !strings.Contains(body, "plumego_test_http_request_duration_seconds_min") {
		t.Fatalf("expected min metric in response: %s", body)
	}
	if !strings.Contains(body, "plumego_test_http_request_duration_seconds_max") {
		t.Fatalf("expected max metric in response: %s", body)
	}

	// Check for uptime metric
	if !strings.Contains(body, "plumego_test_uptime_seconds") {
		t.Fatalf("expected uptime metric in response: %s", body)
	}

	// Check for total requests metric
	if !strings.Contains(body, "plumego_test_http_requests_total_all") {
		t.Fatalf("expected total requests metric in response: %s", body)
	}
}

func TestPrometheusCollectorMultipleRequests(t *testing.T) {
	collector := NewPrometheusCollector("plumego_test")

	// Add multiple requests with different labels
	requests := []struct {
		method string
		path   string
		status int
	}{
		{http.MethodGet, "/api/users", http.StatusOK},
		{http.MethodPost, "/api/users", http.StatusCreated},
		{http.MethodGet, "/api/posts", http.StatusOK},
		{http.MethodGet, "/api/users", http.StatusInternalServerError},
	}

	for _, req := range requests {
		collector.Observe(context.Background(), middleware.RequestMetrics{
			Method:   req.method,
			Path:     req.path,
			Status:   req.status,
			Duration: 100 * time.Millisecond,
		})
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	collector.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Check that all expected metrics are present
	expectedMetrics := []string{
		`plumego_test_http_requests_total{method="GET",path="/api/users",status="200"} 1`,
		`plumego_test_http_requests_total{method="POST",path="/api/users",status="201"} 1`,
		`plumego_test_http_requests_total{method="GET",path="/api/posts",status="200"} 1`,
		`plumego_test_http_requests_total{method="GET",path="/api/users",status="500"} 1`,
	}

	for _, expected := range expectedMetrics {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected metric not found: %s\nBody: %s", expected, body)
		}
	}
}

func TestPrometheusCollectorStats(t *testing.T) {
	collector := NewPrometheusCollector("plumego_test")

	// Add some requests
	for i := 0; i < 5; i++ {
		collector.Observe(context.Background(), middleware.RequestMetrics{
			Method:   http.MethodGet,
			Path:     "/test",
			Status:   http.StatusOK,
			Duration: time.Duration(100+i*10) * time.Millisecond,
		})
	}

	stats := collector.GetStats()

	if stats.Series != 1 {
		t.Fatalf("expected 1 series, got %d", stats.Series)
	}
	if stats.TotalRequests != 5 {
		t.Fatalf("expected 5 total requests, got %d", stats.TotalRequests)
	}
	if stats.AverageLatency == 0 {
		t.Fatalf("expected non-zero average latency")
	}
	if stats.StartTime.IsZero() {
		t.Fatalf("expected non-zero start time")
	}
}

func TestPrometheusCollectorClear(t *testing.T) {
	collector := NewPrometheusCollector("plumego_test")

	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/test",
		Status:   http.StatusOK,
		Duration: 100 * time.Millisecond,
	})

	stats := collector.GetStats()
	if stats.TotalRequests != 1 {
		t.Fatalf("expected 1 request before clear")
	}

	collector.Clear()
	stats = collector.GetStats()
	if stats.TotalRequests != 0 {
		t.Fatalf("expected 0 requests after clear")
	}
}

func TestPrometheusCollectorMaxMemory(t *testing.T) {
	collector := NewPrometheusCollector("plumego_test").WithMaxMemory(3)

	// Add more requests than max memory
	for i := 0; i < 5; i++ {
		collector.Observe(context.Background(), middleware.RequestMetrics{
			Method:   http.MethodGet,
			Path:     "/test" + string(rune('A'+i)),
			Status:   http.StatusOK,
			Duration: 100 * time.Millisecond,
		})
	}

	stats := collector.GetStats()
	if stats.Series > 3 {
		t.Fatalf("expected at most 3 series due to max memory limit, got %d", stats.Series)
	}
}

func TestPrometheusCollectorConcurrency(t *testing.T) {
	collector := NewPrometheusCollector("plumego_test")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			collector.Observe(context.Background(), middleware.RequestMetrics{
				Method:   http.MethodGet,
				Path:     "/concurrent",
				Status:   http.StatusOK,
				Duration: 50 * time.Millisecond,
			})
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	stats := collector.GetStats()
	if stats.TotalRequests != 10 {
		t.Fatalf("expected 10 requests, got %d", stats.TotalRequests)
	}
}

func TestPrometheusCollectorMetricsFormat(t *testing.T) {
	collector := NewPrometheusCollector("plumego_test")

	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodPost,
		Path:     "/api/data",
		Status:   http.StatusCreated,
		Duration: 250 * time.Millisecond,
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	collector.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()

	// Verify Prometheus format
	if !strings.Contains(body, "# HELP") {
		t.Fatalf("expected HELP comments in Prometheus format")
	}
	if !strings.Contains(body, "# TYPE") {
		t.Fatalf("expected TYPE comments in Prometheus format")
	}
	if !strings.Contains(body, "plumego_test_http_requests_total{method=\"POST\",path=\"/api/data\",status=\"201\"} 1") {
		t.Fatalf("expected properly formatted counter metric")
	}
}

func TestPrometheusCollectorEmptyNamespace(t *testing.T) {
	collector := NewPrometheusCollector("")

	if collector.namespace != "plumego" {
		t.Fatalf("expected default namespace 'plumego', got '%s'", collector.namespace)
	}
}

func TestPrometheusCollectorEviction(t *testing.T) {
	collector := NewPrometheusCollector("plumego_test").WithMaxMemory(5)

	// Fill up to limit
	for i := 0; i < 5; i++ {
		collector.Observe(context.Background(), middleware.RequestMetrics{
			Method:   http.MethodGet,
			Path:     "/fill" + string(rune('0'+i)),
			Status:   http.StatusOK,
			Duration: 100 * time.Millisecond,
		})
	}

	stats := collector.GetStats()
	if stats.Series != 5 {
		t.Fatalf("expected 5 series, got %d", stats.Series)
	}

	// Add one more, should trigger eviction
	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/new",
		Status:   http.StatusOK,
		Duration: 100 * time.Millisecond,
	})

	stats = collector.GetStats()
	// Should be at most 5 (max memory)
	if stats.Series > 5 {
		t.Fatalf("expected at most 5 series after eviction, got %d", stats.Series)
	}
}
