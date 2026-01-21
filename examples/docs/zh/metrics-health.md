# Metrics 与 Health 模块

Plumego 自带 Prometheus / OpenTelemetry 适配器和轻量健康探针，免去额外样板代码。

## 指标与追踪
- **Prometheus**：`metrics.NewPrometheusCollector(namespace)` 实现 `middleware.MetricsCollector`，通过 `app.GetHandler("/metrics", prom.Handler())` 暴露。
- **OpenTelemetry**：`metrics.NewOpenTelemetryTracer(serviceName)` 实现 `middleware.Tracer`，由日志中间件自动产生日志与 span。
- 通过 `core.WithMetricsCollector`、`core.WithTracer` 注入后，日志中间件会自动记录耗时、状态码与 TraceID。

```go
prom := metrics.NewPrometheusCollector("plumego")
tracer := metrics.NewOpenTelemetryTracer("my-service")
app := core.New(core.WithMetricsCollector(prom), core.WithTracer(tracer), core.WithLogging())
app.GetHandler("/metrics", prom.Handler())
```

## 健康端点
提供开箱即用的处理器：

```go
app.GetHandler("/health/ready", health.ReadinessHandler())
app.GetHandler("/health/build", health.BuildInfoHandler())
```

- `ReadinessHandler` 在启动完成后返回 200，启动/关闭阶段返回 503。
- `BuildInfoHandler` 以 JSON 形式暴露 `health.BuildInfo`（版本、提交、构建时间）；可通过 ldflags 注入。

## 组件健康上报
组件可报告结构化健康状态，便于就绪决策与看板展示：

```go
func (w *worker) Health() (string, health.HealthStatus) {
    if w.backlog.Load() > 1000 {
        return "worker", health.Degraded
    }
    return "worker", health.Healthy
}
```

`HealthStatus` 具备类型安全（`Healthy`、`Degraded`、`Unhealthy`），方便聚合各组件状态。

## 运维提示
- 若 `/metrics` 需要认证或只在内网暴露，可配合中间件或独立监听路径；处理器本身是标准 `http.Handler`。
- 保持就绪检查轻量，避免下游调用或大规模分配。
- 日志中间件与 Prometheus/OTel 采集器组合使用，可为每个请求自动关联指标与 TraceID。

## 代码位置
- `metrics/prometheus.go` 与 `metrics/otel.go`：指标采集器与追踪适配器。
- `health/health.go`：就绪标记与状态类型；`health/http.go` 提供 HTTP 处理器。
- `examples/reference/main.go`：在真实应用中挂载 `/metrics`、`/health/ready`、`/health/build` 的示例。
