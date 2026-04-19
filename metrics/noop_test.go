package metrics

import (
	"testing"
	"time"
)

func TestNoopCollectorImplementsInterface(t *testing.T) {
	var _ AggregateCollector = (*NoopCollector)(nil)
}

func TestNoopCollectorRecord(t *testing.T) {
	collector := NewNoopCollector()

	// Should not panic
	collector.Record(t.Context(), MetricRecord{
		Name:  "test",
		Value: 100,
	})
}

func TestNoopCollectorObserveHTTP(t *testing.T) {
	collector := NewNoopCollector()

	// Should not panic
	collector.ObserveHTTP(t.Context(), "GET", "/test", 200, 100, 50*time.Millisecond)
}

func TestNoopCollectorGetStats(t *testing.T) {
	collector := NewNoopCollector()

	// Record some metrics
	collector.ObserveHTTP(t.Context(), "GET", "/test", 200, 100, 50*time.Millisecond)

	stats := collector.GetStats()

	// Stats should be zero
	if stats.TotalRecords != 0 {
		t.Fatalf("expected 0 total records, got %d", stats.TotalRecords)
	}
	if stats.ErrorRecords != 0 {
		t.Fatalf("expected 0 error records, got %d", stats.ErrorRecords)
	}
	if stats.ActiveSeries != 0 {
		t.Fatalf("expected 0 active series, got %d", stats.ActiveSeries)
	}
	if !stats.StartTime.IsZero() {
		t.Fatalf("expected zero start time, got %v", stats.StartTime)
	}
}

func TestNoopCollectorClear(t *testing.T) {
	collector := NewNoopCollector()

	// Should not panic
	collector.Clear()
}

func TestNoopCollectorConcurrency(t *testing.T) {
	collector := NewNoopCollector()

	// Test concurrent operations
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			collector.ObserveHTTP(t.Context(), "GET", "/test", 200, 100, 50*time.Millisecond)
			collector.GetStats()
			collector.Clear()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}
