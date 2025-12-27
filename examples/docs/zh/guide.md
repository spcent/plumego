# Plumego 文档索引（中文）

本目录与源码结构一致，每个能力都有独立的、可直接复制的指南。阅读此页可以快速了解 Plumego 提供的全部功能以及如何在你的 `main` 包中接入。

## 你将获得
- 100% 标准库构建的 HTTP 生命周期（基于 `net/http`、`context`）。
- 前缀树路由，支持分组、路径参数和中间件继承。
- 内置防护中间件（恢复、日志、gzip、CORS、超时、请求体/并发限制、限流、简易鉴权）。
- 带 JWT 鉴权的 WebSocket Hub，可选广播端点。
- 进程内 Pub/Sub 总线，可用于 webhook 事件和 WebSocket 消息扇出。
- 入站（GitHub/Stripe）与出站 webhook 服务，支持触发令牌与重试。
- 指标（Prometheus）、追踪（OpenTelemetry）、就绪/构建健康检查。
- 从磁盘或嵌入资源挂载前端静态站点。

## 参考示例一键运行
运行演示应用即可看到所有能力的集成效果：

```bash
go run ./examples/reference
```

服务监听 `:8080`，同时提供本目录文档、`/metrics`、`/health/ready`、`/health/build`、`/ws` WebSocket Hub，并将入/出站 webhook 通过 Pub/Sub 打通。

## 构建你的应用
在调用 `Boot()` 前完成路由注册：

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware"
)

func main() {
    app := core.New(core.WithAddr(":8080"))
    app.EnableRecovery()
    app.EnableLogging()

    app.Use(middleware.Gzip())
    app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte("pong"))
    })

    if err := app.Boot(); err != nil {
        log.Fatalf("server stopped: %v", err)
    }
}
```

需要与参考示例一致的全量能力？添加 `core.WithPubSub`、`core.ConfigureWebSocketWithOptions`、webhook 配置以及下文模块页提到的指标/健康处理器即可。

## 模块索引
- [Core](core.md)：生命周期、配置、组件、优雅退出。
- [Router](router.md)：分组、路径参数、Context Handler、静态前端挂载。
- [Middleware](middleware.md)：内置防护与组合实践。
- [Pub/Sub & WebSocket](pubsub-websocket.md)：实时消息、Hub 接线与扇出示例。
- [Metrics & Health](metrics-health.md)：Prometheus/OpenTelemetry、就绪/构建探针。
- [Security & Webhook](security-webhook.md)：入站签名校验、出站投递管理与密钥说明。

每个页面都包含可直接运行的代码片段，可复制到 `examples/reference` 或你的服务中验证。
