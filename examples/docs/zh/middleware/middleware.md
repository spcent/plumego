# Middleware 模块

Plumego 中间件完全兼容标准库签名（`func(http.Handler) http.Handler`）。
可用 `app.Use(...)` 全局注册，也可通过 `group.Use(...)` 作用于路由分组。

## 常用内置中间件
- Request ID：`observability.RequestID()`
- 结构化日志：`observability.Logging(app.Logger(), metricsCollector, tracer)`
- panic 恢复：`recovery.RecoveryMiddleware`
- CORS：`cors.CORS` / `cors.CORSWithOptions(...)`
- Gzip 压缩：`compression.Gzip()`
- 超时控制：`timeout.Timeout(duration)`
- 请求体限制：`limits.BodyLimit(maxBytes, logger)`
- 并发限制：`limits.ConcurrencyLimit(maxConcurrent, queueDepth, queueTimeout, logger)`
- 鉴权辅助：`auth.SimpleAuth(token)`
- 安全头：`security.SecurityHeaders(nil)`
- 防滥用/令牌桶：`ratelimit.AbuseGuard(...)`、`ratelimit.TokenBucket(...)`

## 全局链路示例
```go
app := core.New(core.WithAddr(":8080"))

if err := app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), nil, nil),
    recovery.RecoveryMiddleware,
    cors.CORS,
    compression.Gzip(),
    timeout.Timeout(3*time.Second),
    limits.ConcurrencyLimit(200, 400, 250*time.Millisecond, app.Logger()),
); err != nil {
    log.Fatal(err)
}
```

## 分组级控制
```go
api := app.Router().Group("/api")
api.Use(auth.SimpleAuth(os.Getenv("AUTH_TOKEN")))
api.Use(timeout.Timeout(2 * time.Second))

api.Get("/stats", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("ok"))
}))
```

## Ctx 绑定/校验（显式适配）
```go
app.Post("/v1/users", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
    var payload CreateUserRequest
    if err := ctx.BindAndValidateJSONWithOptions(&payload, contract.BindOptions{
        DisallowUnknownFields: true,
        Logger:                ctx.Logger,
    }); err != nil {
        contract.WriteBindError(ctx.W, ctx.R, err)
        return
    }

    _ = ctx.Response(http.StatusCreated, map[string]any{"ok": true}, nil)
}, app.Logger()).ServeHTTP)
```

## 运维建议
- 保持中间件顺序显式，并用回归测试锁定顺序语义。
- 依赖通过构造函数或闭包注入，避免隐藏全局状态。
- 公私路由混合时，优先按分组挂鉴权/安全中间件，降低误配风险。
