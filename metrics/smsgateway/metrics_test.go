package smsgateway

import (
	"context"
	"testing"
	"time"

	"github.com/spcent/plumego/metrics"
)

func TestReporterRecords(t *testing.T) {
	collector := metrics.NewMockCollector()
	reporter := NewReporter(collector)

	ctx := context.Background()
	reporter.RecordQueueDepth(ctx, "send", QueueStats{Queued: 2, Leased: 1, Dead: 0, Expired: 0})
	reporter.RecordSendLatency(ctx, "tenant-1", "provider-a", 120*time.Millisecond, nil)
	reporter.RecordProviderResult(ctx, "tenant-1", "provider-a", true)
	reporter.RecordRetry(ctx, "tenant-1", "provider-a", 2)
	reporter.RecordStatus(ctx, "tenant-1", "sent")

	records := collector.GetRecords()
	if len(records) == 0 {
		t.Fatalf("expected records")
	}

	found := false
	for _, rec := range records {
		if rec.Type == metrics.MetricSMSGateway && rec.Name == MetricSendLatency {
			found = true
			if rec.Labels[LabelProvider] != "provider-a" {
				t.Fatalf("unexpected provider label: %v", rec.Labels)
			}
		}
	}
	if !found {
		t.Fatalf("missing send latency record")
	}
}
