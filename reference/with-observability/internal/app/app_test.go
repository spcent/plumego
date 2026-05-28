package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
	"with-observability/internal/config"
)

func TestRegisterRoutesCanonicalShape(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	got := a.Core.Routes()
	want := []router.RouteInfo{
		{Method: http.MethodGet, Path: "/"},
		{Method: http.MethodGet, Path: "/api/hello"},
		{Method: http.MethodGet, Path: "/api/v1/collector-stats"},
		{Method: http.MethodGet, Path: "/api/v1/spans"},
		{Method: http.MethodGet, Path: "/api/v1/stats"},
		{Method: http.MethodGet, Path: "/healthz"},
		{Method: http.MethodGet, Path: "/metrics"},
		{Method: http.MethodGet, Path: "/readyz"},
	}
	if len(got) != len(want) {
		t.Fatalf("routes count = %d, want %d\ngot:  %#v\nwant: %#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("routes[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestMetricsEndpointReturnsPrometheusText(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare app: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("get server: %v", err)
	}

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics: status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("GET /metrics: Content-Type = %q, want text/plain", ct)
	}
	// Uptime gauge is always present.
	body := rec.Body.String()
	if !strings.Contains(body, "_uptime_seconds") {
		t.Fatalf("GET /metrics: body missing uptime gauge\ngot: %s", body)
	}
}

func TestSpansEndpointReturnsJSON(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare app: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("get server: %v", err)
	}

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/spans", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/spans: status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != contract.ContentTypeJSON {
		t.Fatalf("GET /api/v1/spans: Content-Type = %q, want %q", got, contract.ContentTypeJSON)
	}
}

func TestMetricsAccumulateAfterRequests(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare app: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("get server: %v", err)
	}

	// Send two requests so the collector has something to report.
	for range 2 {
		rec := httptest.NewRecorder()
		srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/hello", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/hello: status = %d, want %d", rec.Code, http.StatusOK)
		}
	}

	stats := a.Collector.GetStats()
	if stats.TotalRecords == 0 {
		t.Fatal("PrometheusCollector.GetStats: TotalRecords = 0, want > 0 after requests")
	}
}

func TestTracingMiddlewareRecordsSpans(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare app: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("get server: %v", err)
	}

	// One request through the tracing middleware should produce one span.
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/hello", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/hello: status = %d, want %d", rec.Code, http.StatusOK)
	}

	spans := a.Tracer.Spans()
	if len(spans) == 0 {
		t.Fatal("OpenTelemetryTracer.Spans: got 0 spans, want > 0 after request")
	}
	// Each span must carry the standard HTTP attributes.
	for _, s := range spans {
		if s.TraceID == "" {
			t.Errorf("span %q: TraceID is empty", s.Name)
		}
		if s.SpanID == "" {
			t.Errorf("span %q: SpanID is empty", s.Name)
		}
		if s.Attributes["http.method"] == "" {
			t.Errorf("span %q: missing http.method attribute", s.Name)
		}
	}
}

func TestRequestIDPropagationThroughMiddlewareStack(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare app: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("get server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	req.Header.Set(contract.RequestIDHeader, "obs-test-1")
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(contract.RequestIDHeader); got != "obs-test-1" {
		t.Fatalf("%s = %q, want %q", contract.RequestIDHeader, got, "obs-test-1")
	}
}
