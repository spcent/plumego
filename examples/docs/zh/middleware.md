# Middleware 模块

**middleware** 提供恢复、结构化日志、CORS、压缩、超时、限流/限并发、请求体限制、认证助手以及指标/追踪适配器等跨切面能力。

## 职责
- 以 `func(http.Handler) http.Handler` 包装标准 Handler；通过 `app.Use(...)` 全局注册，或在路由/分组上追加。
- 提供 `middleware.MetricsCollector` 与 `middleware.Tracer` 接口，用于接入 Prometheus / OpenTelemetry。
- 通过 `core.With...` 选项与 `app.EnableRecovery()`、`app.EnableLogging()` 等方式启用常用中间件。

## 内置能力
- **恢复 & 日志**：`app.EnableRecovery()`、`app.EnableLogging()`；当配置了 Tracer 时日志会带 TraceID。
- **CORS**：`app.EnableCORS()` 默认开启安全的跨域策略，可按需自定义。
- **防护**：并发/队列限制、超时、Gzip、体积限制、简单 Token 鉴权（见 `middleware/simple_auth.go`）。
- **指标**：通过 `core.WithMetricsCollector`、`core.WithTracer` 注入 Prometheus 采集器和 OpenTelemetry Tracer。

## 组合规则
- 顺序敏感：恢复/日志适合放在最外层；超时、鉴权等靠近业务 Handler。
- 分组中间件会被子路由继承，优先在边界分组（如 `/api/admin`）堆叠安全策略。
- 避免可变全局变量，使用闭包或构造函数传递配置。

## 自定义模板
```go
func NewMaskingLogger() middleware.Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 掩码敏感头
            r.Header.Del("Authorization")
            next.ServeHTTP(w, r)
        })
    }
}
```
通过 `app.Use(NewMaskingLogger())` 或分组注册即可生效。

## 排查
- **重复写响应**：遵循 `http.ResponseWriter` 语义，避免在 `next.ServeHTTP` 返回后再次写入已结束的响应。
- **超时未生效**：确保超时中间件位于耗时 Handler 之前，并且 Handler 全程传播 `context`。
- **缺少指标**：确认在 `core.New` 时传入了采集器/Tracer；日志中间件只有在有 Tracer 时才会发出 Span。

## 代码位置
- `middleware` 目录：恢复、日志、cors、gzip、timeout、限流、鉴权等实现。
- `examples/reference/main.go`：演示默认中间件栈的实战用法。
