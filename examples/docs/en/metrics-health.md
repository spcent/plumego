# Metrics and Health modules

Plumego ships with Prometheus/OpenTelemetry adapters and lightweight health probes so you can wire observability without extra boilerplate.

## Metrics and tracing
- **Prometheus**: `metrics.NewPrometheusCollector(namespace)` implements `middleware.MetricsCollector`. Expose it with `app.GetHandler("/metrics", prom.Handler())`.
- **OpenTelemetry**: `metrics.NewOpenTelemetryTracer(serviceName)` implements `middleware.Tracer` and emits spans via the logging middleware.
- Inject either into `core.New` using `core.WithMetricsCollector` and `core.WithTracer`; logging middleware will automatically record durations, status codes, and trace IDs.

```go
prom := metrics.NewPrometheusCollector("plumego")
tracer := metrics.NewOpenTelemetryTracer("my-service")
app := core.New(core.WithMetricsCollector(prom), core.WithTracer(tracer), core.WithLogging())
app.GetHandler("/metrics", prom.Handler())
```

## Health endpoints
Two ready-to-serve handlers are available:

```go
app.GetHandler("/health/ready", health.ReadinessHandler())
app.GetHandler("/health/build", health.BuildInfoHandler())
```

- `ReadinessHandler` returns 200 after boot flips the ready flag; returns 503 during startup/shutdown.
- `BuildInfoHandler` surfaces `health.BuildInfo` (version, commit, build time) as JSON; set these fields via ldflags at build time.

## Health metrics
Attach a collector to a `HealthManager` to expose structured health metrics:

```go
manager, _ := health.NewHealthManager(health.HealthCheckConfig{})
collector := health.NewMetricsCollector(manager) // auto-attaches
app.GetHandler("/health/metrics", health.MetricsHandler(collector))
```

- If you construct a collector separately, call `health.AttachMetrics(manager, collector)` to wire it in.
- `HealthMetrics` includes `check_count`, `success_count`, `failure_count`, and per-component metrics in JSON.

## Component health reporting
Components can report structured health to feed readiness decisions or dashboards:

```go
func (w *worker) Health() (string, health.HealthStatus) {
    if w.backlog.Load() > 1000 {
        return "worker", health.Degraded
    }
    return "worker", health.Healthy
}
```

`HealthStatus` is type-safe (`Healthy`, `Degraded`, `Unhealthy`), making it easy to summarize component states.

## Operational tips
- Expose `/metrics` on an authenticated or internal path if your deployment requires it; the handler is plain `http.Handler` and can sit behind middleware.
- Keep readiness checks fast—avoid downstream calls or large allocations.
- Pair logging middleware with Prometheus/OTel collectors so every request gets correlated metrics and trace IDs automatically.

## Where to look in the repo
- `metrics/prometheus.go` and `metrics/otel.go`: collector and tracer adapters.
- `health/health.go`: readiness flag and status types; `health/http.go` for HTTP handlers.
- `examples/reference/main.go`: mounting `/metrics`, `/health/ready`, `/health/build` in a real app.
