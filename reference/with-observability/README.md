# with-observability

The canonical Plumego reference for **baseline production observability**. This
example demonstrates request logging, health probes, metrics, distributed tracing,
request IDs, and panic recovery — the essentials of operational visibility.

**Start here**: `reference/standard-service` first. This app extends it with two
capabilities and does not repeat the baseline explanation.

## What This Demonstrates

✓ **Health probes** — liveness (`/healthz`) and readiness (`/readyz`)  
✓ **Request IDs** — correlation IDs stamped on every request and in logs  
✓ **Request logging** — access logs with method, path, status, and latency  
✓ **Metrics collection** — HTTP counters and latency histograms  
✓ **Prometheus export** — Prometheus text format at `/metrics`  
✓ **Distributed tracing** — in-process span collection (W3C trace context)  
✓ **Panic recovery** — 500 responses instead of crashes  
✓ **Graceful shutdown** — drain in-flight requests on SIGTERM  

## What This Intentionally Excludes

✗ **OpenTelemetry exporter wiring** — spans are in-memory only (see production notes)  
✗ **Alerting or dashboards** — this is signal collection, not consumption  
✗ **Custom business metrics** — only HTTP transport metrics  
✗ **Distributed tracing backends** — Jaeger, Zipkin, or OTLP export  
✗ **Rate limiting or circuit breakers** — operational policies belong elsewhere  
✗ **Metrics authentication** — add your own gating middleware  

## Endpoint Reference

### Core
| Method | Path | Purpose | Status |
|---|---|---|---|
| `GET` | `/` | Service identity and docs link | 200 |
| `GET` | `/api/hello` | Service metadata | 200 |

### Health & Readiness
| Method | Path | Purpose | Status |
|---|---|---|---|
| `GET` | `/healthz` | Liveness probe (always 200 while running) | 200 |
| `GET` | `/readyz` | Readiness probe (checks registered components) | 200 or 503 |

### Observability
| Method | Path | Purpose | Status | Notes |
|---|---|---|---|---|
| `GET` | `/metrics` | Prometheus text exposition format | 200 | Gate in production |
| `GET` | `/api/v1/stats` | JSON metrics summary | 200 | JSON envelope |
| `GET` | `/api/v1/collector-stats` | Collector stats via interface | 200 | Demonstrates pattern |
| `GET` | `/api/v1/spans` | In-process tracing spans | 200 | **Remove in production** |

## Running

```bash
go run . -addr :8080
```

If `.env` exists, variables like `APP_ADDR` and `APP_METRICS_NAMESPACE` are loaded.
Environment variables override `.env`, and `-addr` flag overrides both.

```bash
# Health checks
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz

# Service info
curl http://localhost:8080/
curl http://localhost:8080/api/hello

# Metrics (Prometheus format)
curl http://localhost:8080/metrics

# Tracing and stats (JSON)
curl http://localhost:8080/api/v1/stats
curl http://localhost:8080/api/v1/spans?limit=5
```

### Example: Request Metrics

After a few requests, `GET /metrics` returns:

```
# HELP plumego_http_requests_total Total number of HTTP requests processed.
# TYPE plumego_http_requests_total counter
plumego_http_requests_total{method="GET",path="/api/hello",status="200"} 3
plumego_http_requests_total{method="GET",path="/metrics",status="200"} 1
# HELP plumego_http_request_duration_seconds HTTP request latency in seconds.
# TYPE plumego_http_request_duration_seconds summary
plumego_http_request_duration_seconds{method="GET",path="/api/hello",status="200",quantile="0.5"} 0.00025
plumego_http_request_duration_seconds{method="GET",path="/api/hello",status="200",quantile="0.95"} 0.0005
# HELP plumego_uptime_seconds Total uptime in seconds.
# TYPE plumego_uptime_seconds gauge
plumego_uptime_seconds 5.234
```

### Example: Span Inspection

`GET /api/v1/spans` returns:

```json
{
  "data": {
    "total": 3,
    "error_spans": 0,
    "spans": [
      {
        "trace_id": "6c5e8d4a9b2f1e3c",
        "span_id": "a7c9e2b1d6f4",
        "name": "GET /api/hello",
        "status": "ok",
        "duration_ms": 1,
        "timestamp": "2026-06-12T14:23:45Z",
        "attributes": {
          "http.method": "GET",
          "http.route": "/api/hello",
          "http.status_code": "200"
        }
      }
    ]
  }
}
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `APP_ADDR` | `:8080` | Listen address |
| `APP_SERVICE_NAME` | `plumego-observability` | Service identity in responses and metrics |
| `APP_METRICS_NAMESPACE` | `plumego` | Prometheus metric name prefix |
| `APP_METRICS_MAX_SERIES` | `10000` | Maximum distinct label combinations retained |
| `APP_MAX_BODY_BYTES` | `1048576` | Maximum request body; 0 disables the limit |
| `APP_TLS_ENABLED` | `false` | Enable TLS |
| `APP_TLS_CERT_FILE` | — | TLS certificate path (required when TLS enabled) |
| `APP_TLS_KEY_FILE` | — | TLS private key path (required when TLS enabled) |
| `APP_CORS_ALLOWED_ORIGINS` | `*` | Comma-separated CORS origins; leave empty for `*` |

## How It Works

### Health
`HealthHandler` serves `/healthz` (always 200) and `/readyz` (probes component
checkers). Checkers can be registered to verify database, cache, or external
service connectivity. See `internal/handler/health.go` for the pattern.

### Request ID
`middleware/requestid` stamps every request with a `request-id` header
(generated or propagated from inbound), visible in access logs and response
headers. This correlates logs and spans across services.

### Request Logging
`middleware/accesslog` logs every request/response with method, path, status,
latency, and request ID. Output format is JSON or plain text (configurable).

### Metrics
`httpmetrics.Middleware` wires a real `PrometheusCollector` (not noop). Every
request increments counters and records latency. `GET /metrics` exports
Prometheus text format, scrape-ready for Prometheus.

### Tracing
`middleware/tracing` wires an `OpenTelemetryTracer` that stores spans in-process.
Each request produces one span with HTTP attributes (method, route, status).
`GET /api/v1/spans` insposes recent spans for development. Replace with OTLP
for production.

### Recovery
`middleware/recovery` converts panics to 500 responses so the server stays up
and operators see the failure in logs.

### Graceful Shutdown
`main.run` catches `SIGTERM` and `SIGINT`, stops accepting new requests, drains
in-flight requests with a 15-second timeout, and exits cleanly.

## Production Checklist

- [ ] Remove `GET /api/v1/spans` endpoint (exposes internal trace state)
- [ ] Gate `GET /metrics` behind a bearer token or internal network (metric
      cardinality reveals endpoint structure)
- [ ] Replace `observability.NewOpenTelemetryTracer` with an OTLP gRPC exporter
      pointing at Jaeger, Zipkin, or your vendor (interface unchanged)
- [ ] Set `APP_METRICS_NAMESPACE` to your service name (avoid collisions when
      multiple services share Prometheus)
- [ ] Set `APP_CORS_ALLOWED_ORIGINS` to your frontend domain (default `*` is
      unsafe)
- [ ] Verify `/healthz` and `/readyz` routes are accessible to your load balancer
- [ ] Configure access log shipping if not stdout (see `middleware/accesslog`
      config)

## Files

| File | Role |
|---|---|
| `main.go` | Process entry point; signals and graceful shutdown |
| `internal/config/` | Configuration loading (env, flags, .env file) |
| `internal/app/app.go` | App construction; middleware wiring; tracer and collector setup |
| `internal/app/routes.go` | Route table |
| `internal/handler/` | HTTP handlers (health, metrics, tracing, API) |
