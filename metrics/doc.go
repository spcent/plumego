// Package metrics provides a unified interface for collecting, recording, and exporting
// application metrics and traces without external dependencies.
//
// The metrics package is designed to be lightweight, flexible, and compatible with
// popular observability systems like Prometheus and OpenTelemetry while requiring
// zero external dependencies.
//
// # Core Components
//
// The package provides collector implementations for different needs:
//
//  1. BaseMetricsCollector - A simple in-memory collector for testing and development
//  2. DevCollector - A lightweight dev dashboard collector with HTTP aggregation
//  3. PrometheusCollector - Prometheus-compatible metrics exposition
//  4. OpenTelemetryTracer - OpenTelemetry-compatible distributed tracing
//  5. NoopCollector - A no-op collector for disabled metrics or testing
//
// # Unified Interface
//
// All collectors implement the MetricsCollector interface, providing a consistent
// API for recording metrics across different subsystems:
//
//   - HTTP requests and responses
//   - Pub/Sub operations
//   - Message Queue operations
//   - Key-Value store operations
//   - IPC (Inter-Process Communication) operations
//
// # Basic Usage
//
// Creating and using a Prometheus collector:
//
//	import "github.com/spcent/plumego/metrics"
//
//	// Create collector
//	collector := metrics.NewPrometheusCollector("myapp")
//
//	// Record HTTP metrics
//	collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 1024, 50*time.Millisecond)
//
//	// Expose metrics endpoint
//	http.Handle("/metrics", collector.Handler())
//
// Using the OpenTelemetry tracer:
//
//	import "github.com/spcent/plumego/metrics"
//
//	// Create tracer
//	tracer := metrics.NewOpenTelemetryTracer("my-service")
//
//	// Start span
//	ctx, span := tracer.Start(context.Background(), request)
//	defer span.End(middleware.RequestMetrics{
//		Status:  http.StatusOK,
//		Bytes:   1024,
//		TraceID: span.TraceID(),
//	})
//
// # Recording Different Metric Types
//
// HTTP Metrics:
//
//	collector.ObserveHTTP(ctx, "POST", "/api/data", 201, 512, 100*time.Millisecond)
//
// Pub/Sub Metrics:
//
//	collector.ObservePubSub(ctx, "publish", "user.created", 5*time.Millisecond, nil)
//
// Message Queue Metrics:
//
//	collector.ObserveMQ(ctx, "subscribe", "orders", 10*time.Millisecond, nil, false)
//
// Key-Value Store Metrics:
//
//	collector.ObserveKV(ctx, "get", "user:123", 2*time.Millisecond, nil, true)
//
// IPC Metrics:
//
//	collector.ObserveIPC(ctx, "read", "/tmp/app.sock", "unix", 256, 1*time.Millisecond, nil)
//
// # Generic Metric Recording
//
// For custom metrics or advanced use cases, use the Record method:
//
//	record := metrics.MetricRecord{
//		Type:     metrics.MetricHTTPRequest,
//		Name:     "custom_metric",
//		Value:    42.5,
//		Labels:   metrics.MetricLabels{"env": "prod", "region": "us-east"},
//		Duration: 100 * time.Millisecond,
//		Error:    nil,
//	}
//	collector.Record(ctx, record)
//
// # Metric Types
//
// The package defines metric types for different operations:
//
//	HTTP Metrics:
//	  - MetricHTTPRequest
//
//	Pub/Sub Metrics:
//	  - MetricPubSubPublish
//	  - MetricPubSubSubscribe
//	  - MetricPubSubDeliver
//	  - MetricPubSubDrop
//
//	Message Queue Metrics:
//	  - MetricMQPublish
//	  - MetricMQSubscribe
//	  - MetricMQClose
//	  - MetricMQMetrics
//
//	Key-Value Store Metrics:
//	  - MetricKVSet, MetricKVGet, MetricKVDelete, MetricKVExists, MetricKVKeys
//	  - MetricKVHit, MetricKVMiss, MetricKVEvict
//
//	IPC Metrics:
//	  - MetricIPCAccept, MetricIPCDial, MetricIPCRead, MetricIPCWrite, MetricIPCClose
//
// # Prometheus Integration
//
// The PrometheusCollector exposes metrics in Prometheus exposition format:
//
//	collector := metrics.NewPrometheusCollector("myapp").
//		WithMaxMemory(50000)
//
//	// The handler produces Prometheus-compatible output
//	http.Handle("/metrics", collector.Handler())
//
// Metrics exposed:
//   - {namespace}_http_requests_total - Counter of HTTP requests
//   - {namespace}_http_request_duration_seconds_sum - Sum of request durations
//   - {namespace}_http_request_duration_seconds_count - Count of requests
//   - {namespace}_http_request_duration_seconds_min - Minimum request duration
//   - {namespace}_http_request_duration_seconds_max - Maximum request duration
//   - {namespace}_uptime_seconds - Application uptime
//   - {namespace}_http_requests_total_all - Total requests across all labels
//
// # OpenTelemetry Integration
//
// The OpenTelemetryTracer provides distributed tracing capabilities:
//
//	tracer := metrics.NewOpenTelemetryTracer("my-service")
//
//	// In HTTP handler
//	ctx, span := tracer.Start(context.Background(), r)
//	defer span.End(middleware.RequestMetrics{
//		Status:  w.Status(),
//		Bytes:   w.BytesWritten(),
//		TraceID: span.TraceID(),
//	})
//
//	// Retrieve spans for export
//	spans := tracer.Spans()
//	for _, span := range spans {
//		// Export to external system
//		exportSpan(span)
//	}
//
// # Memory Management
//
// Both collectors support memory limits to prevent unbounded growth:
//
//	// Limit Prometheus collector to 10,000 unique metric series
//	prom := metrics.NewPrometheusCollector("app").WithMaxMemory(10000)
//
//	// Limit base collector to 5,000 records
//	base := metrics.NewBaseMetricsCollector().WithMaxRecords(5000)
//
// When limits are exceeded, oldest or least-used entries are evicted.
//
// # Statistics and Monitoring
//
// All collectors provide statistics about collected metrics:
//
//	stats := collector.GetStats()
//	fmt.Printf("Total records: %d\n", stats.TotalRecords)
//	fmt.Printf("Error records: %d\n", stats.ErrorRecords)
//	fmt.Printf("Active series: %d\n", stats.ActiveSeries)
//
// For the tracer:
//
//	stats := tracer.GetSpanStats()
//	fmt.Printf("Total spans: %d\n", stats.TotalSpans)
//	fmt.Printf("Error spans: %d\n", stats.ErrorSpans)
//	fmt.Printf("Average duration: %v\n", stats.AverageDuration)
//
// # Clearing Metrics
//
// All collectors support clearing accumulated metrics:
//
//	collector.Clear() // Resets all metrics and statistics
//
// # Concurrency
//
// All collectors are safe for concurrent use from multiple goroutines.
// Internal synchronization is handled automatically.
//
// # Testing
//
// For testing, use either the NoopCollector or BaseMetricsCollector:
//
//	// No-op collector (discards all metrics)
//	collector := metrics.NewNoopCollector()
//
//	// Base collector (keeps metrics in memory for inspection)
//	collector := metrics.NewBaseMetricsCollector()
//	// ... perform operations ...
//	records := collector.GetRecords()
//	// Verify metrics were recorded correctly
//
// # Performance Considerations
//
//   - BaseMetricsCollector: Low overhead, suitable for testing and development
//   - PrometheusCollector: Optimized for Prometheus scraping, minimal allocations
//   - OpenTelemetryTracer: Adds span context overhead, use when tracing is needed
//   - NoopCollector: Zero overhead, all operations are no-ops
//
// # Integration with Plumego
//
// The metrics package integrates seamlessly with other Plumego components:
//
//	import (
//		"github.com/spcent/plumego/core"
//		"github.com/spcent/plumego/metrics"
//		"github.com/spcent/plumego/middleware"
//	)
//
//	app := core.New(
//		core.WithAddr(":8080"),
//	)
//
//	// Create and register metrics collector
//	collector := metrics.NewPrometheusCollector("myapp")
//	app.Get("/metrics", collector.Handler().ServeHTTP)
//
//	// Use in middleware
//	app.Use(middleware.Metrics(collector))
//
// # Best Practices
//
//  1. Use consistent metric naming across your application
//  2. Set reasonable memory limits to prevent unbounded growth
//  3. Clear metrics periodically in long-running applications
//  4. Use appropriate collector for your use case (Prometheus for metrics, OpenTelemetry for traces)
//  5. Include relevant labels but avoid high-cardinality labels (like user IDs)
//  6. Monitor collector statistics to ensure metrics are being recorded
//  7. Use NoopCollector in production when metrics are disabled
//
// # Examples
//
// See the examples directory for complete working examples:
//   - examples/reference/ - Full-featured application with metrics
//   - examples/docs/ - API documentation with metrics integration
package metrics
