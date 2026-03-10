# Core 模块

`core` 包负责 HTTP 服务生命周期，围绕标准 `net/http` 处理器提供路由、中间件注册、组件启停和优雅退出能力。

## 创建并启动应用
```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
)

if err := app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), nil, nil),
    recovery.RecoveryMiddleware,
    cors.CORS,
); err != nil {
    log.Fatalf("register middleware: %v", err)
}

app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
    w.Write([]byte("pong"))
})

if err := app.Boot(); err != nil {
    log.Fatalf("server stopped: %v", err)
}
```

## 配置与默认值
主要配置通过 `AppConfig` / `core.With...` 完成：

- `WithAddr`、`WithEnvPath`
- `WithShutdownTimeout`
- `WithServerTimeouts`（读/读头/写/空闲超时）
- `WithMaxHeaderBytes`
- `WithHTTP2`、`WithTLS`、`WithTLSConfig`
- `WithDebug`、`WithLogger`
- `WithMethodNotAllowed`

## 组件与生命周期
`core.App` 支持组件编排，使后台任务与服务启停严格对齐。

```go
type worker struct{ bus *pubsub.InProcPubSub }

func (w *worker) RegisterRoutes(r *router.Router)           {}
func (w *worker) RegisterMiddleware(m *middleware.Registry) {}
func (w *worker) Start(ctx context.Context) error           { return nil }
func (w *worker) Stop(ctx context.Context) error            { return nil }
func (w *worker) Health() (string, health.HealthStatus) {
    return "worker", health.HealthStatus{Status: health.StatusHealthy}
}

app := core.New(core.WithComponent(&worker{bus: pubsub.New()}))
```

后台任务和关闭钩子可通过 `WithRunner` / `app.Register(...)`、`WithShutdownHook` / `app.OnShutdown(...)` 接入。

## 参考接线示例
```go
prom := metrics.NewPrometheusCollector("plumego")
tracer := metrics.NewOpenTelemetryTracer("plumego")
healthManager, err := health.NewHealthManager(health.HealthCheckConfig{})
if err != nil {
    log.Fatal(err)
}

app := core.New(
    core.WithAddr(":8080"),
    core.WithMetricsCollector(prom),
    core.WithTracer(tracer),
    core.WithHealthManager(healthManager),
)

if err := app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), prom, tracer),
    recovery.RecoveryMiddleware,
); err != nil {
    log.Fatal(err)
}

app.Get("/metrics", prom.Handler().ServeHTTP)
app.Get("/health/ready", health.ReadinessHandler(healthManager).ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)

if err := app.Boot(); err != nil {
    log.Fatal(err)
}
```

## Ctx 风格处理器（显式适配）
`core.App` 的标准签名仍是 `func(http.ResponseWriter, *http.Request)`。如需使用 `contract.Ctx`，请显式适配：

```go
app.Post("/users", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
    var req CreateUserRequest
    if err := ctx.BindAndValidateJSONWithOptions(&req, contract.BindOptions{}); err != nil {
        contract.WriteBindError(ctx.W, ctx.R, err)
        return
    }
    _ = ctx.Response(http.StatusOK, map[string]any{"ok": true}, nil)
}, app.Logger()).ServeHTTP)
```

## 安全与排障提示
- `Boot()` 前完成路由和中间件注册；启动后变更会被拒绝。
- 处理器应支持 `context` 取消，保证优雅退出可控。
- WebSocket/Webhook 建议通过显式组件和配置接入，避免隐藏副作用。
