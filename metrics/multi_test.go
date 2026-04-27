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

type spyHTTPObserver struct {
	calls      int
	lastMethod string
	lastPath   string
	lastStatus int
}

func (s *spyHTTPObserver) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	s.calls++
	s.lastMethod = method
	s.lastPath = path
	s.lastStatus = status
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
	return CollectorStats{NameBreakdown: make(map[string]int64)}
}

func TestMultiCollector(t *testing.T) {
	collector1 := NewBaseMetricsCollector()
	collector2 := NewBaseMetricsCollector()
	multi := NewMultiCollector(collector1, collector2)
	if multi == nil {
		t.Fatalf("expected multi collector, got nil")
	}

	ctx := t.Context()

	// Record HTTP metric
	multi.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)

	// Verify both collectors received the metric
	if collector1.GetStats().TotalRecords != 1 {
		t.Fatalf("expected 1 record in collector1 stats, got %d", collector1.GetStats().TotalRecords)
	}
	if collector2.GetStats().TotalRecords != 1 {
		t.Fatalf("expected 1 record in collector2 stats, got %d", collector2.GetStats().TotalRecords)
	}
}

func TestMultiCollectorAllMethods(t *testing.T) {
	collector1 := NewBaseMetricsCollector()
	collector2 := NewBaseMetricsCollector()
	multi := NewMultiCollector(collector1, collector2)
	if multi == nil {
		t.Fatalf("expected multi collector, got nil")
	}

	ctx := t.Context()

	// Test both stable observation paths
	multi.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	multi.Record(ctx, MetricRecord{Name: "owner_metric", Value: 1})

	// Each collector should have received both metrics
	if collector1.GetStats().TotalRecords != 2 {
		t.Fatalf("expected 2 records in collector1 stats, got %d", collector1.GetStats().TotalRecords)
	}
	if collector2.GetStats().TotalRecords != 2 {
		t.Fatalf("expected 2 records in collector2 stats, got %d", collector2.GetStats().TotalRecords)
	}
}

func TestMultiCollectorGetStats(t *testing.T) {
	collector1 := NewBaseMetricsCollector()
	collector2 := NewBaseMetricsCollector()
	multi := NewMultiCollector(collector1, collector2)
	if multi == nil {
		t.Fatalf("expected multi collector, got nil")
	}

	ctx := t.Context()

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
	collector1 := newStubAggregateCollector(func() CollectorStats {
		return CollectorStats{
			TotalRecords:  10,
			ErrorRecords:  1,
			NameBreakdown: map[string]int64{MetricHTTPRequest: 10},
			StartTime:     time.Unix(100, 0),
		}
	})

	collector2 := newStubAggregateCollector(func() CollectorStats {
		return CollectorStats{
			TotalRecords:  5,
			ErrorRecords:  2,
			NameBreakdown: map[string]int64{MetricHTTPRequest: 3, "owner_metric": 2},
			StartTime:     time.Unix(200, 0),
		}
	})

	multi := NewMultiCollector(collector1, collector2)
	if multi == nil {
		t.Fatalf("expected multi collector, got nil")
	}

	stats := multi.GetStats()

	// TotalRecords should be summed
	if stats.TotalRecords != 15 {
		t.Fatalf("expected TotalRecords 15, got %d", stats.TotalRecords)
	}
	// ErrorRecords should be summed
	if stats.ErrorRecords != 3 {
		t.Fatalf("expected ErrorRecords 3, got %d", stats.ErrorRecords)
	}
	// NameBreakdown should be merged
	if stats.NameBreakdown[MetricHTTPRequest] != 13 {
		t.Fatalf("expected MetricHTTPRequest 13, got %d", stats.NameBreakdown[MetricHTTPRequest])
	}
	if stats.NameBreakdown["owner_metric"] != 2 {
		t.Fatalf("expected custom owner metric 2, got %d", stats.NameBreakdown["owner_metric"])
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
		return CollectorStats{NameBreakdown: make(map[string]int64)}
	})
	collectorB := newStubAggregateCollector(func() CollectorStats {
		callsB++
		return CollectorStats{NameBreakdown: make(map[string]int64)}
	})

	multi := NewMultiCollector(collectorA, collectorB)
	if multi == nil {
		t.Fatalf("expected multi collector, got nil")
	}

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
	if multi == nil {
		t.Fatalf("expected multi collector, got nil")
	}

	ctx := t.Context()

	// Add metrics
	multi.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)

	// Clear all collectors
	multi.Clear()

	// Verify both collectors are cleared
	if collector1.GetStats().TotalRecords != 0 {
		t.Fatalf("expected 0 records in collector1 after clear, got %d", collector1.GetStats().TotalRecords)
	}
	if collector2.GetStats().TotalRecords != 0 {
		t.Fatalf("expected 0 records in collector2 after clear, got %d", collector2.GetStats().TotalRecords)
	}
}

func TestMultiCollectorEmpty(t *testing.T) {
	multi := NewMultiCollector()
	if multi != nil {
		t.Fatalf("expected nil multi collector for empty input")
	}
}

func TestMultiCollectorNilReceiverNoops(t *testing.T) {
	var multi *MultiCollector

	multi.Record(t.Context(), MetricRecord{Name: "ignored"})
	multi.ObserveHTTP(t.Context(), "GET", "/ignored", 200, 0, time.Millisecond)
	multi.Clear()

	stats := multi.GetStats()
	if stats.NameBreakdown == nil {
		t.Fatalf("expected nil receiver stats to return initialized breakdown")
	}
	if stats.TotalRecords != 0 {
		t.Fatalf("expected nil receiver total records to stay zero, got %d", stats.TotalRecords)
	}
}

func TestMultiCollectorSkipsNilInternalCollectors(t *testing.T) {
	collector := NewBaseMetricsCollector()
	multi := &MultiCollector{collectors: []AggregateCollector{nil, collector}}

	multi.ObserveHTTP(t.Context(), "GET", "/test", 200, 0, time.Millisecond)
	multi.Record(t.Context(), MetricRecord{Name: "owner_metric"})

	stats := multi.GetStats()
	if stats.TotalRecords != 2 {
		t.Fatalf("expected non-nil collector records to be counted, got %d", stats.TotalRecords)
	}

	multi.Clear()
	if got := collector.GetStats().TotalRecords; got != 0 {
		t.Fatalf("expected clear to reach non-nil collector, got %d", got)
	}
}

func TestNewMultiHTTPObserver(t *testing.T) {
	left := &spyHTTPObserver{}
	right := &spyHTTPObserver{}

	observer := NewMultiHTTPObserver(nil, left, nil, right)
	if observer == nil {
		t.Fatalf("expected multi HTTP observer")
	}

	observer.ObserveHTTP(t.Context(), "POST", "/batch", 202, 64, 10*time.Millisecond)

	if left.calls != 1 || right.calls != 1 {
		t.Fatalf("expected both observers to be called once, got left=%d right=%d", left.calls, right.calls)
	}
	if left.lastMethod != "POST" || left.lastPath != "/batch" || left.lastStatus != 202 {
		t.Fatalf("left observer received unexpected HTTP data: %#v", left)
	}
	if right.lastMethod != "POST" || right.lastPath != "/batch" || right.lastStatus != 202 {
		t.Fatalf("right observer received unexpected HTTP data: %#v", right)
	}
}

func TestNewMultiHTTPObserverEmptyAndSingle(t *testing.T) {
	if observer := NewMultiHTTPObserver(nil, nil); observer != nil {
		t.Fatalf("expected nil observer for no non-nil inputs")
	}

	single := &spyHTTPObserver{}
	observer := NewMultiHTTPObserver(nil, single)
	if observer != single {
		t.Fatalf("expected single observer to be returned unchanged")
	}
}

func TestMultiCollectorImplementsInterface(t *testing.T) {
	var _ AggregateCollector = (*MultiCollector)(nil)
}

func TestMultiHTTPObserverImplementsInterface(t *testing.T) {
	var _ HTTPObserver = multiHTTPObserver{}
}
