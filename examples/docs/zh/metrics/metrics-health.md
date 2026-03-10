# Metrics 与 Health 模块

Plumego 提供 Prometheus / OpenTelemetry 适配器以及可直接挂载的健康检查端点。

## 指标与追踪
- `metrics.NewPrometheusCollector(namespace)` 提供 Prometheus 采集器与 `prom.Handler()`。
- `metrics.NewOpenTelemetryTracer(serviceName)` 提供与可观测性中间件兼容的 tracer。
- 通过 `core.WithMetricsCollector(...)` 与 `core.WithTracer(...)` 注入。

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

## 健康端点
```go
app.Get("/health/ready", health.ReadinessHandler().ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)
```

- `ReadinessHandler`：ready 后返回 `200`，启动/关闭过程中返回 `503`。
- `BuildInfoHandler`：返回构建信息 JSON（`version`、`commit`、`build_time`）。

## 组件健康上报
```go
func (w *worker) Health() (string, health.HealthStatus) {
    if w.backlog.Load() > 1000 {
        return "worker", health.HealthStatus{Status: health.StatusDegraded, Message: "backlog high"}
    }
    return "worker", health.HealthStatus{Status: health.StatusHealthy}
}
```

## 运维建议
- 就绪检查保持轻量、确定性强。
- `/metrics` 按部署要求放在内网或鉴权后路径。
- 建议日志 + 指标 + 追踪联合启用，便于请求级关联排障。
