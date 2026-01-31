package metrics

import (
	"context"
	"testing"
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
