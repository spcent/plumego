package metrics

import (
	"sync"
	"testing"
	"time"
)

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
	if stats.NameBreakdown == nil {
		t.Fatalf("expected initialized name breakdown")
	}
	stats.NameBreakdown["caller_owned"] = 1
}

func TestNoopCollectorGetStatsReturnsCallerOwnedBreakdown(t *testing.T) {
	collector := NewNoopCollector()

	stats := collector.GetStats()
	stats.NameBreakdown["caller_owned"] = 1

	stats = collector.GetStats()
	if stats.NameBreakdown["caller_owned"] != 0 {
		t.Fatalf("expected noop stats breakdown to be caller-owned")
	}
}

func TestNoopCollectorClear(t *testing.T) {
	collector := NewNoopCollector()

	// Should not panic
	collector.Clear()
}

func TestNoopCollectorConcurrency(t *testing.T) {
	collector := NewNoopCollector()

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			collector.ObserveHTTP(t.Context(), "GET", "/test", 200, 100, 50*time.Millisecond)
			collector.GetStats()
			collector.Clear()
		}()
	}
	wg.Wait()
}
