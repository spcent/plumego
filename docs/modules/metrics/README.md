# Metrics Module

> **Package Path**: `github.com/spcent/plumego/metrics` | **Stability**: High | **Priority**: P1

## Overview

The `metrics/` package provides observability through metrics collection and export. It supports both Prometheus and OpenTelemetry standards, enabling monitoring, alerting, and performance analysis of Plumego applications.

**Key Features**:
- **Prometheus Integration**: Native Prometheus metrics and scraping endpoint
- **OpenTelemetry Support**: OTLP export for distributed systems
- **Standard Metrics**: Request counters, duration histograms, gauge metrics
- **Custom Metrics**: Define application-specific metrics
- **Auto-instrumentation**: Automatic HTTP request metrics
- **Multi-backend**: Support multiple exporters simultaneously

## Quick Start

### Prometheus Metrics

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/metrics"
)

// Enable Prometheus metrics
app := core.New(
    core.WithAddr(":8080"),
    core.WithPrometheusMetrics(true),
)

// Metrics endpoint: GET /metrics
app.Boot()
```

**Access metrics**:
```bash
curl http://localhost:8080/metrics
```

### Custom Metrics

```go
import "github.com/spcent/plumego/metrics"

// Define counter
var (
    requestsTotal = metrics.NewCounter("http_requests_total", "Total HTTP requests")
    requestDuration = metrics.NewHistogram("http_request_duration_seconds", "HTTP request duration")
    activeUsers = metrics.NewGauge("active_users", "Number of active users")
)

// Increment counter
requestsTotal.Inc()

// Observe duration
start := time.Now()
// ... handle request ...
requestDuration.Observe(time.Since(start).Seconds())

// Set gauge value
activeUsers.Set(float64(getUserCount()))
```

## Metric Types

### Counter

Monotonically increasing value (never decreases).

**Use cases**: Request counts, error counts, task completions

```go
counter := metrics.NewCounter("tasks_completed_total", "Total completed tasks")

// Increment by 1
counter.Inc()

// Increment by N
counter.Add(5)
```

**Example**:
```go
var (
    httpRequestsTotal = metrics.NewCounterVec(
        "http_requests_total",
        "Total HTTP requests",
        []string{"method", "path", "status"},
    )
)

// Track request
httpRequestsTotal.WithLabelValues("GET", "/api/users", "200").Inc()
```

### Gauge

Value that can go up or down.

**Use cases**: Active connections, memory usage, queue depth

```go
gauge := metrics.NewGauge("active_connections", "Number of active connections")

// Set value
gauge.Set(42)

// Increment
gauge.Inc()

// Decrement
gauge.Dec()

// Add delta
gauge.Add(10)
gauge.Sub(5)
```

**Example**:
```go
var activeUsers = metrics.NewGauge("active_users", "Current active users")

// Update periodically
go func() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        count := db.GetActiveUserCount()
        activeUsers.Set(float64(count))
    }
}()
```

### Histogram

Samples observations and counts them in configurable buckets.

**Use cases**: Request durations, response sizes, latencies

```go
histogram := metrics.NewHistogram(
    "http_request_duration_seconds",
    "HTTP request duration in seconds",
)

// Observe value
duration := time.Since(start).Seconds()
histogram.Observe(duration)
```

**Example with custom buckets**:
```go
histogram := metrics.NewHistogramWithBuckets(
    "api_latency_seconds",
    "API endpoint latency",
    []float64{0.001, 0.01, 0.1, 0.5, 1.0, 5.0}, // 1ms, 10ms, 100ms, 500ms, 1s, 5s
)

start := time.Now()
// ... handle API request ...
histogram.Observe(time.Since(start).Seconds())
```

### Summary

Similar to histogram but calculates quantiles (e.g., p50, p95, p99).

**Use cases**: Request latencies when you need percentiles

```go
summary := metrics.NewSummary(
    "request_latency_seconds",
    "Request latency in seconds",
)

summary.Observe(0.125) // 125ms
```

## Labels

Add dimensions to metrics for filtering and aggregation:

```go
// Counter with labels
requestsTotal := metrics.NewCounterVec(
    "http_requests_total",
    "Total HTTP requests",
    []string{"method", "path", "status"},
)

// Increment with specific labels
requestsTotal.WithLabelValues("GET", "/api/users", "200").Inc()
requestsTotal.WithLabelValues("POST", "/api/users", "201").Inc()
requestsTotal.WithLabelValues("GET", "/api/orders", "200").Inc()
```

**Query in Prometheus**:
```promql
# Total requests
sum(http_requests_total)

# Requests by method
sum by (method) (http_requests_total)

# 200 responses only
http_requests_total{status="200"}

# Rate of requests per second
rate(http_requests_total[5m])
```

## Standard HTTP Metrics

### Auto-instrumented Metrics

When Prometheus metrics are enabled, Plumego automatically tracks:

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `http_requests_total` | Counter | method, path, status | Total requests |
| `http_request_duration_seconds` | Histogram | method, path | Request duration |
| `http_request_size_bytes` | Histogram | method, path | Request body size |
| `http_response_size_bytes` | Histogram | method, path | Response body size |
| `http_requests_in_flight` | Gauge | - | Current active requests |

### Manual Instrumentation

```go
func instrumentedHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Track in-flight requests
        inFlightGauge.Inc()
        defer inFlightGauge.Dec()

        // Wrap response writer to capture status and size
        wrapped := &metricsResponseWriter{ResponseWriter: w, statusCode: 200}
        next.ServeHTTP(wrapped, r)

        // Record metrics
        duration := time.Since(start).Seconds()
        method := r.Method
        path := r.URL.Path
        status := fmt.Sprintf("%d", wrapped.statusCode)

        requestsTotal.WithLabelValues(method, path, status).Inc()
        requestDuration.WithLabelValues(method, path).Observe(duration)
        responseSize.WithLabelValues(method, path).Observe(float64(wrapped.bytesWritten))
    })
}
```

## Business Metrics

### Application-Specific Metrics

```go
var (
    // User metrics
    userRegistrations = metrics.NewCounter("user_registrations_total", "Total user registrations")
    userLogins = metrics.NewCounter("user_logins_total", "Total user logins")
    activeSubscriptions = metrics.NewGauge("active_subscriptions", "Current active subscriptions")

    // Order metrics
    ordersCreated = metrics.NewCounter("orders_created_total", "Total orders created")
    orderValue = metrics.NewHistogram("order_value_usd", "Order value in USD")
    paymentDuration = metrics.NewHistogram("payment_processing_duration_seconds", "Payment processing time")

    // Error metrics
    paymentFailures = metrics.NewCounterVec(
        "payment_failures_total",
        "Payment failures by reason",
        []string{"reason"},
    )
)

// Track business events
func handleOrderCreation(order *Order) {
    ordersCreated.Inc()
    orderValue.Observe(order.TotalUSD)

    start := time.Now()
    if err := processPayment(order); err != nil {
        paymentFailures.WithLabelValues(err.Error()).Inc()
        return
    }
    paymentDuration.Observe(time.Since(start).Seconds())
}
```

## Prometheus Integration

### Metrics Endpoint

```go
import "github.com/spcent/plumego/metrics/prometheus"

// Enable Prometheus endpoint
app := core.New(
    core.WithPrometheusMetrics(true),
)

// Metrics available at: GET /metrics
```

### Prometheus Configuration

`prometheus.yml`:
```yaml
scrape_configs:
  - job_name: 'plumego-app'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
```

### Common Queries

```promql
# Request rate (requests per second)
rate(http_requests_total[5m])

# Average response time
avg(rate(http_request_duration_seconds_sum[5m]))
/
avg(rate(http_request_duration_seconds_count[5m]))

# p95 latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error rate
sum(rate(http_requests_total{status=~"5.."}[5m]))
/
sum(rate(http_requests_total[5m]))

# Active users
active_users
```

## OpenTelemetry Integration

### Setup

```go
import "github.com/spcent/plumego/metrics/opentelemetry"

// Configure OpenTelemetry
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("otel-collector:4317"),
    opentelemetry.WithServiceName("plumego-app"),
    opentelemetry.WithServiceVersion("1.0.0"),
)

app := core.New(
    core.WithMetricsExporter(otel),
)
```

### OTLP Export

```go
// Metrics automatically exported to OpenTelemetry Collector
// Supports:
// - gRPC endpoint (default)
// - HTTP endpoint
// - Batch processing
// - Compression
```

## Dashboards

### Grafana Dashboard Example

```json
{
  "dashboard": {
    "title": "Plumego Application",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [{
          "expr": "rate(http_requests_total[5m])"
        }]
      },
      {
        "title": "Average Latency",
        "targets": [{
          "expr": "avg(rate(http_request_duration_seconds_sum[5m])) / avg(rate(http_request_duration_seconds_count[5m]))"
        }]
      },
      {
        "title": "Error Rate",
        "targets": [{
          "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m]))"
        }]
      }
    ]
  }
}
```

## Alerting

### Prometheus Alert Rules

`alerts.yml`:
```yaml
groups:
  - name: plumego_alerts
    rules:
      # High error rate
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m]))
          /
          sum(rate(http_requests_total[5m]))
          > 0.05
        for: 5m
        annotations:
          summary: "High error rate (> 5%)"

      # Slow responses
      - alert: SlowResponses
        expr: |
          histogram_quantile(0.95,
            rate(http_request_duration_seconds_bucket[5m])
          ) > 1.0
        for: 10m
        annotations:
          summary: "p95 latency > 1s"

      # High memory usage
      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes > 1e9
        for: 5m
        annotations:
          summary: "Memory usage > 1GB"
```

## Best Practices

### 1. Use Descriptive Names

```go
// ✅ Good: Clear, descriptive names
metrics.NewCounter("http_requests_total", "Total HTTP requests")
metrics.NewHistogram("db_query_duration_seconds", "Database query duration")

// ❌ Bad: Vague names
metrics.NewCounter("requests", "Requests")
metrics.NewHistogram("latency", "Latency")
```

### 2. Include Units in Names

```go
// ✅ Good: Units in name
"http_request_duration_seconds"
"response_size_bytes"
"cache_hit_ratio" // unitless ratio

// ❌ Bad: No units
"request_time"
"response_size"
```

### 3. Limit Label Cardinality

```go
// ✅ Good: Low cardinality labels
requestsTotal.WithLabelValues("GET", "/api/users", "200")

// ❌ Bad: High cardinality (creates too many time series)
requestsTotal.WithLabelValues("GET", "/api/users/123456789", "200") // user ID in label!
```

### 4. Use Histograms for Latency

```go
// ✅ Good: Histogram for duration
duration := metrics.NewHistogram("request_duration_seconds", "Request duration")
duration.Observe(time.Since(start).Seconds())

// ❌ Bad: Gauge for duration (loses distribution)
duration := metrics.NewGauge("request_duration_seconds", "Request duration")
duration.Set(time.Since(start).Seconds())
```

## Module Documentation

- **[Prometheus Adapter](prometheus.md)** — Prometheus integration details
- **[OpenTelemetry Adapter](opentelemetry.md)** — OpenTelemetry integration details

## Related Documentation

- [Log Module](../log/) — Structured logging
- [Middleware: Observability](../middleware/observability.md) — Request tracing
- [Health Module](../health/) — Health checks

## Reference Implementation

See examples:
- `examples/reference/` — Production metrics setup
- `examples/api-gateway/` — Gateway metrics patterns

---

**Next Steps**:
1. Read [Prometheus Adapter](prometheus.md) for Prometheus setup
2. Read [OpenTelemetry Adapter](opentelemetry.md) for OTLP export
3. Set up Grafana dashboards for visualization
4. Configure Prometheus alerts for your SLOs
