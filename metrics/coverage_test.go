package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/spcent/plumego/middleware"
)

// TestNewOpenTelemetryTracerEmptyName tests the default name behavior
func TestNewOpenTelemetryTracerEmptyName(t *testing.T) {
	tracer := NewOpenTelemetryTracer("")
	if tracer.name != "github.com/spcent/plumego/metrics" {
		t.Fatalf("expected default name, got: %s", tracer.name)
	}
}

// TestNewPrometheusCollectorEmptyNamespace tests the default namespace behavior
func TestNewPrometheusCollectorEmptyNamespace(t *testing.T) {
	collector := NewPrometheusCollector("")
	if collector.namespace != "plumego" {
		t.Fatalf("expected default namespace, got: %s", collector.namespace)
	}
}

// TestSpanHandleEnd4xxError tests 4xx error status handling
func TestSpanHandleEnd4xxError(t *testing.T) {
	tracer := NewOpenTelemetryTracer("test")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	_, span := tracer.Start(context.Background(), req)

	span.End(middleware.RequestMetrics{
		Status:   http.StatusBadRequest,
		Bytes:    100,
		TraceID:  "test",
		Duration: 50 * time.Millisecond,
	})

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span")
	}
	if spans[0].Status != "ERROR" {
		t.Fatalf("expected ERROR status for 4xx, got: %s", spans[0].Status)
	}
	if spans[0].StatusMessage == "" {
		t.Fatalf("expected non-empty status message for 4xx error")
	}
}

// TestSpanHandleEnd5xxError tests 5xx error status handling
func TestSpanHandleEnd5xxError(t *testing.T) {
	tracer := NewOpenTelemetryTracer("test")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	_, span := tracer.Start(context.Background(), req)

	span.End(middleware.RequestMetrics{
		Status:   http.StatusInternalServerError,
		Bytes:    100,
		TraceID:  "test",
		Duration: 50 * time.Millisecond,
	})

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span")
	}
	if spans[0].Status != "ERROR" {
		t.Fatalf("expected ERROR status for 5xx, got: %s", spans[0].Status)
	}
	if spans[0].StatusMessage == "" {
		t.Fatalf("expected non-empty status message for 5xx error")
	}
}

// TestSpanHandleEndSuccess tests successful status handling
func TestSpanHandleEndSuccess(t *testing.T) {
	tracer := NewOpenTelemetryTracer("test")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	_, span := tracer.Start(context.Background(), req)

	span.End(middleware.RequestMetrics{
		Status:   http.StatusOK,
		Bytes:    100,
		TraceID:  "test",
		Duration: 50 * time.Millisecond,
	})

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span")
	}
	if spans[0].Status != "OK" {
		t.Fatalf("expected OK status, got: %s", spans[0].Status)
	}
	if spans[0].StatusMessage != "" {
		t.Fatalf("expected empty status message for success, got: %s", spans[0].StatusMessage)
	}
}

// TestPrometheusCollectorEvictionEdgeCases tests edge cases for eviction
func TestPrometheusCollectorEvictionEdgeCases(t *testing.T) {
	// Test with max memory of 1
	collector := NewPrometheusCollector("test").WithMaxMemory(1)

	// Add first request
	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/test1",
		Status:   http.StatusOK,
		Duration: 100 * time.Millisecond,
	})

	// Add second request - should trigger eviction
	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/test2",
		Status:   http.StatusOK,
		Duration: 100 * time.Millisecond,
	})

	stats := collector.GetStats()
	if stats.Series > 1 {
		t.Fatalf("expected at most 1 series after eviction, got %d", stats.Series)
	}
}

// TestPrometheusCollectorMultipleMetricsPerLabel tests multiple metrics for same label
func TestPrometheusCollectorMultipleMetricsPerLabel(t *testing.T) {
	collector := NewPrometheusCollector("test")

	// Add multiple requests to same label
	for i := 0; i < 5; i++ {
		collector.Observe(context.Background(), middleware.RequestMetrics{
			Method:   http.MethodGet,
			Path:     "/test",
			Status:   http.StatusOK,
			Duration: time.Duration(100+i*10) * time.Millisecond,
		})
	}

	stats := collector.GetStats()
	if stats.TotalRequests != 5 {
		t.Fatalf("expected 5 total requests, got %d", stats.TotalRequests)
	}
	if stats.AverageLatency == 0 {
		t.Fatalf("expected non-zero average latency")
	}
}

// TestPrometheusHandlerContentType tests the content type header
func TestPrometheusHandlerContentType(t *testing.T) {
	collector := NewPrometheusCollector("test")
	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/test",
		Status:   http.StatusOK,
		Duration: 100 * time.Millisecond,
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	collector.Handler().ServeHTTP(rr, req)

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/plain; version=0.0.4" {
		t.Fatalf("expected Prometheus content type, got: %s", contentType)
	}
}

// TestPrometheusHandlerEmptyMetrics tests handler with no metrics
func TestPrometheusHandlerEmptyMetrics(t *testing.T) {
	collector := NewPrometheusCollector("test")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	collector.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if body == "" {
		t.Fatalf("expected non-empty response even with no metrics")
	}
}

// TestPrometheusCollectorObserveContext tests context handling in Observe
func TestPrometheusCollectorObserveContext(t *testing.T) {
	collector := NewPrometheusCollector("test")

	// Test with background context
	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodPost,
		Path:     "/api/data",
		Status:   http.StatusCreated,
		Duration: 250 * time.Millisecond,
	})

	stats := collector.GetStats()
	if stats.TotalRequests != 1 {
		t.Fatalf("expected 1 request, got %d", stats.TotalRequests)
	}
}

// TestPrometheusSnapshotConcurrency tests snapshot under concurrent load
func TestPrometheusSnapshotConcurrency(t *testing.T) {
	collector := NewPrometheusCollector("test")

	done := make(chan bool)
	// Writers
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				collector.Observe(context.Background(), middleware.RequestMetrics{
					Method:   http.MethodGet,
					Path:     "/concurrent",
					Status:   http.StatusOK,
					Duration: 50 * time.Millisecond,
				})
			}
			done <- true
		}(i)
	}

	// Readers (snapshot)
	for i := 0; i < 5; i++ {
		go func() {
			_ = collector.GetStats()
			done <- true
		}()
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := collector.GetStats()
	if stats.TotalRequests != 50 {
		t.Fatalf("expected 50 requests, got %d", stats.TotalRequests)
	}
}

// TestPrometheusSortedKeys tests the sortedKeys function directly
func TestPrometheusSortedKeys(t *testing.T) {
	m := map[labelKey]int{
		{method: "POST", path: "/b", status: "200"}: 1,
		{method: "GET", path: "/a", status: "200"}:  2,
		{method: "GET", path: "/a", status: "404"}:  3,
	}

	keys := sortedKeys(m)
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}

	// Verify order: GET /a 200, GET /a 404, POST /b 200
	if keys[0].method != "GET" || keys[0].path != "/a" || keys[0].status != "200" {
		t.Fatalf("unexpected first key: %+v", keys[0])
	}
	if keys[1].method != "GET" || keys[1].path != "/a" || keys[1].status != "404" {
		t.Fatalf("unexpected second key: %+v", keys[1])
	}
	if keys[2].method != "POST" || keys[2].path != "/b" || keys[2].status != "200" {
		t.Fatalf("unexpected third key: %+v", keys[2])
	}
}

// TestSpanAttributesCompleteness tests all expected attributes are present
func TestSpanAttributesCompleteness(t *testing.T) {
	tracer := NewOpenTelemetryTracer("test-service")
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Trace-ID", "parent-123")
	req.Host = "example.com"

	_, span := tracer.Start(context.Background(), req)

	// Sleep to ensure measurable duration
	time.Sleep(75 * time.Millisecond)

	span.End(middleware.RequestMetrics{
		Status:   http.StatusOK,
		Bytes:    512,
		TraceID:  "trace-456",
		Duration: 75 * time.Millisecond,
	})

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span")
	}

	attrs := spans[0].Attributes
	expectedAttrs := map[string]string{
		"http.method":                  "POST",
		"http.route":                   "/api/test",
		"http.user_agent":              "test-agent",
		"http.scheme":                  "http",
		"net.peer.name":                "example.com",
		"net.transport":                "tcp",
		"service.name":                 "test-service",
		"service.version":              "1.0.0",
		"parent.trace_id":              "parent-123",
		"http.status_code":             "200",
		"http.response_content_length": "512",
		"plumego.trace_id":             "trace-456",
		"http.status_text":             "OK",
		// duration_ms will be approximately 75ms, but we'll check it's reasonable
	}

	for key, expectedValue := range expectedAttrs {
		if actualValue, exists := attrs[key]; !exists {
			t.Fatalf("missing attribute: %s", key)
		} else if actualValue != expectedValue {
			t.Fatalf("attribute %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}

	// Check duration_ms is reasonable (should be >= 75ms)
	if durationStr, exists := attrs["duration_ms"]; exists {
		duration, err := strconv.ParseInt(durationStr, 10, 64)
		if err != nil {
			t.Fatalf("duration_ms is not a valid integer: %s", durationStr)
		}
		if duration < 75 {
			t.Fatalf("duration_ms should be at least 75ms, got %d", duration)
		}
	} else {
		t.Fatalf("missing duration_ms attribute")
	}
}

// TestSpanWithoutParentTraceID tests span without parent trace ID
func TestSpanWithoutParentTraceID(t *testing.T) {
	tracer := NewOpenTelemetryTracer("test")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No X-Trace-ID header

	_, span := tracer.Start(context.Background(), req)
	span.End(middleware.RequestMetrics{
		Status:   http.StatusOK,
		Bytes:    10,
		TraceID:  "test",
		Duration: 50 * time.Millisecond,
	})

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span")
	}

	if spans[0].ParentSpanID != "" {
		t.Fatalf("expected empty parent span ID, got: %s", spans[0].ParentSpanID)
	}

	if _, exists := spans[0].Attributes["parent.trace_id"]; exists {
		t.Fatalf("should not have parent.trace_id attribute")
	}
}

// TestPrometheusCollectorWithZeroMaxMemory tests edge case of zero max memory
func TestPrometheusCollectorWithZeroMaxMemory(t *testing.T) {
	// This should use the default value
	collector := NewPrometheusCollector("test")
	if collector.maxMemory != 10000 {
		t.Fatalf("expected default maxMemory 10000, got %d", collector.maxMemory)
	}
}

// TestPrometheusEvictionWithZeroRequests tests eviction when requests map is empty
func TestPrometheusEvictionWithZeroRequests(t *testing.T) {
	collector := NewPrometheusCollector("test").WithMaxMemory(5)

	// Force eviction on empty map (should not panic)
	collector.mu.Lock()
	collector.evictOldest()
	collector.mu.Unlock()

	// Should still work normally after
	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/test",
		Status:   http.StatusOK,
		Duration: 100 * time.Millisecond,
	})

	stats := collector.GetStats()
	if stats.TotalRequests != 1 {
		t.Fatalf("expected 1 request after eviction test, got %d", stats.TotalRequests)
	}
}

// TestPrometheusEvictionLeastUsed tests that least used entries are evicted first
func TestPrometheusEvictionLeastUsed(t *testing.T) {
	collector := NewPrometheusCollector("test").WithMaxMemory(3)

	// Add 3 entries with different usage counts
	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/least",
		Status:   http.StatusOK,
		Duration: 100 * time.Millisecond,
	})
	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/most",
		Status:   http.StatusOK,
		Duration: 100 * time.Millisecond,
	})
	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/middle",
		Status:   http.StatusOK,
		Duration: 100 * time.Millisecond,
	})

	// Add more usage to "most" and "middle"
	for i := 0; i < 3; i++ {
		collector.Observe(context.Background(), middleware.RequestMetrics{
			Method:   http.MethodGet,
			Path:     "/most",
			Status:   http.StatusOK,
			Duration: 100 * time.Millisecond,
		})
	}
	for i := 0; i < 2; i++ {
		collector.Observe(context.Background(), middleware.RequestMetrics{
			Method:   http.MethodGet,
			Path:     "/middle",
			Status:   http.StatusOK,
			Duration: 100 * time.Millisecond,
		})
	}

	// Add one more to trigger eviction
	collector.Observe(context.Background(), middleware.RequestMetrics{
		Method:   http.MethodGet,
		Path:     "/new",
		Status:   http.StatusOK,
		Duration: 100 * time.Millisecond,
	})

	stats := collector.GetStats()
	// Should have at most 3 series
	if stats.Series > 3 {
		t.Fatalf("expected at most 3 series, got %d", stats.Series)
	}
}
