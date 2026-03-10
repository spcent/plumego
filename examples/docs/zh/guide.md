# Plumego — 基于 Go 标准库的 Web 工具包

Plumego 是围绕 `net/http` 构建的轻量 Go HTTP 工具包，强调显式路由、中间件组合、生命周期安全，以及按需挂载能力组件（webhook、websocket、可观测性、多租户等）。

## 亮点
- 路由支持分组、参数、路由元信息与逆向路由。
- 中间件链显式可控（`func(http.Handler) http.Handler`）。
- 组件/runner 生命周期管理与优雅关闭。
- 可选 Prometheus / OpenTelemetry 适配器。
- 全面保持标准库兼容。

## 组件模型
`core.App` 通过可插拔组件编排能力：

```go
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
}
```

通过 `core.WithComponent(...)` / `core.WithComponents(...)` 挂载组件。

## 快速开始
```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware/cors"
    "github.com/spcent/plumego/middleware/observability"
    "github.com/spcent/plumego/middleware/recovery"
)

func main() {
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
}
```

## 配置基础
- `.env` 加载：`core.WithEnvPath(...)`。
- 地址、HTTP 超时、优雅关闭、header 限制、TLS、HTTP2 等通过 `core.With...` 选项配置。
- 指标/链路追踪通过 `core.WithMetricsCollector(...)` 与 `core.WithTracer(...)` 注入。

## 关键运行时能力
- 路由：`app.Get/Post/Put/Delete/Patch/Any`；高级能力通过 `app.Router()`。
- 中间件：`app.Use(...)` 显式注册；`group.Use(...)` 做分组级控制。
- Contract：`contract.WriteError`、`contract.WriteResponse`、`contract.AdaptCtxHandler`。
- WebSocket：`app.ConfigureWebSocket()` / `app.ConfigureWebSocketWithOptions(...)`。
- 健康检查：`health.ReadinessHandler(manager)` 与 `health.BuildInfoHandler()`。

## 参考应用
`examples/reference` 提供接近生产的组合示例：
- Request ID + logging + recovery + CORS 中间件链。
- 可选 metrics/tracing 注入。
- 显式 WebSocket 配置。
- 通过 `core.WithComponent(...)` 挂载 webhook 组件。

运行：

```bash
go run ./examples/reference
```

## 开发检查
- `go test -timeout 20s ./...`
- `go vet ./...`
- `gofmt -w .`
