package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/core"
	"workerfleet/internal/handler"
	workerfleetmetrics "workerfleet/internal/platform/metrics"
)

func TestRegisterRoutesWiresMetricsWhenProvided(t *testing.T) {
	app := core.New(core.DefaultConfig(), core.AppDependencies{})
	metrics := workerfleetmetrics.NewCollector()
	if err := metrics.SetGauge(workerfleetmetrics.MetricWorkersTotal, map[string]string{
		workerfleetmetrics.LabelStatus: "online",
	}, 1); err != nil {
		t.Fatalf("set gauge: %v", err)
	}

	if err := RegisterRoutes(app, handler.New(nil), metrics.Handler()); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "workerfleet_workers") {
		t.Fatalf("body missing metrics output: %s", rec.Body.String())
	}
}
