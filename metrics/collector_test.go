package metrics

import (
	"errors"
	"testing"
	"time"
)

func TestBaseMetricsCollectorRecordTracksTotalsAndBreakdown(t *testing.T) {
	collector := NewBaseMetricsCollector()
	labels := MetricLabels{"route": "/users"}

	collector.Record(t.Context(), MetricRecord{
		Name:   "custom_http_metric",
		Value:  1,
		Labels: labels,
	})

	labels["route"] = "/mutated"

	stats := collector.GetStats()
	if stats.TotalRecords != 1 {
		t.Fatalf("expected one total record, got %d", stats.TotalRecords)
	}
	if stats.NameBreakdown["custom_http_metric"] != 1 {
		t.Fatalf("expected name breakdown to track custom metric")
	}

	collector.Record(t.Context(), MetricRecord{
		Name:   "custom_http_metric",
		Value:  1,
		Labels: labels,
	})

	stats = collector.GetStats()
	if stats.TotalRecords != 2 {
		t.Fatalf("expected two total records, got %d", stats.TotalRecords)
	}
	if stats.NameBreakdown["custom_http_metric"] != 2 {
		t.Fatalf("expected custom metric count 2, got %d", stats.NameBreakdown["custom_http_metric"])
	}
}

func TestBaseMetricsCollectorObserveHTTP(t *testing.T) {
	collector := NewBaseMetricsCollector()
	duration := 125 * time.Millisecond

	collector.ObserveHTTP(t.Context(), "GET", "/test", 200, 100, duration)

	stats := collector.GetStats()
	if stats.TotalRecords != 1 {
		t.Fatalf("expected 1 total record, got %d", stats.TotalRecords)
	}
	if stats.NameBreakdown[MetricHTTPRequest] != 1 {
		t.Fatalf("expected HTTP name breakdown of 1, got %d", stats.NameBreakdown[MetricHTTPRequest])
	}
}

func TestBaseMetricsCollectorClassifiesHTTPStatusErrors(t *testing.T) {
	collector := NewBaseMetricsCollector()

	collector.ObserveHTTP(t.Context(), "GET", "/missing", 404, 100, 50*time.Millisecond)
	collector.Record(t.Context(), MetricRecord{
		Name: MetricHTTPRequest,
		Labels: MetricLabels{
			labelMethod: "POST",
			labelPath:   "/submit",
			labelStatus: "500",
		},
	})
	collector.Record(t.Context(), MetricRecord{
		Name:  "owner_metric_error",
		Error: errors.New("boom"),
	})

	stats := collector.GetStats()
	if stats.TotalRecords != 3 {
		t.Fatalf("expected 3 total records, got %d", stats.TotalRecords)
	}
	if stats.ErrorRecords != 3 {
		t.Fatalf("expected HTTP and explicit errors to be classified, got %d", stats.ErrorRecords)
	}
}

func TestBaseMetricsCollectorIgnoresInvalidHTTPStatusForErrorClassification(t *testing.T) {
	collector := NewBaseMetricsCollector()

	collector.Record(t.Context(), MetricRecord{
		Name: MetricHTTPRequest,
		Labels: MetricLabels{
			labelMethod: "GET",
			labelPath:   "/broken",
			labelStatus: "not-a-status",
		},
	})

	stats := collector.GetStats()
	if stats.ErrorRecords != 0 {
		t.Fatalf("expected invalid HTTP status label to stay non-error, got %d", stats.ErrorRecords)
	}
}

func TestBaseMetricsCollectorClear(t *testing.T) {
	collector := NewBaseMetricsCollector()
	collector.ObserveHTTP(t.Context(), "GET", "/test", 200, 100, 50*time.Millisecond)

	collector.Clear()

	stats := collector.GetStats()
	if stats.TotalRecords != 0 {
		t.Fatalf("expected zero total records after clear, got %d", stats.TotalRecords)
	}
	if len(stats.NameBreakdown) != 0 {
		t.Fatalf("expected empty name breakdown after clear, got %#v", stats.NameBreakdown)
	}
}

func TestBaseMetricsCollectorStatsCloneNameBreakdown(t *testing.T) {
	collector := NewBaseMetricsCollector()
	collector.Record(t.Context(), MetricRecord{Name: "custom_metric"})

	stats := collector.GetStats()
	stats.NameBreakdown["custom_metric"] = 999

	stats = collector.GetStats()
	if stats.NameBreakdown["custom_metric"] != 1 {
		t.Fatalf("expected collector stats clone to protect internal breakdown, got %d", stats.NameBreakdown["custom_metric"])
	}
}
