package metrics

import (
	"context"
	"testing"
	"time"
)

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
		_ = MeasureFunc(ctx, collector, "test", "subject", func() error {
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
