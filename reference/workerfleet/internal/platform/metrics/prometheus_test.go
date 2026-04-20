package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrometheusTextExportsGaugeCounterAndHistogram(t *testing.T) {
	collector := NewCollector()
	if err := collector.SetGauge(MetricWorkersTotal, map[string]string{
		LabelNamespace: "sim",
		LabelNode:      "node-a",
		LabelStatus:    "online",
	}, 2); err != nil {
		t.Fatalf("set gauge: %v", err)
	}
	if err := collector.AddCounter(MetricCaseStartedTotal, map[string]string{
		LabelNamespace: "sim",
		LabelNode:      "node-a",
		LabelTaskType:  "simulation",
	}, 3); err != nil {
		t.Fatalf("add counter: %v", err)
	}
	if err := collector.ObserveHistogram(MetricCasePhaseDurationSeconds, map[string]string{
		LabelNamespace: "sim",
		LabelNode:      "node-a",
		LabelTaskType:  "simulation",
		LabelPhase:     "running",
	}, 0.2); err != nil {
		t.Fatalf("observe histogram: %v", err)
	}

	text := collector.PrometheusText()
	for _, want := range []string{
		"# TYPE workerfleet_workers gauge",
		`workerfleet_workers{namespace="sim",node="node-a",status="online"} 2.000000000`,
		"# TYPE workerfleet_case_started_total counter",
		`workerfleet_case_started_total{namespace="sim",node="node-a",task_type="simulation"} 3.000000000`,
		`workerfleet_case_phase_duration_seconds_bucket{namespace="sim",node="node-a",phase="running",task_type="simulation",le="+Inf"} 1`,
		`workerfleet_case_phase_duration_seconds_count{namespace="sim",node="node-a",phase="running",task_type="simulation"} 1`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q\n%s", want, text)
		}
	}
}

func TestCollectorRejectsForbiddenLabels(t *testing.T) {
	collector := NewCollector()
	if err := collector.SetGauge(MetricWorkersTotal, map[string]string{"worker_id": "worker-1"}, 1); err == nil {
		t.Fatalf("SetGauge with worker_id succeeded, want error")
	}
}

func TestHandlerServesPrometheusText(t *testing.T) {
	collector := NewCollector()
	if err := collector.SetGauge(MetricWorkersTotal, map[string]string{LabelStatus: "online"}, 1); err != nil {
		t.Fatalf("set gauge: %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	collector.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/plain; version=0.0.4" {
		t.Fatalf("content-type = %q", got)
	}
	if !strings.Contains(rec.Body.String(), "workerfleet_workers") {
		t.Fatalf("body missing metric: %s", rec.Body.String())
	}
}
