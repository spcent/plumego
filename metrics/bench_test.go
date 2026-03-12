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

func BenchmarkDevStatBucketSnapshotCachedPercentiles(b *testing.B) {
	bucket := newStatBucket(2000)
	for i := 0; i < 2000; i++ {
		bucket.record(float64((i*53)%1000) / 10.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bucket.snapshot()
	}
}

// Benchmark baseline operations
func BenchmarkBaselineMapAccess(b *testing.B) {
	m := make(map[string]int)
	m["key"] = 100

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

// Collector benchmarks
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

func BenchmarkPrometheusCollectorObserveHTTP(b *testing.B) {
	collector := NewPrometheusCollector("bench")
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
	base := NewBaseMetricsCollector()
	prom := NewPrometheusCollector("bench")
	multi := NewMultiCollector(base, prom)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		multi.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}
}

// PubSub benchmarks
func BenchmarkBaseCollectorObservePubSub(b *testing.B) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObservePubSub(ctx, "publish", "test.topic", 10*time.Millisecond, nil)
	}
}

func BenchmarkPrometheusCollectorObservePubSub(b *testing.B) {
	collector := NewPrometheusCollector("bench")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObservePubSub(ctx, "publish", "test.topic", 10*time.Millisecond, nil)
	}
}

// KV benchmarks
func BenchmarkBaseCollectorObserveKV(b *testing.B) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObserveKV(ctx, "get", "key123", 2*time.Millisecond, nil, true)
	}
}

func BenchmarkPrometheusCollectorObserveKV(b *testing.B) {
	collector := NewPrometheusCollector("bench")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObserveKV(ctx, "get", "key123", 2*time.Millisecond, nil, true)
	}
}

// GetStats benchmarks
func BenchmarkBaseCollectorGetStats(b *testing.B) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = collector.GetStats()
	}
}

func BenchmarkPrometheusCollectorGetStats(b *testing.B) {
	collector := NewPrometheusCollector("bench")
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = collector.GetStats()
	}
}

// Clear benchmarks
func BenchmarkBaseCollectorClear(b *testing.B) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Pre-populate before each clear
		for j := 0; j < 100; j++ {
			collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
		}
		b.StartTimer()

		collector.Clear()
	}
}

// Helper function benchmarks
func BenchmarkTimerElapsedCall(b *testing.B) {
	timer := NewTimer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = timer.Elapsed()
	}
}

func BenchmarkMeasureFuncSimple(b *testing.B) {
	collector := NewNoopCollector()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MeasureFunc(ctx, collector, "test", func() error {
			return nil
		})
	}
}

// Memory allocation benchmarks
func BenchmarkBaseCollectorRecordAllocs(b *testing.B) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()
	record := MetricRecord{
		Type:  MetricHTTPRequest,
		Name:  "test",
		Value: 100,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Record(ctx, record)
	}
}

func BenchmarkPrometheusCollectorObserveHTTPAllocs(b *testing.B) {
	collector := NewPrometheusCollector("bench")
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}
}

// Concurrent benchmarks
func BenchmarkBaseCollectorConcurrent(b *testing.B) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
		}
	})
}

func BenchmarkPrometheusCollectorConcurrent(b *testing.B) {
	collector := NewPrometheusCollector("bench")
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
		}
	})
}

func BenchmarkNoopCollectorConcurrent(b *testing.B) {
	collector := NewNoopCollector()
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
		}
	})
}

// High cardinality benchmarks (stress test)
func BenchmarkPrometheusCollectorHighCardinality(b *testing.B) {
	collector := NewPrometheusCollector("bench").WithMaxMemory(10000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := "/api/endpoint" + string(rune(i%1000))
		collector.ObserveHTTP(ctx, "GET", path, 200, 100, 50*time.Millisecond)
	}
}

// Label cloning benchmarks
func BenchmarkCloneLabelsSmall(b *testing.B) {
	labels := MetricLabels{
		"method": "GET",
		"path":   "/test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cloneLabels(labels)
	}
}

func BenchmarkCloneLabelsMedium(b *testing.B) {
	labels := MetricLabels{
		"method":    "GET",
		"path":      "/test",
		"status":    "200",
		"host":      "localhost",
		"user":      "test",
		"operation": "read",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cloneLabels(labels)
	}
}
