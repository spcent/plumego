package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/x/observability"
)

func discardLogger() plumelog.StructuredLogger {
	return plumelog.NewLogger(plumelog.LoggerConfig{Format: plumelog.LoggerFormatDiscard})
}

// --- APIHandler ---

func TestAPIHandlerRootReturnsServiceIdentity(t *testing.T) {
	h := APIHandler{Logger: discardLogger(), ServiceName: "test-svc", Version: "v1.2.3"}
	rec := httptest.NewRecorder()
	h.Root(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var env struct {
		Data struct {
			Service string `json:"service"`
			Version string `json:"version"`
			Docs    string `json:"docs"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if env.Data.Service != "test-svc" {
		t.Errorf("service = %q, want %q", env.Data.Service, "test-svc")
	}
	if env.Data.Version != "v1.2.3" {
		t.Errorf("version = %q, want %q", env.Data.Version, "v1.2.3")
	}
	if env.Data.Docs == "" {
		t.Error("docs field is empty")
	}
}

func TestAPIHandlerHelloReturnsMetadata(t *testing.T) {
	h := APIHandler{Logger: discardLogger(), ServiceName: "obs-svc", Version: "dev"}
	rec := httptest.NewRecorder()
	h.Hello(rec, httptest.NewRequest(http.MethodGet, "/api/hello", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != contract.ContentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", ct, contract.ContentTypeJSON)
	}
	var env struct {
		Data struct {
			Message   string `json:"message"`
			Service   string `json:"service"`
			Timestamp string `json:"timestamp"`
			Version   string `json:"version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !strings.Contains(env.Data.Message, "with-observability") {
		t.Errorf("message = %q, want to contain 'with-observability'", env.Data.Message)
	}
	if env.Data.Service != "obs-svc" {
		t.Errorf("service = %q, want %q", env.Data.Service, "obs-svc")
	}
	if env.Data.Timestamp == "" {
		t.Error("timestamp is empty")
	}
}

// --- ObservabilityHandler ---

func TestObservabilityHandlerStatsReturnsJSON(t *testing.T) {
	collector := observability.NewPrometheusCollector("test")
	h := ObservabilityHandler{
		Logger:    discardLogger(),
		Collector: collector,
		Tracer:    observability.NewOpenTelemetryTracer("test"),
	}
	rec := httptest.NewRecorder()
	h.Stats(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != contract.ContentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", ct, contract.ContentTypeJSON)
	}
	var env struct {
		Data struct {
			TotalRequests int64 `json:"total_requests"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode body: %v", err)
	}
}

func TestObservabilityHandlerSpansReturnsJSON(t *testing.T) {
	tracer := observability.NewOpenTelemetryTracer("test")
	h := ObservabilityHandler{
		Logger:    discardLogger(),
		Collector: observability.NewPrometheusCollector("test"),
		Tracer:    tracer,
	}
	rec := httptest.NewRecorder()
	h.Spans(rec, httptest.NewRequest(http.MethodGet, "/api/v1/spans", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var env struct {
		Data struct {
			Total int           `json:"total"`
			Spans []interface{} `json:"spans"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode body: %v", err)
	}
}

// --- MetricsHandler ---

func TestMetricsHandlerCollectorStatsReturnsJSON(t *testing.T) {
	collector := observability.NewPrometheusCollector("test")
	h := MetricsHandler{
		Logger:   discardLogger(),
		Observer: collector,
	}
	rec := httptest.NewRecorder()
	h.CollectorStats(rec, httptest.NewRequest(http.MethodGet, "/api/v1/collector-stats", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != contract.ContentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", ct, contract.ContentTypeJSON)
	}
}

// --- HealthHandler ---

func TestHealthHandlerLiveReturnsOK(t *testing.T) {
	h := HealthHandler{
		ServiceName: "test-svc",
		Logger:      discardLogger(),
		Checkers:    []health.ComponentChecker{},
	}
	rec := httptest.NewRecorder()
	h.Live(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != contract.ContentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", ct, contract.ContentTypeJSON)
	}
	var env struct {
		Data struct {
			Status    string `json:"status"`
			Service   string `json:"service"`
			Timestamp string `json:"timestamp"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if env.Data.Status != "ok" {
		t.Errorf("status = %q, want %q", env.Data.Status, "ok")
	}
	if env.Data.Service != "test-svc" {
		t.Errorf("service = %q, want %q", env.Data.Service, "test-svc")
	}
	if env.Data.Timestamp == "" {
		t.Error("timestamp is empty")
	}
}

func TestHealthHandlerReadyWithNoCheckersReturnsOK(t *testing.T) {
	h := HealthHandler{
		ServiceName: "test-svc",
		Logger:      discardLogger(),
		Checkers:    []health.ComponentChecker{},
	}
	rec := httptest.NewRecorder()
	h.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var env struct {
		Data struct {
			Ready      bool            `json:"ready"`
			Components map[string]bool `json:"components"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !env.Data.Ready {
		t.Errorf("ready = %v, want true", env.Data.Ready)
	}
}

// --- parsePositiveInt ---

func TestParsePositiveInt(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"5", 5},
		{"100", 100},
		{"0", 0},
		{"-1", 0},
		{"abc", 0},
		{"", 0},
	}
	for _, c := range cases {
		if got := parsePositiveInt(c.in); got != c.want {
			t.Errorf("parsePositiveInt(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
