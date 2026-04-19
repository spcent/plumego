package metrics

import (
	"testing"
	"time"
)

func BenchmarkBaselineMapAccess(b *testing.B) {
	m := map[string]int{"key": 100}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m["key"]
	}
}

func BenchmarkBaselineTimeNow(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = time.Now()
	}
}

func BenchmarkBaseCollectorRecord(b *testing.B) {
	collector := NewBaseMetricsCollector()
	ctx := b.Context()
	record := MetricRecord{
		Name:  "test",
		Value: 100,
		Labels: MetricLabels{
			"method": "GET",
			"path":   "/test",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Record(ctx, record)
	}
}

func BenchmarkBaseCollectorObserveHTTP(b *testing.B) {
	collector := NewBaseMetricsCollector()
	ctx := b.Context()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}
}

func BenchmarkNoopCollectorObserveHTTP(b *testing.B) {
	collector := NewNoopCollector()
	ctx := b.Context()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}
}

func BenchmarkMultiCollectorObserveHTTP(b *testing.B) {
	left := NewBaseMetricsCollector()
	right := NewBaseMetricsCollector()
	collector := NewMultiCollector(left, right)
	ctx := b.Context()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}
}
