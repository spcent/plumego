package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/metrics"
)

func TestConfigureObservabilityMetrics(t *testing.T) {
	app := New()

	prom := metrics.NewPrometheusCollector("plumego_test")
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.Path = "/metrics"
	cfg.Metrics.Collector = prom

	if err := app.ConfigureObservability(cfg); err != nil {
		t.Fatalf("configure observability: %v", err)
	}

	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	stats := prom.GetStats()
	if stats.TotalRequests == 0 {
		t.Fatalf("expected metrics to be recorded")
	}

	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from metrics endpoint, got %d", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("unexpected content type: %s", ct)
	}
}

func TestConfigureObservabilityTracing(t *testing.T) {
	app := New()

	cfg := DefaultObservabilityConfig()
	cfg.Tracing.Enabled = true

	if err := app.ConfigureObservability(cfg); err != nil {
		t.Fatalf("configure observability: %v", err)
	}

	if app.tracer == nil {
		t.Fatalf("expected tracer to be configured")
	}

	if !app.loggingEnabled {
		t.Fatalf("expected logging to be enabled")
	}
}
