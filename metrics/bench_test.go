package metrics

import (
	"context"
	"sort"
	"testing"
	"time"
)

func percentileInsertionSort(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)

	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	index := int(float64(len(sorted)-1) * p)
	return sorted[index]
}

func benchmarkPercentileValues(n int) []float64 {
	values := make([]float64, n)
	for i := 0; i < n; i++ {
		values[i] = float64((i*7919)%104729) / 10.0
	}
	return values
}

func BenchmarkPercentileInsertionSort(b *testing.B) {
	values := benchmarkPercentileValues(2000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = percentileInsertionSort(values, 0.95)
	}
}

func BenchmarkPercentileSortFloat64s(b *testing.B) {
	values := benchmarkPercentileValues(2000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = percentile(values, 0.95)
	}
}

func BenchmarkPercentileSortedReuse(b *testing.B) {
	values := benchmarkPercentileValues(2000)
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = percentileSorted(sorted, 0.95)
	}
}

func BenchmarkAggregatorGetStatsCachedPercentiles(b *testing.B) {
	agg := NewAggregator(5 * time.Minute)
	for i := 0; i < 2000; i++ {
		agg.Record("latency", float64((i*37)%1000))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = agg.GetStats("latency")
	}
}

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
	ctx := context.Background()
	record := MetricRecord{
		Type:  MetricHTTPRequest,
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
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}
}

func BenchmarkNoopCollectorObserveHTTP(b *testing.B) {
	collector := NewNoopCollector()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}
}

func BenchmarkMultiCollectorObserveHTTP(b *testing.B) {
	left := NewBaseMetricsCollector()
	right := NewBaseMetricsCollector()
	collector := NewMultiCollector(left, right)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}
}
