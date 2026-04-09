package messaging

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/x/observability/recordbuffer"
	testmetrics "github.com/spcent/plumego/x/observability/testmetrics"
)

func TestReporterRecords(t *testing.T) {
	collector := testmetrics.NewMockCollector()
	reporter := NewSMSMetricsReporter(collector)

	ctx := context.Background()
	reporter.RecordQueueDepth(ctx, "send", SMSQueueStats{Queued: 2, Leased: 1, Dead: 0, Expired: 0})
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
		if rec.Type == SMSGatewayMetricType && rec.Name == MetricSendLatency {
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

func TestSMSPrometheusExporterWritesMetrics(t *testing.T) {
	collector := recordbuffer.NewCollector()
	reporter := NewSMSMetricsReporter(collector)

	ctx := context.Background()
	reporter.RecordQueueDepth(ctx, "send", SMSQueueStats{Queued: 3})
	reporter.RecordSendLatency(ctx, "tenant-1", "provider-a", 150*time.Millisecond, nil)
	reporter.RecordProviderResult(ctx, "tenant-1", "provider-a", true)
	reporter.RecordRetry(ctx, "tenant-1", "provider-a", 2)
	reporter.RecordStatus(ctx, "tenant-1", "sent")
	reporter.RecordReceiptDelay(ctx, "tenant-1", "provider-a", 2*time.Second)

	exporter := NewSMSPrometheusExporter("plumego_test", collector)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/sms", nil)
	exporter.Handler().ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "plumego_test_sms_gateway_queue_depth") {
		t.Fatalf("expected sms gateway queue depth metric")
	}
	if !strings.Contains(body, "plumego_test_sms_gateway_send_latency_seconds_sum") {
		t.Fatalf("expected send latency metric")
	}
	if !strings.Contains(body, "plumego_test_sms_gateway_provider_result_total") {
		t.Fatalf("expected provider result metric")
	}
	if !strings.Contains(body, "plumego_test_sms_gateway_receipt_delay_seconds_sum") {
		t.Fatalf("expected receipt delay metric")
	}
}

func TestSMSPrometheusExporterUsesValueWhenDurationMissing(t *testing.T) {
	collector := recordbuffer.NewCollector()
	collector.Record(context.Background(), metrics.MetricRecord{
		Type:  SMSGatewayMetricType,
		Name:  MetricSendLatency,
		Value: 0.125,
		Labels: metrics.MetricLabels{
			LabelTenant:   "tenant-1",
			LabelProvider: "provider-a",
		},
	})

	exporter := NewSMSPrometheusExporter("plumego_test", collector)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/sms", nil)
	exporter.Handler().ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), `plumego_test_sms_gateway_send_latency_seconds_sum{tenant="tenant-1",provider="provider-a"} 0.125000000`) {
		t.Fatalf("expected seconds-based send latency sum, body:\n%s", rec.Body.String())
	}
}
