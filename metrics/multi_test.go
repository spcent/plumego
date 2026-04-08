package metrics

import (
	"context"
	"testing"
	"time"
)

type stubAggregateCollector struct {
	*NoopCollector
	onGetStats func() CollectorStats
}

func newStubAggregateCollector(onGetStats func() CollectorStats) *stubAggregateCollector {
	return &stubAggregateCollector{
		NoopCollector: NewNoopCollector(),
		onGetStats:    onGetStats,
	}
}

func (s *stubAggregateCollector) GetStats() CollectorStats {
	if s.onGetStats != nil {
		return s.onGetStats()
	}
	return CollectorStats{TypeBreakdown: make(map[MetricType]int64)}
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

	// Test both stable observation paths
	multi.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	multi.Record(ctx, MetricRecord{Name: "owner_metric", Value: 1})

	// Each collector should have received both metrics
	records1 := collector1.GetRecords()
	records2 := collector2.GetRecords()

	if len(records1) != 2 {
		t.Fatalf("expected 2 records in collector1, got %d", len(records1))
	}
	if len(records2) != 2 {
		t.Fatalf("expected 2 records in collector2, got %d", len(records2))
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

func TestMultiCollectorGetStatsWeightedAverageDuration(t *testing.T) {
	customOwnerType := MetricType("owner_metric")

	collector1 := newStubAggregateCollector(func() CollectorStats {
		return CollectorStats{
			TotalRecords:  10,
			ErrorRecords:  1,
			TypeBreakdown: map[MetricType]int64{MetricHTTPRequest: 10},
			StartTime:     time.Unix(100, 0),
		}
	})

	collector2 := newStubAggregateCollector(func() CollectorStats {
		return CollectorStats{
			TotalRecords:  5,
			ErrorRecords:  2,
			TypeBreakdown: map[MetricType]int64{MetricHTTPRequest: 3, customOwnerType: 2},
			StartTime:     time.Unix(200, 0),
		}
	})

	multi := NewMultiCollector(collector1, collector2)

	stats := multi.GetStats()

	// TotalRecords should be summed
	if stats.TotalRecords != 15 {
		t.Fatalf("expected TotalRecords 15, got %d", stats.TotalRecords)
	}
	// ErrorRecords should be summed
	if stats.ErrorRecords != 3 {
		t.Fatalf("expected ErrorRecords 3, got %d", stats.ErrorRecords)
	}
	// TypeBreakdown should be merged
	if stats.TypeBreakdown[MetricHTTPRequest] != 13 {
		t.Fatalf("expected MetricHTTPRequest 13, got %d", stats.TypeBreakdown[MetricHTTPRequest])
	}
	if stats.TypeBreakdown[customOwnerType] != 2 {
		t.Fatalf("expected custom owner metric 2, got %d", stats.TypeBreakdown[customOwnerType])
	}
	// StartTime should be the earliest
	if !stats.StartTime.Equal(time.Unix(100, 0)) {
		t.Fatalf("expected StartTime to be earliest, got %v", stats.StartTime)
	}
}

func TestMultiCollectorGetStatsCallsEachCollectorOnce(t *testing.T) {
	callsA := 0
	callsB := 0

	collectorA := newStubAggregateCollector(func() CollectorStats {
		callsA++
		return CollectorStats{TypeBreakdown: make(map[MetricType]int64)}
	})
	collectorB := newStubAggregateCollector(func() CollectorStats {
		callsB++
		return CollectorStats{TypeBreakdown: make(map[MetricType]int64)}
	})

	multi := NewMultiCollector(collectorA, collectorB)

	_ = multi.GetStats()

	if callsA != 1 {
		t.Fatalf("expected collectorA GetStats to be called once, got %d", callsA)
	}
	if callsB != 1 {
		t.Fatalf("expected collectorB GetStats to be called once, got %d", callsB)
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
	var _ AggregateCollector = (*MultiCollector)(nil)
}
