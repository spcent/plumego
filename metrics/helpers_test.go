package metrics

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	timer := NewTimer()
	time.Sleep(50 * time.Millisecond)
	duration := timer.Elapsed()

	if duration < 50*time.Millisecond {
		t.Fatalf("expected duration >= 50ms, got %v", duration)
	}
}

func TestTimerReset(t *testing.T) {
	timer := NewTimer()
	time.Sleep(50 * time.Millisecond)

	timer.Reset()
	time.Sleep(10 * time.Millisecond)
	duration := timer.Elapsed()

	if duration >= 50*time.Millisecond {
		t.Fatalf("expected duration < 50ms after reset, got %v", duration)
	}
	if duration < 10*time.Millisecond {
		t.Fatalf("expected duration >= 10ms, got %v", duration)
	}
}

func TestMeasureFuncSuccess(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	err := MeasureFunc(ctx, collector, "test_op", "subject", func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	records := collector.GetRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].Name != "test_op" {
		t.Fatalf("expected operation 'test_op', got %s", records[0].Name)
	}
	if records[0].Error != nil {
		t.Fatalf("expected no error in record, got %v", records[0].Error)
	}
}

func TestMeasureFuncError(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()
	testErr := errors.New("test error")

	err := MeasureFunc(ctx, collector, "test_op", "subject", func() error {
		return testErr
	})

	if err != testErr {
		t.Fatalf("expected test error, got %v", err)
	}

	records := collector.GetRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].Error != testErr {
		t.Fatalf("expected error in record, got %v", records[0].Error)
	}
}

func TestMeasureFuncWithKV(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	err := MeasureFunc(ctx, collector, "get", "key123", func() error {
		return nil
	}, MeasureWithKV(true))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	records := collector.GetRecords()
	// KV operations record twice (once for operation, once for hit/miss)
	if len(records) < 1 {
		t.Fatalf("expected at least 1 record, got %d", len(records))
	}
}

func TestMeasureFuncWithPubSub(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	err := MeasureFunc(ctx, collector, "publish", "topic", func() error {
		return nil
	}, MeasureWithPubSub())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	records := collector.GetRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].Type != MetricPubSubPublish {
		t.Fatalf("expected MetricPubSubPublish, got %v", records[0].Type)
	}
}

func TestMeasureFuncWithMQ(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	err := MeasureFunc(ctx, collector, "subscribe", "queue", func() error {
		return nil
	}, MeasureWithMQ(false))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	records := collector.GetRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].Type != MetricMQSubscribe {
		t.Fatalf("expected MetricMQSubscribe, got %v", records[0].Type)
	}
}

func TestMeasureFuncWithIPC(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	err := MeasureFunc(ctx, collector, "read", "/tmp/socket", func() error {
		return nil
	}, MeasureWithIPC("unix", 256))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	records := collector.GetRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].Type != MetricIPCRead {
		t.Fatalf("expected MetricIPCRead, got %v", records[0].Type)
	}
}

func TestRecordSuccess(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	RecordSuccess(ctx, collector, "test_operation", 50*time.Millisecond)

	records := collector.GetRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].Name != "test_operation" {
		t.Fatalf("expected 'test_operation', got %s", records[0].Name)
	}
	if records[0].Error != nil {
		t.Fatalf("expected no error, got %v", records[0].Error)
	}
	if records[0].Duration != 50*time.Millisecond {
		t.Fatalf("expected 50ms duration, got %v", records[0].Duration)
	}
}

func TestRecordError(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()
	testErr := errors.New("test error")

	RecordError(ctx, collector, "failed_operation", 100*time.Millisecond, testErr)

	records := collector.GetRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].Name != "failed_operation" {
		t.Fatalf("expected 'failed_operation', got %s", records[0].Name)
	}
	if records[0].Error != testErr {
		t.Fatalf("expected test error, got %v", records[0].Error)
	}
}

func TestRecordWithLabels(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	labels := MetricLabels{
		"endpoint": "/api/users",
		"method":   "GET",
		"status":   "success",
	}

	RecordWithLabels(ctx, collector, "api_call", 75*time.Millisecond, labels)

	records := collector.GetRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].Labels["endpoint"] != "/api/users" {
		t.Fatalf("expected endpoint label '/api/users', got %s", records[0].Labels["endpoint"])
	}
	if records[0].Labels["method"] != "GET" {
		t.Fatalf("expected method label 'GET', got %s", records[0].Labels["method"])
	}
}

func TestMultiCollector(t *testing.T) {
	collector1 := NewBaseMetricsCollector()
	collector2 := NewBaseMetricsCollector()
	multi := NewMultiCollector(collector1, collector2)

	ctx := context.Background()

	// Record HTTP metric
	multi.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)

	// Verify both collectors received the metric
	records1 := collector1.GetRecords()
	records2 := collector2.GetRecords()

	if len(records1) != 1 {
		t.Fatalf("expected 1 record in collector1, got %d", len(records1))
	}
	if len(records2) != 1 {
		t.Fatalf("expected 1 record in collector2, got %d", len(records2))
	}
}

func TestMultiCollectorAllMethods(t *testing.T) {
	collector1 := NewBaseMetricsCollector()
	collector2 := NewBaseMetricsCollector()
	multi := NewMultiCollector(collector1, collector2)

	ctx := context.Background()

	// Test all observation methods
	multi.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	multi.ObservePubSub(ctx, "publish", "topic", 10*time.Millisecond, nil)
	multi.ObserveMQ(ctx, "subscribe", "queue", 5*time.Millisecond, nil, false)
	multi.ObserveKV(ctx, "get", "key", 2*time.Millisecond, nil, true)
	multi.ObserveIPC(ctx, "read", "/socket", "unix", 256, 1*time.Millisecond, nil)

	// Each collector should have received all metrics
	// Note: KV operations record twice (operation + hit/miss)
	records1 := collector1.GetRecords()
	records2 := collector2.GetRecords()

	if len(records1) < 5 {
		t.Fatalf("expected at least 5 records in collector1, got %d", len(records1))
	}
	if len(records2) < 5 {
		t.Fatalf("expected at least 5 records in collector2, got %d", len(records2))
	}
}

func TestMultiCollectorGetStats(t *testing.T) {
	collector1 := NewBaseMetricsCollector()
	collector2 := NewBaseMetricsCollector()
	multi := NewMultiCollector(collector1, collector2)

	ctx := context.Background()

	// Add metrics to both collectors via multi
	multi.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	multi.ObserveHTTP(ctx, "POST", "/api", 201, 200, 100*time.Millisecond)

	stats := multi.GetStats()

	// Stats should be aggregated from both collectors
	// Each metric goes to both collectors, so total should be 2+2=4
	if stats.TotalRecords < 2 {
		t.Fatalf("expected at least 2 total records, got %d", stats.TotalRecords)
	}
}

func TestMultiCollectorClear(t *testing.T) {
	collector1 := NewBaseMetricsCollector()
	collector2 := NewBaseMetricsCollector()
	multi := NewMultiCollector(collector1, collector2)

	ctx := context.Background()

	// Add metrics
	multi.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)

	// Clear all collectors
	multi.Clear()

	// Verify both collectors are cleared
	records1 := collector1.GetRecords()
	records2 := collector2.GetRecords()

	if len(records1) != 0 {
		t.Fatalf("expected 0 records in collector1 after clear, got %d", len(records1))
	}
	if len(records2) != 0 {
		t.Fatalf("expected 0 records in collector2 after clear, got %d", len(records2))
	}
}

func TestMultiCollectorEmpty(t *testing.T) {
	multi := NewMultiCollector()

	stats := multi.GetStats()

	// Empty multi collector should return zero stats
	if stats.TotalRecords != 0 {
		t.Fatalf("expected 0 total records, got %d", stats.TotalRecords)
	}
}

func TestMultiCollectorImplementsInterface(t *testing.T) {
	var _ MetricsCollector = (*MultiCollector)(nil)
}

// Benchmark helpers
func BenchmarkTimerElapsed(b *testing.B) {
	timer := NewTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = timer.Elapsed()
	}
}

func BenchmarkRecordSuccess(b *testing.B) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RecordSuccess(ctx, collector, "test", 50*time.Millisecond)
	}
}

func BenchmarkMultiCollectorObserveHTTP(b *testing.B) {
	collector1 := NewBaseMetricsCollector()
	collector2 := NewBaseMetricsCollector()
	multi := NewMultiCollector(collector1, collector2)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		multi.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}
}
