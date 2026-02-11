# OpenTelemetry Adapter

> **Package**: `github.com/spcent/plumego/metrics/opentelemetry` | **Integration**: OpenTelemetry

## Overview

The OpenTelemetry adapter enables OTLP (OpenTelemetry Protocol) metrics export for Plumego applications. OpenTelemetry provides vendor-neutral observability with support for distributed tracing, metrics, and logs.

## Quick Start

### Basic Setup

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/metrics/opentelemetry"
)

// Configure OpenTelemetry exporter
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("localhost:4317"), // OTLP gRPC endpoint
    opentelemetry.WithServiceName("plumego-app"),
    opentelemetry.WithServiceVersion("1.0.0"),
)

app := core.New(
    core.WithMetricsExporter(otel),
)

app.Boot()
```

### With TLS

```go
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("otel-collector.example.com:4317"),
    opentelemetry.WithTLS(true),
    opentelemetry.WithServiceName("plumego-app"),
)
```

### With Headers (Authentication)

```go
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("api.honeycomb.io:443"),
    opentelemetry.WithHeaders(map[string]string{
        "x-honeycomb-team": "your-api-key",
    }),
    opentelemetry.WithServiceName("plumego-app"),
)
```

## Configuration Options

### Service Metadata

```go
otel := opentelemetry.New(
    opentelemetry.WithServiceName("payment-service"),
    opentelemetry.WithServiceVersion("2.1.3"),
    opentelemetry.WithServiceNamespace("production"),
    opentelemetry.WithResourceAttributes(map[string]string{
        "deployment.environment": "production",
        "service.instance.id":    "instance-1",
        "cloud.provider":         "aws",
        "cloud.region":           "us-east-1",
    }),
)
```

### Export Configuration

```go
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("localhost:4317"),
    opentelemetry.WithExportInterval(30 * time.Second),  // Export every 30s
    opentelemetry.WithTimeout(10 * time.Second),         // Timeout per export
    opentelemetry.WithBatchSize(512),                    // Batch size
    opentelemetry.WithCompression(true),                 // Enable compression
)
```

## OpenTelemetry Collector

### Collector Configuration

`otel-collector-config.yaml`:
```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 10s
    send_batch_size: 1024

exporters:
  # Export to Prometheus
  prometheus:
    endpoint: 0.0.0.0:8889

  # Export to Jaeger (for traces)
  jaeger:
    endpoint: jaeger:14250

  # Export to cloud provider
  otlp/cloudprovider:
    endpoint: api.cloudprovider.com:443
    headers:
      api-key: ${CLOUD_API_KEY}

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [prometheus, otlp/cloudprovider]
```

### Docker Compose

```yaml
version: '3'
services:
  app:
    build: .
    environment:
      - OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317
    depends_on:
      - otel-collector

  otel-collector:
    image: otel/opentelemetry-collector:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP
      - "8889:8889"   # Prometheus endpoint
```

## Backends

### Honeycomb

```go
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("api.honeycomb.io:443"),
    opentelemetry.WithTLS(true),
    opentelemetry.WithHeaders(map[string]string{
        "x-honeycomb-team": os.Getenv("HONEYCOMB_API_KEY"),
    }),
    opentelemetry.WithServiceName("plumego-app"),
)
```

### Datadog

```go
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("localhost:4317"), // Datadog Agent with OTLP
    opentelemetry.WithServiceName("plumego-app"),
    opentelemetry.WithResourceAttributes(map[string]string{
        "deployment.environment": "production",
    }),
)
```

### New Relic

```go
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("otlp.nr-data.net:4317"),
    opentelemetry.WithTLS(true),
    opentelemetry.WithHeaders(map[string]string{
        "api-key": os.Getenv("NEW_RELIC_LICENSE_KEY"),
    }),
    opentelemetry.WithServiceName("plumego-app"),
)
```

### Lightstep

```go
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("ingest.lightstep.com:443"),
    opentelemetry.WithTLS(true),
    opentelemetry.WithHeaders(map[string]string{
        "lightstep-access-token": os.Getenv("LIGHTSTEP_ACCESS_TOKEN"),
    }),
    opentelemetry.WithServiceName("plumego-app"),
)
```

## Metrics Mapping

### Counter → OTLP Sum

```go
counter := metrics.NewCounter("requests_total", "Total requests")
counter.Inc()

// Exported as:
// MetricDescriptor: requests_total
// DataType: Sum
// AggregationTemporality: Cumulative
```

### Gauge → OTLP Gauge

```go
gauge := metrics.NewGauge("active_connections", "Active connections")
gauge.Set(42)

// Exported as:
// MetricDescriptor: active_connections
// DataType: Gauge
```

### Histogram → OTLP Histogram

```go
histogram := metrics.NewHistogram("request_duration_seconds", "Request duration")
histogram.Observe(0.125)

// Exported as:
// MetricDescriptor: request_duration_seconds
// DataType: Histogram
// Buckets: [explicit bucket boundaries]
```

## Resource Attributes

### Standard Attributes

```go
otel := opentelemetry.New(
    opentelemetry.WithResourceAttributes(map[string]string{
        // Service
        "service.name":        "payment-service",
        "service.version":     "1.0.0",
        "service.namespace":   "production",
        "service.instance.id": "pod-xyz-123",

        // Deployment
        "deployment.environment": "production",

        // Cloud
        "cloud.provider":      "aws",
        "cloud.region":        "us-east-1",
        "cloud.availability_zone": "us-east-1a",

        // Host
        "host.name": "ip-10-0-1-123",
        "host.type": "t3.medium",

        // Container
        "container.name":  "plumego-app",
        "container.image": "plumego-app:v1.0.0",

        // Kubernetes
        "k8s.cluster.name":     "prod-cluster",
        "k8s.namespace.name":   "default",
        "k8s.pod.name":         "plumego-app-xyz-123",
        "k8s.deployment.name":  "plumego-app",
    }),
)
```

## Query Examples

### Honeycomb

```
# Average request duration
AVG(request_duration_seconds)

# P95 latency by endpoint
HEATMAP(request_duration_seconds) WHERE path = "/api/users"

# Error rate
COUNT(requests_total WHERE status >= 500) / COUNT(requests_total) * 100

# Requests per second
COUNT(requests_total) / 60
```

### Datadog

```
# Average latency
avg:request_duration_seconds{service:plumego-app}

# Request rate
sum:requests_total{service:plumego-app}.as_rate()

# Error rate
sum:requests_total{status:5xx,service:plumego-app}.as_count() / sum:requests_total{service:plumego-app}.as_count()
```

## Best Practices

### 1. Set Resource Attributes

Always include service identification:

```go
opentelemetry.WithServiceName("plumego-app"),
opentelemetry.WithServiceVersion("1.0.0"),
opentelemetry.WithServiceNamespace("production"),
```

### 2. Use Batch Export

```go
opentelemetry.WithBatchSize(512),
opentelemetry.WithExportInterval(30 * time.Second),
```

### 3. Enable Compression

```go
opentelemetry.WithCompression(true),
```

### 4. Handle Export Failures

```go
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("otel-collector:4317"),
    opentelemetry.WithRetry(true),
    opentelemetry.WithMaxRetries(3),
)
```

## Troubleshooting

### Metrics Not Appearing

**Check**:
1. Collector endpoint accessible
2. Authentication headers correct
3. Network connectivity
4. Collector logs for errors

### High Memory Usage

**Solutions**:
- Increase export interval
- Reduce batch size
- Enable compression

### Connection Errors

```go
// Add timeout and retry
otel := opentelemetry.New(
    opentelemetry.WithEndpoint("otel-collector:4317"),
    opentelemetry.WithTimeout(10 * time.Second),
    opentelemetry.WithRetry(true),
)
```

## Related Documentation

- [Metrics Overview](README.md) — Metrics module overview
- [Prometheus Adapter](prometheus.md) — Prometheus integration

## External Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OTLP Specification](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md)
- [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)
