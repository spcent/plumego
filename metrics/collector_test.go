package metrics

import (
	"context"
	"testing"
	"time"
)

func TestBaseMetricsCollectorMaxRecordsDefault(t *testing.T) {
	collector := NewBaseMetricsCollector()
	total := defaultMaxRecords + 5

	for i := 0; i < total; i++ {
		collector.Record(context.Background(), MetricRecord{
			Type:  MetricHTTPRequest,
			Name:  "http_request",
			Value: float64(i),
		})
	}

	records := collector.GetRecords()
	if len(records) != defaultMaxRecords {
		t.Fatalf("expected %d records, got %d", defaultMaxRecords, len(records))
	}
	if records[0].Value != 5 {
		t.Fatalf("expected oldest record to be 5, got %v", records[0].Value)
	}
	if records[len(records)-1].Value != float64(total-1) {
		t.Fatalf("expected newest record to be %d, got %v", total-1, records[len(records)-1].Value)
	}
}

func TestBaseMetricsCollectorMaxRecordsDisabled(t *testing.T) {
	collector := NewBaseMetricsCollector().WithMaxRecords(0)
	total := 250

	for i := 0; i < total; i++ {
		collector.Record(context.Background(), MetricRecord{
			Type:  MetricHTTPRequest,
			Name:  "http_request",
			Value: float64(i),
		})
	}

	records := collector.GetRecords()
	if len(records) != total {
		t.Fatalf("expected %d records, got %d", total, len(records))
	}
}

func TestBaseMetricsCollectorRecordClonesLabels(t *testing.T) {
	collector := NewBaseMetricsCollector()
	labels := MetricLabels{"route": "/users"}

	collector.Record(context.Background(), MetricRecord{
		Type:   MetricHTTPRequest,
		Name:   "custom_http_metric",
		Value:  1,
		Labels: labels,
	})

	labels["route"] = "/mutated"

	records := collector.GetRecords()
	if got := records[0].Labels["route"]; got != "/users" {
		t.Fatalf("expected stored labels to be cloned, got %q", got)
	}
}

func TestBaseMetricsCollectorObserveHTTP(t *testing.T) {
	collector := NewBaseMetricsCollector()
	duration := 125 * time.Millisecond

	collector.ObserveHTTP(context.Background(), "GET", "/test", 200, 100, duration)

	records := collector.GetRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.Type != MetricHTTPRequest {
		t.Fatalf("expected type %q, got %q", MetricHTTPRequest, record.Type)
	}
	if record.Name != "http_request" {
		t.Fatalf("expected name http_request, got %q", record.Name)
	}
	if record.Value != duration.Seconds() {
		t.Fatalf("expected value %f, got %f", duration.Seconds(), record.Value)
	}
	if record.Labels[labelMethod] != "GET" {
		t.Fatalf("expected method label GET, got %q", record.Labels[labelMethod])
	}
	if record.Labels[labelPath] != "/test" {
		t.Fatalf("expected path label /test, got %q", record.Labels[labelPath])
	}
	if record.Labels[labelStatus] != "200" {
		t.Fatalf("expected status label 200, got %q", record.Labels[labelStatus])
	}

	stats := collector.GetStats()
	if stats.TotalRecords != 1 {
		t.Fatalf("expected 1 total record, got %d", stats.TotalRecords)
	}
	if stats.TypeBreakdown[MetricHTTPRequest] != 1 {
		t.Fatalf("expected HTTP type breakdown of 1, got %d", stats.TypeBreakdown[MetricHTTPRequest])
	}
}

func TestBaseMetricsCollectorClear(t *testing.T) {
	collector := NewBaseMetricsCollector()
	collector.ObserveHTTP(context.Background(), "GET", "/test", 200, 100, 50*time.Millisecond)

	if len(collector.GetRecords()) != 1 {
		t.Fatalf("expected one record before clear")
	}

	collector.Clear()

	if len(collector.GetRecords()) != 0 {
		t.Fatalf("expected no records after clear")
	}
	stats := collector.GetStats()
	if stats.TotalRecords != 0 {
		t.Fatalf("expected zero total records after clear, got %d", stats.TotalRecords)
	}
}
