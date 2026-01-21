# Middleware 模块

Plumego 的中间件与标准 `http.Handler` 兼容，可通过 `app.Use(...)` 全局注册，也可在分组 `Group("/x", m1, m2)` 或具体路由上包裹使用。

## 内置中间件
- **恢复**：`core.WithRecovery()` 捕获 panic，记录栈并返回结构化错误。
- **日志**：`core.WithLogging()` 采集请求/响应信息，与 `core.WithMetricsCollector`、`core.WithTracer` 注入的指标/追踪对接。
- **CORS**：`core.WithCORS()` 提供宽松默认值，可用 `core.WithCORSOptions(...)` 或 `middleware.CORSWithOptions(...)` 自定义。
- **Gzip**：`middleware.Gzip()` 在客户端声明 `Accept-Encoding: gzip` 时压缩响应。
- **超时**：`middleware.Timeout(duration)` 为单个请求设定截止时间。
- **请求体限制**：`middleware.BodyLimit(maxBytes, logger)` 返回结构化的 413 响应。
- **并发限制**：`middleware.ConcurrencyLimit(maxConcurrent, queueDepth, queueTimeout, logger)` 控制并发与排队深度。
- **限流**：`middleware.RateLimit(ratePerSecond, burst, cleanupInterval, maxIdle)` 基于 IP 的令牌桶实现。
- **鉴权辅助**：`middleware.SimpleAuth("token")` 校验 Bearer Token；`middleware.APIKey("X-API-Key", "secret")` 校验自定义头。

核心在初始化时会自动接入保护性的请求体/并发限制，可按需通过 `app.Use(...)` 或分组中间件追加链路。

## 组合示例
### 全局防护链
```go
app := core.New(core.WithRecovery(), core.WithLogging())
app.Use(
    middleware.Gzip(),
    middleware.Timeout(3*time.Second),
    middleware.ConcurrencyLimit(200, 400, 250*time.Millisecond, logger),
)
```

### 路由级控制
```go
secured := app.Router().Group("/admin", middleware.SimpleAuth(os.Getenv("AUTH_TOKEN")))
secured.Get("/stats", middleware.TimeoutFunc(1*time.Second)(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("ok"))
}))
```

### 使用 `contract.Ctx` 辅助
```go
app.GetCtx("/echo/:msg", middleware.WrapCtx(middleware.Timeout(2*time.Second), func(ctx *contract.Ctx) {
    ctx.JSON(http.StatusOK, map[string]any{"echo": ctx.Param("msg")})
}))
```

## 运维提示
- 链路顺序很重要：通常建议日志包裹恢复，以便 panic 日志包含请求上下文。
- 自定义中间件避免使用全局可变状态；依赖通过闭包或应用级单例传递。
- 当在分组上启用 CORS/鉴权时，将公共静态资源置于独立分组以避免意外头部或鉴权校验。

## 代码位置
- `middleware/` 目录：恢复、日志、gzip、CORS、超时、体积限制、限流、鉴权、并发控制等实现。
- `examples/reference/main.go`：结合恢复、日志、CORS 与默认安全限制的实际链路。
