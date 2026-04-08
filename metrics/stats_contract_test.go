package metrics

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCollectorStatsContract(t *testing.T) {
	ctx := context.Background()
	started := time.Now()

	tests := []struct {
		name      string
		collector AggregateCollector
	}{
		{name: "base", collector: NewBaseMetricsCollector()},
		{name: "noop", collector: NewNoopCollector()},
		{name: "multi", collector: NewMultiCollector(NewBaseMetricsCollector(), NewBaseMetricsCollector())},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.collector.Clear()
			tt.collector.ObserveHTTP(ctx, "GET", "/ok", 200, 123, 10*time.Millisecond)
			tt.collector.ObserveHTTP(ctx, "GET", "/error", 500, 64, 12*time.Millisecond)
			tt.collector.Record(ctx, MetricRecord{
				Name:      "owner_metric_error",
				Value:     1,
				Error:     errors.New("boom"),
				Timestamp: time.Now(),
			})

			stats := tt.collector.GetStats()

			if stats.TotalRecords < 0 {
				t.Fatalf("total records must be non-negative, got %d", stats.TotalRecords)
			}
			if stats.ErrorRecords < 0 {
				t.Fatalf("error records must be non-negative, got %d", stats.ErrorRecords)
			}
			if stats.ActiveSeries < 0 {
				t.Fatalf("active series must be non-negative, got %d", stats.ActiveSeries)
			}
			if stats.ErrorRecords > stats.TotalRecords {
				t.Fatalf("error records cannot exceed total records, got errors=%d total=%d", stats.ErrorRecords, stats.TotalRecords)
			}

			if tt.name == "noop" {
				if stats.TotalRecords != 0 {
					t.Fatalf("noop collector must keep total records at zero, got %d", stats.TotalRecords)
				}
				if stats.ErrorRecords != 0 {
					t.Fatalf("noop collector must keep error records at zero, got %d", stats.ErrorRecords)
				}
				if !stats.StartTime.IsZero() {
					t.Fatalf("noop collector start time should be zero, got %v", stats.StartTime)
				}
				if stats.TypeBreakdown == nil {
					t.Fatalf("noop collector must return an initialized type breakdown map")
				}
				return
			}

			if stats.TotalRecords == 0 {
				t.Fatalf("collector must report total records after observations")
			}
			if stats.ErrorRecords == 0 {
				t.Fatalf("collector must report error records after failing observations")
			}
			if stats.ActiveSeries == 0 {
				t.Fatalf("collector must report active series after observations")
			}
			if stats.StartTime.IsZero() {
				t.Fatalf("collector must report non-zero start time")
			}
			if stats.StartTime.Before(started.Add(-1 * time.Second)) {
				t.Fatalf("collector start time must be recent, got %v", stats.StartTime)
			}

			if stats.TypeBreakdown != nil {
				var breakdownTotal int64
				for _, value := range stats.TypeBreakdown {
					breakdownTotal += value
				}
				if breakdownTotal == 0 {
					t.Fatalf("type breakdown map is present but empty in total")
				}
				if stats.ActiveSeries < len(stats.TypeBreakdown) {
					t.Fatalf("active series should not be lower than breakdown series count, active=%d breakdown=%d", stats.ActiveSeries, len(stats.TypeBreakdown))
				}
			}
		})
	}
}
