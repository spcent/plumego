# with-observability

The canonical Plumego reference for **baseline production observability**. It
demonstrates the minimum wiring for operational visibility: health probes,
readiness probes, request IDs, request logging, panic recovery, HTTP metrics,
and Prometheus exposition — without introducing external platforms, dashboards,
or tracing backends.

**Start here**: `reference/standard-service` first. This app extends it with
observability wiring and does not repeat the baseline explanation.

## What This Demonstrates

- Health endpoint — `/healthz` (liveness probe, always 200 while the process serves HTTP)
- Readiness endpoint — `/readyz` (component health checks, 200 or 503)
- Request IDs — correlation IDs stamped on every request and propagated in logs and response headers
- Request logging — structured access logs with method, path, status, latency, and request ID
- Panic recovery — 500 responses instead of server crashes; process continues serving
- Metrics collection — HTTP counters and latency histograms via `PrometheusCollector`
- Prometheus export — Prometheus text format at `/metrics`, scrape-ready
- Distributed tracing — in-process span collection (W3C trace context; development only)
- Graceful shutdown — drains in-flight requests on SIGTERM before exiting

## What This Intentionally Excludes

- OpenTelemetry exporter wiring (spans are in-memory only; see production notes)
- Alerting or dashboards (signal collection, not consumption)
- Custom business metrics (only HTTP transport metrics are shown)
- Distributed tracing backends — Jaeger, Zipkin, or OTLP export
- Rate limiting or circuit breakers
- Metrics endpoint authentication (add your own gating middleware)

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
| `GET` | `/api/v1/collector-stats` | Collector stats via `metrics.StatsReader` interface | 200 | Demonstrates pattern |
| `GET` | `/api/v1/spans` | In-process tracing spans | 200 | **Remove in production** |

## Running

```bash
go run . -addr :8080
```

If `.env` exists, variables like `APP_ADDR` and `APP_METRICS_NAMESPACE` are
loaded first. Environment variables override `.env`; the `-addr` flag overrides
both.

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

### Example: Liveness and Readiness

```bash
$ curl -s http://localhost:8080/healthz | jq .
{
  "data": {
    "status": "ok",
    "service": "plumego-observability",
    "timestamp": "2026-06-14T10:00:00Z"
  }
}

$ curl -s http://localhost:8080/readyz | jq .
{
  "data": {
    "ready": true,
    "timestamp": "2026-06-14T10:00:00Z",
    "components": {}
  }
}
```

When a component checker fails, `/readyz` returns 503:

```bash
$ curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/readyz
503
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

### Example: Access Log (Request Logging)

Each request produces a structured access log line (text format shown):

```
time=2026-06-14T10:00:01Z level=INFO msg="request" method=GET path=/api/hello status=200 latency_ms=1 request_id=01HWXYZ123ABC
```

The `request_id` field matches the `X-Request-ID` response header, allowing
correlation across logs, spans, and upstream systems.

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
        "timestamp": "2026-06-14T10:00:01Z",
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

### Health vs Readiness

These are two distinct signals with different meanings for orchestration:

**`/healthz` — Liveness probe**: "Is the process alive and serving HTTP?" Always
returns 200 while the server is running. Orchestrators (Kubernetes, ECS) restart
the container only when liveness fails. `HealthHandler.Live` returns 200
unconditionally; it does not probe dependencies.

**`/readyz` — Readiness probe**: "Is the service ready to accept traffic?" Returns
200 when all registered `health.ComponentChecker` instances pass; returns 503 when
any fail. Orchestrators remove the pod from the load-balancer pool on readiness
failure without restarting it. Register checkers (database, cache, external APIs)
when constructing `HealthHandler`.

This distinction is load-bearing: a database timeout should make the pod
unready (stop sending traffic), not dead (trigger a restart). Only wire
checkers that gate traffic acceptance into `/readyz`.

### Request ID

`middleware/requestid` stamps every request with an `X-Request-ID` header.
If the inbound request carries `X-Request-ID`, that value is preserved;
otherwise a new ID is generated. The ID is visible in access logs and response
headers, making it the primary correlation handle across services.

### Request Logging

`middleware/accesslog` logs every request/response with method, path, status,
latency, and request ID. Output format is JSON or plain text (configurable via
`LoggerConfig`). The logger is constructed in `app.New` and shared across all
middleware and handlers.

### Panic Recovery

`middleware/recovery` catches panics in downstream handlers and returns a 500
response instead of crashing the process. Panic metadata is logged server-side;
no internal details leak to the client. The server continues serving all
subsequent requests normally after a recovered panic.

### Metrics

`httpmetrics.Middleware` wires a real `PrometheusCollector` (not noop). Every
request increments counters and records latency into labeled histograms. `GET
/metrics` exports Prometheus text format, scrape-ready for Prometheus server.
`APP_METRICS_NAMESPACE` prefixes all metric names.

### Tracing

`middleware/tracing` wires an `OpenTelemetryTracer` that stores spans
in-process. Each request produces one span with HTTP attributes (method, route,
status, request ID). `GET /api/v1/spans` exposes recent spans for development
inspection without requiring a tracing backend. Replace with an OTLP exporter
for production.

### Graceful Shutdown

`main.run` catches `SIGTERM` and `SIGINT`, stops accepting new connections, drains
in-flight requests with a 15-second timeout, and exits cleanly. Neither
`PrometheusCollector` nor `OpenTelemetryTracer` hold background goroutines;
shutdown only requires draining the HTTP server.

## Production Checklist

- [ ] Remove `GET /api/v1/spans` endpoint (exposes internal trace state)
- [ ] Gate `GET /metrics` behind a bearer token or internal network (metric
      cardinality reveals endpoint structure)
- [ ] Replace `observability.NewOpenTelemetryTracer` with an OTLP gRPC exporter
      pointing at Jaeger, Zipkin, or your vendor (interface unchanged)
- [ ] Set `APP_METRICS_NAMESPACE` to your service name (avoid collisions when
      multiple services share Prometheus)
- [ ] Set `APP_CORS_ALLOWED_ORIGINS` to your frontend domain (default `*` is unsafe)
- [ ] Verify `/healthz` and `/readyz` routes are accessible to your load balancer
- [ ] Configure access log shipping if not stdout (see `middleware/accesslog` config)

## How to Extend

### Add a component health check

Pass a `health.ComponentChecker` when building `HealthHandler` in
`internal/app/routes.go`. The checker is called on every `/readyz` request.

```go
health := handler.HealthHandler{
    ServiceName: a.Cfg.App.ServiceName,
    Logger:      a.Core.Logger(),
    Checkers: []health.ComponentChecker{
        myDatabaseChecker,  // implement health.ComponentChecker
    },
}
```

### Add a custom metrics inspection endpoint

Add a method to `ObservabilityHandler` in `internal/handler/api.go`:

```go
func (h ObservabilityHandler) MyStats(w http.ResponseWriter, r *http.Request) {
    logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, myData, nil))
}
```

Register it in `internal/app/routes.go` under the `v1` group:

```go
v1.get("/my-stats", http.HandlerFunc(obs.MyStats))
```

### Replace the in-process tracer with an OTLP exporter

In `internal/app/app.go`, replace:

```go
tracer := observability.NewOpenTelemetryTracer(cfg.App.ServiceName)
```

with an OTLP gRPC exporter. The `mwtracing.Middleware` interface is unchanged.
Remove `GET /api/v1/spans` from `routes.go` when switching to an external backend.

### Gate `/metrics` behind a bearer token

Wrap the exporter handler at registration time in `internal/app/routes.go`:

```go
root.get("/metrics", metricsAuthMiddleware(exporter.Handler()))
```

Do not add auth inside the handler — keep the middleware boundary explicit.

## Related References

| Reference | When to use it |
|---|---|
| `reference/standard-service` | Baseline JSON API without observability wiring |
| `reference/with-ops` | Health and metrics endpoints only (no tracing) |
| `reference/production-service` | Full auth, tracing, metrics, and tenant isolation |
| `reference/with-ai` | LLM/AI wiring on the same observable foundation |

## Files

| File | Role |
|---|---|
| `main.go` | Process entry point; signals and graceful shutdown |
| `internal/config/` | Configuration loading (env, flags, .env file) |
| `internal/app/app.go` | App construction; middleware wiring; tracer and collector setup |
| `internal/app/routes.go` | Route table |
| `internal/handler/health.go` | Liveness and readiness handlers |
| `internal/handler/api.go` | Identity, metrics stats, and span inspection handlers |
| `internal/handler/write.go` | `logWriteErr` helper |
