# Metrics、Tracing 与 Health 模块

Plumego 内置 Prometheus / OpenTelemetry 适配与健康探针，默认即具备可观察性。

## 指标与追踪
- 使用 `metrics.NewPrometheusCollector(namespace)` 构造采集器，通过 `collector.Handler()` 暴露在 `/metrics` 等路径。
- 通过 `metrics.NewOpenTelemetryTracer(serviceName)` 获取 Tracer，它实现了 `middleware.Tracer`，日志中间件会生成 Span。
- 在 `core.New` 时传入 `core.WithMetricsCollector(...)`、`core.WithTracer(...)` 完成注入。
- 当配置了 Tracer，日志会附带 TraceID，方便链路关联。

## 健康探针
- 使用 `health.ReadinessHandler()`、`health.BuildInfoHandler()` 挂载 `/health/ready` 与 `/health/build`。
- 就绪状态反映 `Boot()` 流程中的 `health.SetReady()` 调用；未就绪时返回 503。
- Build 信息返回 `health.BuildInfo` 元数据，可通过 `-ldflags` 注入版本号和提交信息。

## 组件健康
组件可实现 `Health() (name string, status health.HealthStatus)`，应用会汇总 `healthy`、`degraded`、`unhealthy` 状态，避免自定义字符串带来的歧义。

## 示例
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

## 运维建议
- 按环境添加标签或 namespace 区分仪表盘。
- 在就绪探针中加入数据库、缓存等依赖检查，并使用 context 超时避免探针阻塞。
- 结合延迟分布和状态码直方图设计 SLO/错误预算面板。

## 代码位置
- `metrics/prometheus.go`、`metrics/otel.go`：采集器与 Tracer 实现。
- `health/health.go`：健康处理器与状态定义。
- `examples/reference/main.go`：挂载指标与健康路由的示例。
