# Metrics and Health modules

Plumego provides Prometheus/OpenTelemetry adapters and ready-to-use health endpoints.

## Metrics and tracing
- `metrics.NewPrometheusCollector(namespace)` provides a Prometheus collector and `prom.Handler()`.
- `metrics.NewOpenTelemetryTracer(serviceName)` provides a tracer compatible with observability middleware.
- Inject hooks with `core.WithMetricsCollector(...)` and `core.WithTracer(...)`.

```go
prom := metrics.NewPrometheusCollector("plumego")
tracer := metrics.NewOpenTelemetryTracer("my-service")

app := core.New(
    core.WithMetricsCollector(prom),
    core.WithTracer(tracer),
)

if err := app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), prom, tracer),
); err != nil {
    log.Fatal(err)
}

app.Get("/metrics", prom.Handler().ServeHTTP)
```

## Health endpoints
```go
app.Get("/health/ready", health.ReadinessHandler().ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)
```

- `ReadinessHandler`: returns `200` after ready, `503` during startup/shutdown.
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
