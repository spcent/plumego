# Metrics and Health modules

Plumego provides Prometheus/OpenTelemetry adapters and ready-to-use health endpoints.

## Metrics and tracing
- `metrics.NewPrometheusCollector(namespace)` provides the collector.
- `metrics.NewPrometheusExporter(prom)` provides the `/metrics` HTTP exporter.
- `metrics.NewOpenTelemetryTracer(serviceName)` provides a tracer compatible with observability middleware.
- Inject hooks with `core.WithMetricsCollector(...)` and `core.WithTracer(...)`.

```go
prom := metrics.NewPrometheusCollector("plumego")
exporter := metrics.NewPrometheusExporter(prom)
tracer := metrics.NewOpenTelemetryTracer("my-service")

app := core.New(
    core.WithMetricsCollector(prom),
    core.WithTracer(tracer),
)

if err := app.Use(
    observability.RequestID(),
    observability.Tracing(tracer),
    observability.HTTPMetrics(prom),
    observability.AccessLog(app.Logger()),
); err != nil {
    log.Fatal(err)
}

app.Get("/metrics", exporter.Handler().ServeHTTP)
```

## Health endpoints
```go
healthManager, err := health.NewHealthManager(health.HealthCheckConfig{})
if err != nil {
    log.Fatal(err)
}

app := core.New(core.WithHealthManager(healthManager))
app.Get("/health", health.SummaryHandler(healthManager).ServeHTTP)
app.Get("/health/ready", health.ReadinessHandler(healthManager).ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)
```

- `ReadinessHandler`: returns readiness from `healthManager` (`200` when ready, `503` when not ready).
- `BuildInfoHandler`: returns build metadata JSON (`version`, `commit`, `build_time`).

## Component health reporting
```go
func (w *worker) Health() (string, health.HealthStatus) {
    if w.backlog.Load() > 1000 {
        return "worker", health.HealthStatus{Status: health.StatusDegraded, Message: "backlog high"}
    }
    return "worker", health.HealthStatus{Status: health.StatusHealthy}
}
```

## Operational notes
- Keep readiness checks fast and deterministic.
- Place `/metrics` behind network or auth boundaries if needed.
- Run logging + metrics + tracing together for request-level correlation.
