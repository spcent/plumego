# Core 模块

**core** 负责构建 HTTP 应用的全生命周期，将路由、中间件、组件启动/停止、WebSocket、Webhook 以及可观察性统一到 `core.App` 中。

## 职责
- 使用 `core.New` 结合函数式选项创建应用（监听地址、调试开关、指标采集器、Tracer、Pub/Sub 总线、WebSocket 配置、Webhook 配置）。
- 在 `Boot()` 之前注册路由和中间件；启动时会冻结路由表，防止遗漏注册。
- 通过 `Component` 接口驱动生命周期：`RegisterRoutes`、`RegisterMiddleware`、`Start`、`Stop`、`Health`。
- 管理环境加载（`core.WithEnvPath`）、优雅退出、TLS/HTTP2 开关，以及默认限制（请求体、并发、排队深度、HTTP 超时）。

## 最小示例
```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
)
app.EnableRecovery()
app.EnableLogging()
app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
    w.Write([]byte("pong"))
})
if err := app.Boot(); err != nil {
    log.Fatalf("server stopped: %v", err)
}
```

## 配置清单
- **监听与退出**：用 `core.WithAddr` 覆盖默认 `:8080`；优雅退出会在 SIGTERM 时等待正在处理的请求完成，超时由 `ShutdownTimeout` 控制。
- **限制**：默认提供 10 MiB 体积、256 并发、排队深度和 HTTP 读写超时。可通过 `core.WithMaxBodyBytes`、`core.WithMaxConcurrency`、`core.WithHTTPTimeouts` 或 `env.example` 中的环境变量覆盖。
- **TLS/HTTP2**：通过 `core.WithTLSConfig` 或 `AppConfig.TLS` 启用；若需关闭 HTTP/2，可使用 `core.WithHTTP2(false)`。
- **组件**：使用 `core.WithComponent(...)` 复用能力，内置组件涵盖 Webhook 服务、Pub/Sub 调试页、WebSocket Hub、前端静态资源。
- **可观察性**：将 `metrics.NewPrometheusCollector(namespace)`、`metrics.NewOpenTelemetryTracer(name)` 注入 `core.New`，日志中间件会自动输出指标和 TraceID。

## 启动与安全注意
- 只在完成路由与中间件注册后调用 `Boot()`；启动后新增注册会 panic，用于暴露初始化顺序问题。
- 将长生命周期的 goroutine（Webhook 调度、Pub/Sub 分发）封装为组件，并通过 `Start/Stop` 与应用生命周期对齐。
- 生产环境关闭 `WithDebug()`，日志级别交由外部日志系统管理。

## 排查清单
- **端口冲突**：检查 `WithAddr` 设置，在容器内用 `netstat -tnlp` 诊断占用。
- **密钥缺失**：Webhook 签名报错多因 `GITHUB_WEBHOOK_SECRET`、`STRIPE_WEBHOOK_SECRET`、`WS_SECRET` 未加载；确认 `.env` 生效。
- **退出缓慢**：确保 Handler 遵循 `context` 取消；若滚动升级卡住，可适当降低 `ShutdownTimeout`。

## 代码位置
- `core/app.go`：`App` 结构、选项与生命周期逻辑。
- `core/options.go`：地址、调试、TLS、Webhook、WebSocket、Pub/Sub 等配置选项。
- `examples/reference/main.go`：演示大部分核心能力的完整接线。
