# with-observability

Extends `standard-service` with production-grade HTTP metrics and distributed
tracing. Read `standard-service` first — this app adds exactly two capabilities
and does not repeat the baseline explanation.

## What this app adds

| What | Package | Route |
|---|---|---|
| Prometheus HTTP metrics | `x/observability.PrometheusCollector` | `GET /metrics` |
| OpenTelemetry trace spans | `x/observability.OpenTelemetryTracer` | `GET /api/v1/spans` |
| JSON metrics summary | `x/observability.PrometheusCollector.GetStats()` | `GET /api/v1/stats` |
| Interface-based stats | `metrics.StatsReader` | `GET /api/v1/collector-stats` |

## How it works

### Metrics

`app.go` creates a `PrometheusCollector` and wires it into
`httpmetrics.Middleware` instead of the noop collector used in
`standard-service`. Every HTTP request automatically increments the counter and
records latency. `GET /metrics` returns the Prometheus text exposition format
ready for Prometheus to scrape.

```
# HELP plumego_http_requests_total Total number of HTTP requests processed.
# TYPE plumego_http_requests_total counter
plumego_http_requests_total{method="GET",path="/api/hello",status="200"} 42
...
# HELP plumego_uptime_seconds Total uptime in seconds.
plumego_uptime_seconds 120.003
```

### Tracing

`app.go` creates an `OpenTelemetryTracer` and wires it into
`middleware/tracing.Middleware`. Each request produces one span carrying HTTP
attributes (`http.method`, `http.route`, `http.status_code`, etc.).

In this reference the tracer stores spans in memory. For production, replace
`OpenTelemetryTracer` with an OTLP gRPC exporter — the `middleware/tracing`
interface is the same, so only the constructor changes.

`GET /api/v1/spans?limit=10` returns the 10 most recent spans as JSON so you
can inspect trace IDs and latencies without a tracing backend during
development. Remove this endpoint in production.

### Interface pattern

`MetricsHandler.CollectorStats` is wired through `metrics.StatsReader` rather
than the concrete `*PrometheusCollector`. This pattern lets you swap the backing
collector without touching handler code — any `AggregateCollector` (noop,
`BaseMetricsCollector`, `MultiCollector`) satisfies the interface.

## Running

```bash
go run . -addr :8080
```

Then:

```bash
curl localhost:8080/metrics          # Prometheus text format
curl localhost:8080/api/v1/spans     # OpenTelemetry spans (JSON)
curl localhost:8080/api/v1/stats     # Metrics summary (JSON)
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `APP_ADDR` | `:8080` | Listen address |
| `APP_SERVICE_NAME` | `plumego-observability` | Service identity |
| `APP_METRICS_NAMESPACE` | `plumego` | Prometheus metric name prefix |
| `APP_METRICS_MAX_SERIES` | `10000` | Maximum distinct label combinations retained |
| `APP_MAX_BODY_BYTES` | `1048576` | Maximum request body (0 disables) |
| `APP_TLS_ENABLED` | `false` | Enable TLS |
| `APP_TLS_CERT_FILE` | — | TLS certificate path (required when TLS enabled) |
| `APP_TLS_KEY_FILE` | — | TLS private key path (required when TLS enabled) |

## Production notes

- Gate `GET /metrics` behind an internal network or bearer-token middleware;
  metric cardinality can reveal endpoint structure.
- Remove `GET /api/v1/spans`; it exposes request internals.
- Replace `OpenTelemetryTracer` with an OTLP exporter pointing at Jaeger,
  Zipkin, or your vendor's collector.
- Set `APP_METRICS_NAMESPACE` to your service name to avoid metric collisions
  when multiple services share a Prometheus instance.

## Key files

| File | Role |
|---|---|
| `internal/app/app.go` | Constructs `PrometheusCollector` and `OpenTelemetryTracer`, wires both into middleware |
| `internal/app/routes.go` | Registers `/metrics` with `PrometheusExporter.Handler()`, inspection routes |
| `internal/handler/api.go` | `ObservabilityHandler` (stats, spans), `MetricsHandler` (StatsReader demo) |
