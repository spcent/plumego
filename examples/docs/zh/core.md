# Core 模块

**core** 管理 HTTP 服务器的完整生命周期：路由、中间件、组件启动/停止、WebSocket 辅助、webhook 服务、可观察性与优雅退出，全部基于 Go 标准库实现。

## 创建并启动应用
```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
)
app.EnableRecovery()
app.EnableLogging()
app.EnableCORS()

// 在 Boot 冻结路由前完成注册。
app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
    w.Write([]byte("pong"))
})

if err := app.Boot(); err != nil {
    log.Fatalf("server stopped: %v", err)
}
```

## 配置与默认值
大部分开关可在 `env.example` 找到对应环境变量。

- **监听与退出**：`core.WithAddr` 覆盖默认 `:8080`；`ShutdownTimeout`（默认 5s）在 SIGTERM 时等待处理中的请求完成。
- **限制**：默认 10 MiB 请求体、256 并发带排队、HTTP 读写超时。通过 `core.WithMaxBodyBytes`、`core.WithMaxConcurrency`、`core.WithHTTPTimeouts` 或环境变量覆盖。
- **TLS / HTTP2**：`core.WithTLSConfig` 或 `AppConfig.TLS` 启用；`core.WithHTTP2(false)` 可关闭 HTTP/2。
- **环境加载**：`core.WithEnvPath` 控制 `.env` 路径。
- **可观察性**：注入 `metrics.NewPrometheusCollector` 与 `metrics.NewOpenTelemetryTracer`，日志中间件会自动上报指标与追踪。
- **调试**：`WithDebug()` 会输出注册的路由并放宽部分失败保护，生产环境建议关闭。

## 组件与生命周期
`core.App` 通过组件保持长生命周期任务与服务器启动/关闭对齐。

```go
type worker struct{ bus *pubsub.PubSub }
func (w *worker) RegisterRoutes(r *router.Router)          {}
func (w *worker) RegisterMiddleware(m *middleware.Registry) {}
func (w *worker) Start(ctx context.Context) error {
    return w.bus.Subscribe("jobs.*", func(ctx context.Context, evt pubsub.Event) error {
        // ...处理事件...
        return nil
    })
}
func (w *worker) Stop(ctx context.Context) error { return nil }
func (w *worker) Health() (string, health.HealthStatus) { return "worker", health.Healthy }

app := core.New(core.WithComponent(&worker{bus: pubsub.New()}))
```

内置组件涵盖入站 webhook 接收、出站 webhook 服务、WebSocket Hub 注册、Pub/Sub 调试页以及前端挂载。通过 `core.WithComponent` 或诸如 `core.WithWebhookIn`、`core.WithWebhookOut`、`core.ConfigureWebSocketWithOptions` 等专用选项加入。

## 参考式接线示例
```go
bus := pubsub.New()
prom := metrics.NewPrometheusCollector("plumego")
tracer := metrics.NewOpenTelemetryTracer("plumego")

app := core.New(
    core.WithAddr(":8080"),
    core.WithPubSub(bus),
    core.WithMetricsCollector(prom),
    core.WithTracer(tracer),
    core.WithWebhookIn(core.WebhookInConfig{Enabled: true, Pub: bus}),
    core.WithWebhookOut(core.WebhookOutConfig{Enabled: true, TriggerToken: "secret"}),
)
app.EnableRecovery()
app.EnableLogging()

app.GetHandler("/metrics", prom.Handler())
app.GetHandler("/health/ready", health.ReadinessHandler())
app.GetHandler("/health/build", health.BuildInfoHandler())
_, _ = app.ConfigureWebSocketWithOptions(core.DefaultWebSocketConfig())

if err := app.Boot(); err != nil { log.Fatal(err) }
```

## 安全与排障提示
- `Boot()` 之后新增路由/中间件会 panic，这是有意的初始化顺序保护，请提前注册。
- 处理函数应支持 `context` 取消，以便优雅退出更快；若编排工具要求更短/更长耗时，可调整 `ShutdownTimeout`。
- 启动报错多由 WebSocket JWT 密钥或 webhook 密钥缺失导致，确认 `.env` 已加载。

## 代码位置
- `core/app.go`：`App` 结构、选项与生命周期逻辑。
- `core/options.go`：地址、调试、TLS、webhook、WebSocket、Pub/Sub 配置。
- `examples/reference/main.go`：涵盖大部分核心能力的完整接线示例。
