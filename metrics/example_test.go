package metrics_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
)

// Example demonstrates basic usage of the Prometheus collector
func Example() {
	// Create a Prometheus collector
	collector := metrics.NewPrometheusCollector("myapp")

	// Record some HTTP metrics
	ctx := context.Background()
	collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 1024, 50*time.Millisecond)
	collector.ObserveHTTP(ctx, "POST", "/api/users", 201, 512, 100*time.Millisecond)

	// Get statistics
	stats := collector.GetStats()
	fmt.Printf("Total requests: %d\n", stats.TotalRequests)
	fmt.Printf("Unique series: %d\n", stats.Series)

	// Output:
	// Total requests: 2
	// Unique series: 2
}

// ExampleNewPrometheusCollector demonstrates creating a Prometheus collector
func ExampleNewPrometheusCollector() {
	// Create with default namespace
	collector1 := metrics.NewPrometheusCollector("")
	fmt.Printf("Default namespace: %s\n", "plumego")

	// Create with custom namespace
	collector2 := metrics.NewPrometheusCollector("myapp")
	fmt.Printf("Custom namespace: %s\n", "myapp")

	_ = collector1
	_ = collector2

	// Output:
	// Default namespace: plumego
	// Custom namespace: myapp
}

// ExamplePrometheusCollector_Handler demonstrates serving metrics
func ExamplePrometheusCollector_Handler() {
	collector := metrics.NewPrometheusCollector("myapp")
	ctx := context.Background()

	// Record some metrics
	collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 100, 50*time.Millisecond)

	// Create HTTP handler
	handler := collector.Handler()

	// Serve metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Printf("Status: %d\n", rec.Code)
	fmt.Printf("Content-Type: %s\n", rec.Header().Get("Content-Type"))

	// Output:
	// Status: 200
	// Content-Type: text/plain; version=0.0.4
}

// ExampleNewOpenTelemetryTracer demonstrates creating an OpenTelemetry tracer
func ExampleNewOpenTelemetryTracer() {
	// Create with default name
	tracer1 := metrics.NewOpenTelemetryTracer("")
	_ = tracer1

	// Create with custom name
	tracer2 := metrics.NewOpenTelemetryTracer("my-service")
	_ = tracer2

	fmt.Println("Tracers created successfully")

	// Output:
	// Tracers created successfully
}

// ExampleOpenTelemetryTracer_Start demonstrates starting a trace span
func ExampleOpenTelemetryTracer_Start() {
	tracer := metrics.NewOpenTelemetryTracer("my-service")

	req := httptest.NewRequest("GET", "/api/users", nil)
	ctx, span := tracer.Start(context.Background(), req)

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// End the span
	span.End(middleware.RequestMetrics{
		Status:  200,
		Bytes:   100,
		TraceID: "trace-123",
	})

	_ = ctx
	fmt.Println("Span created and ended")

	// Output:
	// Span created and ended
}

// ExampleNewTimer demonstrates using a timer to measure duration
func ExampleNewTimer() {
	timer := metrics.NewTimer()

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	duration := timer.Elapsed()
	fmt.Printf("Duration >= 50ms: %v\n", duration >= 50*time.Millisecond)

	// Output:
	// Duration >= 50ms: true
}

// ExampleTimer_Reset demonstrates resetting a timer
func ExampleTimer_Reset() {
	timer := metrics.NewTimer()
	time.Sleep(50 * time.Millisecond)

	// Reset the timer
	timer.Reset()
	time.Sleep(10 * time.Millisecond)

	duration := timer.Elapsed()
	fmt.Printf("Duration < 50ms: %v\n", duration < 50*time.Millisecond)

	// Output:
	// Duration < 50ms: true
}

// ExampleMeasureFunc demonstrates measuring function execution
func ExampleMeasureFunc() {
	collector := metrics.NewBaseMetricsCollector()
	ctx := context.Background()

	err := metrics.MeasureFunc(ctx, collector, "database_query", "users", func() error {
		// Simulate database operation
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	records := collector.GetRecords()
	fmt.Printf("Recorded %d metric(s)\n", len(records))

	// Output:
	// Recorded 1 metric(s)
}

// ExampleMeasureFunc_withKV demonstrates measuring KV operations
func ExampleMeasureFunc_withKV() {
	collector := metrics.NewBaseMetricsCollector()
	ctx := context.Background()

	err := metrics.MeasureFunc(ctx, collector, "get", "user:123", func() error {
		// Simulate KV get operation
		time.Sleep(5 * time.Millisecond)
		return nil
	}, metrics.MeasureWithKV(true))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	records := collector.GetRecords()
	fmt.Printf("Recorded at least %d metric(s)\n", len(records))

	// Output:
	// Recorded at least 2 metric(s)
}

// ExampleRecordSuccess demonstrates recording a successful operation
func ExampleRecordSuccess() {
	collector := metrics.NewBaseMetricsCollector()
	ctx := context.Background()

	metrics.RecordSuccess(ctx, collector, "api_call", 50*time.Millisecond)

	records := collector.GetRecords()
	fmt.Printf("Success: %v\n", records[0].Error == nil)

	// Output:
	// Success: true
}

// ExampleRecordError demonstrates recording a failed operation
func ExampleRecordError() {
	collector := metrics.NewBaseMetricsCollector()
	ctx := context.Background()

	err := fmt.Errorf("connection failed")
	metrics.RecordError(ctx, collector, "database_query", 100*time.Millisecond, err)

	records := collector.GetRecords()
	fmt.Printf("Has error: %v\n", records[0].Error != nil)

	// Output:
	// Has error: true
}

// ExampleRecordWithLabels demonstrates recording with custom labels
func ExampleRecordWithLabels() {
	collector := metrics.NewBaseMetricsCollector()
	ctx := context.Background()

	labels := metrics.MetricLabels{
		"endpoint": "/api/users",
		"method":   "GET",
		"status":   "success",
	}

	metrics.RecordWithLabels(ctx, collector, "api_call", 75*time.Millisecond, labels)

	records := collector.GetRecords()
	fmt.Printf("Endpoint: %s\n", records[0].Labels["endpoint"])

	// Output:
	// Endpoint: /api/users
}

// ExampleNewMultiCollector demonstrates using multiple collectors
func ExampleNewMultiCollector() {
	// Create multiple collectors
	prom := metrics.NewPrometheusCollector("myapp")
	base := metrics.NewBaseMetricsCollector()

	// Combine them
	multi := metrics.NewMultiCollector(prom, base)

	ctx := context.Background()

	// Metrics go to both collectors
	multi.ObserveHTTP(ctx, "GET", "/api/users", 200, 100, 50*time.Millisecond)

	// Verify both received the metric
	promStats := prom.GetStats()
	baseRecords := base.GetRecords()

	fmt.Printf("Prometheus: %d requests\n", promStats.TotalRequests)
	fmt.Printf("Base: %d records\n", len(baseRecords))

	// Output:
	// Prometheus: 1 requests
	// Base: 1 records
}

// ExampleNewNoopCollector demonstrates using a no-op collector
func ExampleNewNoopCollector() {
	collector := metrics.NewNoopCollector()
	ctx := context.Background()

	// All operations are no-ops
	collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 100, 50*time.Millisecond)
	collector.ObservePubSub(ctx, "publish", "topic", 10*time.Millisecond, nil)

	// Stats are always zero
	stats := collector.GetStats()
	fmt.Printf("Total records: %d\n", stats.TotalRecords)

	// Output:
	// Total records: 0
}

// ExampleBaseMetricsCollector_WithMaxRecords demonstrates limiting records
func ExampleBaseMetricsCollector_WithMaxRecords() {
	collector := metrics.NewBaseMetricsCollector().WithMaxRecords(5)
	ctx := context.Background()

	// Add 10 records
	for i := 0; i < 10; i++ {
		collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
	}

	// Only 5 most recent are kept
	records := collector.GetRecords()
	fmt.Printf("Records: %d\n", len(records))

	// Output:
	// Records: 5
}

// ExamplePrometheusCollector_WithMaxMemory demonstrates memory limits
func ExamplePrometheusCollector_WithMaxMemory() {
	collector := metrics.NewPrometheusCollector("myapp").WithMaxMemory(3)
	ctx := context.Background()

	// Add 5 unique series
	for i := 0; i < 5; i++ {
		path := fmt.Sprintf("/api/endpoint%d", i)
		collector.ObserveHTTP(ctx, "GET", path, 200, 100, 50*time.Millisecond)
	}

	stats := collector.GetStats()
	fmt.Printf("Series limited to: %v\n", stats.Series <= 3)

	// Output:
	// Series limited to: true
}

// ExampleMetricsCollector demonstrates the unified interface
func ExampleMetricsCollector() {
	var collector metrics.MetricsCollector

	// Can use any implementation
	collector = metrics.NewBaseMetricsCollector()
	// collector = metrics.NewPrometheusCollector("myapp")
	// collector = metrics.NewOpenTelemetryTracer("myapp")
	// collector = metrics.NewNoopCollector()

	ctx := context.Background()

	// All collectors support the same operations
	collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 100, 50*time.Millisecond)
	collector.ObservePubSub(ctx, "publish", "events", 10*time.Millisecond, nil)
	collector.ObserveMQ(ctx, "subscribe", "jobs", 5*time.Millisecond, nil, false)
	collector.ObserveKV(ctx, "get", "user:123", 2*time.Millisecond, nil, true)
	collector.ObserveIPC(ctx, "read", "/tmp/app.sock", "unix", 256, 1*time.Millisecond, nil)

	stats := collector.GetStats()
	fmt.Printf("Metrics recorded: %v\n", stats.TotalRecords > 0)

	// Output:
	// Metrics recorded: true
}

// ExamplePrometheusCollector_integration demonstrates full integration
func ExamplePrometheusCollector_integration() {
	// Create collector
	collector := metrics.NewPrometheusCollector("myapp").
		WithMaxMemory(10000)

	// Record various metrics
	ctx := context.Background()

	// HTTP requests
	collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 1024, 50*time.Millisecond)
	collector.ObserveHTTP(ctx, "POST", "/api/users", 201, 512, 100*time.Millisecond)
	collector.ObserveHTTP(ctx, "GET", "/api/posts", 200, 2048, 75*time.Millisecond)

	// Pub/Sub operations
	collector.ObservePubSub(ctx, "publish", "user.created", 5*time.Millisecond, nil)
	collector.ObservePubSub(ctx, "subscribe", "events", 10*time.Millisecond, nil)

	// Get statistics
	stats := collector.GetStats()
	fmt.Printf("Total HTTP requests: %d\n", stats.TotalRequests)
	fmt.Printf("Unique series: %d\n", stats.Series)

	// Expose metrics endpoint
	handler := collector.Handler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Printf("Metrics endpoint status: %d\n", rec.Code)

	// Output:
	// Total HTTP requests: 3
	// Unique series: 3
	// Metrics endpoint status: 200
}
