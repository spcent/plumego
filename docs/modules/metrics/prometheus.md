# Prometheus Adapter

> **Package**: `github.com/spcent/plumego/metrics/prometheus` | **Integration**: Prometheus

## Overview

The Prometheus adapter enables native Prometheus metrics collection and exposition in Plumego applications. It provides automatic HTTP instrumentation and a `/metrics` endpoint for Prometheus scraping.

## Quick Start

### Enable Prometheus Metrics

```go
import "github.com/spcent/plumego/core"

app := core.New(
    core.WithAddr(":8080"),
    core.WithPrometheusMetrics(true),
)

app.Boot()
```

**Metrics endpoint**: `http://localhost:8080/metrics`

### Custom Metrics Path

```go
app := core.New(
    core.WithPrometheusMetrics(true),
    core.WithPrometheusPath("/internal/metrics"), // Custom path
)
```

## Metrics Format

### Example Output

```
# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/users",status="200"} 1543

# HELP http_request_duration_seconds HTTP request duration
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.005"} 523
http_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.01"} 892
http_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.025"} 1234
http_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.05"} 1456
http_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.1"} 1520
http_request_duration_seconds_bucket{method="GET",path="/api/users",le="+Inf"} 1543
http_request_duration_seconds_sum{method="GET",path="/api/users"} 28.456
http_request_duration_seconds_count{method="GET",path="/api/users"} 1543

# HELP http_requests_in_flight Current active HTTP requests
# TYPE http_requests_in_flight gauge
http_requests_in_flight 5
```

## Prometheus Server Configuration

### Scrape Configuration

`prometheus.yml`:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'plumego'
    static_configs:
      - targets: ['localhost:8080']
        labels:
          environment: 'production'
          service: 'api'
```

### Service Discovery

**Kubernetes**:
```yaml
scrape_configs:
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
```

## Query Examples

### Request Rate

```promql
# Requests per second (5-minute average)
rate(http_requests_total[5m])

# By endpoint
sum by (path) (rate(http_requests_total[5m]))

# By method
sum by (method) (rate(http_requests_total[5m]))
```

### Latency

```promql
# Average latency
rate(http_request_duration_seconds_sum[5m])
/
rate(http_request_duration_seconds_count[5m])

# p50 (median)
histogram_quantile(0.5, rate(http_request_duration_seconds_bucket[5m]))

# p95
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# p99
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
```

### Error Rate

```promql
# Overall error rate
sum(rate(http_requests_total{status=~"5.."}[5m]))
/
sum(rate(http_requests_total[5m]))

# By endpoint
sum by (path) (rate(http_requests_total{status=~"5.."}[5m]))
/
sum by (path) (rate(http_requests_total[5m]))
```

### Traffic Patterns

```promql
# Total requests
sum(http_requests_total)

# Requests by status code
sum by (status) (http_requests_total)

# Top 5 endpoints by traffic
topk(5, sum by (path) (rate(http_requests_total[5m])))
```

## Grafana Integration

### Dashboard JSON

```json
{
  "dashboard": {
    "title": "Plumego Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total[5m]))",
            "legendFormat": "Total RPS"
          }
        ]
      },
      {
        "title": "Latency Percentiles",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "p50"
          },
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "p95"
          },
          {
            "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "p99"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m]))",
            "legendFormat": "5xx Error Rate"
          }
        ]
      }
    ]
  }
}
```

## Alert Rules

`rules/alerts.yml`:
```yaml
groups:
  - name: plumego_alerts
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m]))
          /
          sum(rate(http_requests_total[5m]))
          > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High 5xx error rate ({{ $value | humanizePercentage }})"
          description: "Error rate is above 5% for the last 5 minutes"

      - alert: HighLatency
        expr: |
          histogram_quantile(0.95,
            rate(http_request_duration_seconds_bucket[5m])
          ) > 1.0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High p95 latency ({{ $value }}s)"
          description: "p95 latency is above 1s for the last 10 minutes"

      - alert: ServiceDown
        expr: up{job="plumego"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service is down"
          description: "Prometheus cannot scrape metrics endpoint"
```

## Recording Rules

`rules/recording.yml`:
```yaml
groups:
  - name: plumego_recording
    interval: 30s
    rules:
      # Pre-compute request rate
      - record: job:http_requests:rate5m
        expr: sum by (job) (rate(http_requests_total[5m]))

      # Pre-compute p95 latency
      - record: job:http_request_duration:p95
        expr: histogram_quantile(0.95, sum by (job, le) (rate(http_request_duration_seconds_bucket[5m])))

      # Pre-compute error rate
      - record: job:http_errors:rate5m
        expr: sum by (job) (rate(http_requests_total{status=~"5.."}[5m]))
```

## Best Practices

### 1. Metric Naming

Follow Prometheus conventions:
- Use `_total` suffix for counters
- Use `_seconds` for durations
- Use `_bytes` for sizes
- Use base units (seconds, bytes, not ms or KB)

### 2. Label Cardinality

```go
// ✅ Good: Low cardinality
http_requests_total{method="GET", path="/api/users", status="200"}

// ❌ Bad: High cardinality (user ID in label)
http_requests_total{user_id="123456789"}
```

### 3. Histogram Buckets

Choose buckets appropriate for your SLOs:

```go
// API latency (milliseconds to seconds)
[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}

// Response size (bytes to megabytes)
[]float64{100, 1000, 10000, 100000, 1000000, 10000000}
```

## Troubleshooting

### Metrics Not Appearing

**Check**:
1. Prometheus metrics enabled: `core.WithPrometheusMetrics(true)`
2. Endpoint accessible: `curl http://localhost:8080/metrics`
3. Prometheus configuration correct
4. Network connectivity

### High Memory Usage

**Causes**:
- Too many time series (high label cardinality)
- Unbounded label values

**Solutions**:
- Limit label cardinality
- Use recording rules to pre-aggregate
- Increase Prometheus memory

## Related Documentation

- [Metrics Overview](README.md) — Metrics module overview
- [OpenTelemetry Adapter](opentelemetry.md) — OTLP export

## External Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
- [PromQL Cheat Sheet](https://promlabs.com/promql-cheat-sheet/)
