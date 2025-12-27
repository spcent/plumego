# Metrics, tracing, and health modules

Plumego ships with Prometheus/OpenTelemetry adapters and health endpoints so observability is on by default.

## Metrics and tracing
- Create a collector with `metrics.NewPrometheusCollector(namespace)`; expose it via `collector.Handler()` (usually on `/metrics`).
- Add tracing with `metrics.NewOpenTelemetryTracer(serviceName)`; it implements `middleware.Tracer` so the logging middleware can emit spans.
- Pass the collector/tracer into `core.New` using `core.WithMetricsCollector(...)` and `core.WithTracer(...)`.
- Logs include trace IDs when a tracer is configured, simplifying correlation.

## Health endpoints
- Use `health.ReadinessHandler()` and `health.BuildInfoHandler()` to expose `/health/ready` and `/health/build` routes.
- Readiness reflects the `health.SetReady()` lifecycle hook triggered during `Boot()`; returns 503 until ready.
- Build info returns version metadata populated via `health.BuildInfo` (can be set through `-ldflags`).

## Component health integration
Components can implement `Health() (name string, status health.HealthStatus)` so the app aggregates readiness/degradation states. Use typed statuses (`healthy`, `degraded`, `unhealthy`) instead of ad-hoc strings.

## Example
```go
collector := metrics.NewPrometheusCollector("plumego_example")
tracer := metrics.NewOpenTelemetryTracer("plumego_example")
app := core.New(
    core.WithMetricsCollector(collector),
    core.WithTracer(tracer),
)
app.Get("/metrics", collector.Handler().ServeHTTP)
app.Get("/health/ready", health.ReadinessHandler().ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)
```

## Operational tips
- Add labels or namespaces per deployment environment to separate dashboards.
- Extend readiness checks to downstream dependencies (DB, cache) and respect context deadlines to avoid probe timeouts.
- Combine latency and status-code histograms to craft SLO/error-budget panels.

## Where to look in the repo
- `metrics/prometheus.go` and `metrics/otel.go` for collectors and tracers.
- `health/health.go` for readiness/build handlers and status definitions.
- `examples/reference/main.go` for mounting metrics and health routes in a live app.
