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
	if !strings.Contains(body, "plumego_test_http_requests_total") {
		t.Fatalf("expected counter metric in response: %s", body)
	}
	if !strings.Contains(body, "plumego_test_http_request_duration_seconds_sum") {
		t.Fatalf("expected latency metric in response: %s", body)
	}
}
