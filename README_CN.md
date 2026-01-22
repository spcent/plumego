# Plumego — 仅基于golang标准库的 Web 工具包

Plumego 是一个小型 Go HTTP 工具包，完全基于标准库实现，同时覆盖路由、中间件、优雅关闭、WebSocket 辅助工具、Webhook 管道以及静态前端托管。它设计为嵌入到你自己的 `main` 包中，而不是作为一个独立的框架二进制文件运行。

## 亮点
- **路由器支持分组和参数**：基于 Trie 的匹配器，支持 `/:param` 段、路由冻结，以及每路由/分组的中件栈。
- **中间件链**：日志、恢复、gzip、CORS、超时（默认缓冲上限 10 MiB）、限流、并发限制、请求体大小限制、安全头，以及认证辅助工具，全部包装标准 `http.Handler`。
- **安全辅助**：JWT + 密码工具、安全头策略、输入安全校验与基础防滥用组件，便于进行安全基线加固。
- **集成扩展**：提供 `database/sql`、Redis 缓存与消息队列的轻量适配器/扩展点。
- **结构化日志钩子**：接入自定义日志器，并通过中间件钩子收集指标/链路追踪。
- **优雅生命周期**：环境变量加载、连接排水、就绪标志，以及可选的 TLS/HTTP2 配置，带有合理默认值。
- **可选服务**：内置带认证的 WebSocket 中心、进程内 Pub/Sub（带调试快照）、入站/出站 Webhook 路由器，以及从磁盘或嵌入资源提供静态前端。

## 组件
`core.App` 通过可插拔组件进行编排，而不是硬编码功能。组件可以注册路由、中间件和生命周期钩子：

```
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
}
```

`HealthStatus` 使用限定的状态值（`healthy`、`degraded`、`unhealthy`）确保组件以结构化且类型安全的方式报告健康状况。

在构造应用时使用 `core.WithComponent`（或 `WithComponents`）来添加功能。内置特性（Webhook 管理、入站 Webhook 接收器、PubSub 调试、WebSocket 辅助工具、前端服务）都可以作为组件挂载，因此示例可以只混合所需的部分。

## 快速开始
创建一个小型 `main.go`，连接路由和中间件，然后启动服务器：

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithDebug(),
        core.WithRecovery(),
        core.WithLogging(),
        core.WithCORS(),
    )

    app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte("pong"))
    })

    if err := app.Boot(); err != nil {
        log.Fatalf("server stopped: %v", err)
    }
}
```

`plumego.App` 也实现了 `http.Handler`，可以直接挂载到标准库的服务器中。上下文处理器可以使用统一的 `plumego.Context`：

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego"
)

func main() {
    app := plumego.New()

    app.GetCtx("/health", func(ctx *plumego.Context) {
        ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
    })

    log.Println("server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", app))
}
```

## 配置基础
- 环境变量可以从 `.env` 文件加载（默认路径 `.env`；可通过 `core.WithEnvPath` 覆盖）。
- 常用变量：`AUTH_TOKEN`（SimpleAuth 中间件）、`WS_SECRET`（WebSocket JWT 签名密钥，至少 32 字节）、`WEBHOOK_TRIGGER_TOKEN`、`GITHUB_WEBHOOK_SECRET` 和 `STRIPE_WEBHOOK_SECRET`（详见 `env.example`）。
- 应用默认包括 10 MiB 请求体限制、256 并发请求限制（带队列）、HTTP 读/写超时，以及 5 秒优雅关闭窗口。可通过 `core.With...` 选项覆盖。
- 安全基线默认启用（安全头 + 防滥用中间件）。防滥用默认每客户端 100 req/s，突发 200，并最多跟踪 10 万个活跃 key。可通过 `core.WithSecurityHeadersEnabled`、`core.WithSecurityHeadersPolicy`、`core.WithAbuseGuardEnabled`、`core.WithAbuseGuardConfig` 关闭或调整。
- Debug 模式（`core.WithDebug`）默认开启 `/_debug` 调试端点（路由表、Middleware、配置快照、手动重载）、友好 JSON 错误输出，以及 `.env` 热加载。

## 关键组件
- **路由器**：使用 `Get`、`Post` 等注册处理器，或上下文感知变体（`GetCtx`），后者暴露统一的请求上下文包装器。分组允许附加共享中间件，静态前端可以通过 `frontend.RegisterFromDir` 挂载，并支持缓存/回退选项（`frontend.WithCacheControl`、`frontend.WithIndexCacheControl`、`frontend.WithFallback`、`frontend.WithHeaders`）。
- **中间件**：在启动前使用 `app.Use(...)` 链式添加中间件；防护栏（安全头、防滥用、请求体限制、并发限制）会在设置期间自动注入。恢复/日志/CORS 辅助工具可通过 `core.WithRecovery`、`core.WithLogging`、`core.WithCORS` 启用。
- **WebSocket 中心**：`ConfigureWebSocket()` 挂载受 JWT 保护的 `/ws` 端点，以及可选的广播端点（受共享密钥保护）。通过 `WebSocketConfig` 自定义工作线程数和队列大小。
- **Pub/Sub + Webhook**：提供 `pubsub.PubSub` 实现以启用 Webhook 分发。出站 Webhook 管理包括目标 CRUD、交付重放和触发令牌；入站接收器处理 GitHub/Stripe 签名，带去重和大小限制。
- **健康检查 + 就绪**：生命周期钩子在启动/关闭期间标记就绪状态，构建元数据（`Version`、`Commit`、`BuildTime`）可通过 ldflags 注入。

## 参考应用
`examples/reference` 是一个开箱即用的 `main` 包，整合了常用组件：

- 配置了带 JWT 密钥的 WebSocket 中心和广播端点
- 入站 GitHub/Stripe Webhook，发布到进程内 Pub/Sub
- 基于内存存储的出站 Webhook 管理
- 从嵌入资源提供静态前端
- Prometheus 指标、OpenTelemetry 链路追踪，以及健康端点挂载到路由器

运行方式：

```bash
go run ./examples/reference
```

## 健康端点
`health` 包现在暴露 HTTP 处理程序，无需自行实现就绪/构建信息检查：

```go
app.Get("/health/ready", health.ReadinessHandler().ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)
```

`ReadinessHandler` 在 `health.SetReady()` 被调用后返回 200（启动生命周期会自动调用），否则返回 503。`BuildInfoHandler` 以 JSON 形式返回当前的 `health.BuildInfo` 结构体。

## 可观测性适配器
无需自行编写适配器，即可将日志中间件接入指标/链路追踪后端：

- `metrics.NewPrometheusCollector(namespace)` 实现 `middleware.MetricsCollector`，并通过 `collector.Handler()` 暴露 `/metrics` 处理程序。
- `metrics.NewOpenTelemetryTracer(name)` 实现 `middleware.Tracer`，发出带有 HTTP 元数据的 span。

如 `examples/reference` 所示，使用 `core.WithMetricsCollector(...)` 和 `core.WithTracer(...)` 将它们接入 `core.New`。

如果希望一键启用 Prometheus 指标与 OpenTelemetry 风格追踪：

```go
obs := core.DefaultObservabilityConfig()
obs.Metrics.Enabled = true
obs.Tracing.Enabled = true

if err := app.ConfigureObservability(obs); err != nil {
    log.Fatal(err)
}
```

开启追踪后日志会包含 `trace_id` 与 `span_id`，响应中也会回传 `X-Span-ID` 便于关联。

## 配置参考
使用 `config.LoadEnv` 加载环境变量，或绑定命令行标志；`config.ConfigManager` 也提供 `LoadBestEffort` 用于跳过可选配置源失败，并提供 `ReloadWithValidation` 做事务式热加载；使用下表实现可预测的部署。

| AppConfig 字段             | 默认值          | 环境变量                       | Flag 示例                          |
|----------------------------|-----------------|--------------------------------|------------------------------------|
| Addr                       | :8080          | APP_ADDR                      | --addr :8080                      |
| EnvFile                    | .env           | APP_ENV_FILE                  | --env-file .env                   |
| Debug                      | false          | APP_DEBUG                     | --debug                           |
| ShutdownTimeout            | 5s             | APP_SHUTDOWN_TIMEOUT_MS       | --shutdown-timeout 5s             |
| ReadTimeout                | 30s            | APP_READ_TIMEOUT_MS           | --read-timeout 30s                |
| ReadHeaderTimeout          | 5s             | APP_READ_HEADER_TIMEOUT_MS    | --read-header-timeout 5s          |
| WriteTimeout               | 30s            | APP_WRITE_TIMEOUT_MS          | --write-timeout 30s               |
| IdleTimeout                | 60s            | APP_IDLE_TIMEOUT_MS           | --idle-timeout 60s                |
| MaxHeaderBytes             | 1 MiB          | APP_MAX_HEADER_BYTES          | --max-header-bytes 1048576        |
| EnableHTTP2                | true           | APP_ENABLE_HTTP2              | --http2=false                     |
| DrainInterval              | 500ms          | APP_DRAIN_INTERVAL_MS         | --drain-interval 500ms            |
| MaxBodyBytes               | 10 MiB         | APP_MAX_BODY_BYTES            | --max-body-bytes 10485760         |
| MaxConcurrency             | 256            | APP_MAX_CONCURRENCY           | --max-concurrency 256             |
| QueueDepth                 | 512            | APP_QUEUE_DEPTH               | --queue-depth 512                 |
| QueueTimeout               | 250ms          | APP_QUEUE_TIMEOUT_MS          | --queue-timeout 250ms             |
| TLS.Enabled                | false          | TLS_ENABLED                   | --tls                             |
| TLS.CertFile               | (empty)        | TLS_CERT_FILE                 | --tls-cert /path/cert.pem         |
| TLS.KeyFile                | (empty)        | TLS_KEY_FILE                  | --tls-key /path/key.pem           |
| PubSub.Enabled             | false          | PUBSUB_DEBUG_ENABLED          | --pubsub-debug                    |
| PubSub.Path                | /_debug/pubsub | PUBSUB_DEBUG_PATH             | --pubsub-path /_debug/pubsub      |
| WebhookOut.TriggerToken    | (empty)        | WEBHOOK_TRIGGER_TOKEN         | --webhook-trigger-token TOKEN     |
| WebhookOut.BasePath        | /webhooks      | (inherit)                     | --webhook-base /webhooks          |
| WebhookOut.IncludeStats    | false          | WEBHOOK_INCLUDE_STATS         | --webhook-include-stats           |
| WebhookOut.DefaultPageLimit| 0 (no default) | WEBHOOK_DEFAULT_PAGE_LIMIT    | --webhook-page-limit 50           |
| WebhookIn.GitHubSecret     | env or config  | GITHUB_WEBHOOK_SECRET         | --github-secret value             |
| WebhookIn.StripeSecret     | env or config  | STRIPE_WEBHOOK_SECRET         | --stripe-secret value             |
| WebhookIn.MaxBodyBytes     | 1 MiB          | WEBHOOK_MAX_BODY_BYTES        | --webhook-max-body 1048576        |
| WebhookIn.StripeTolerance  | 5m             | WEBHOOK_STRIPE_TOLERANCE_MS   | --stripe-tolerance 5m             |
| WebhookIn.TopicPrefixGitHub| in.github.     | WEBHOOK_TOPIC_PREFIX_GITHUB   | --github-topic-prefix in.github.  |
| WebhookIn.TopicPrefixStripe | in.stripe.     | WEBHOOK_TOPIC_PREFIX_STRIPE   | --stripe-topic-prefix in.stripe.  |
| WebSocket.Secret           | env or config  | WS_SECRET                     | --ws-secret value                 |
| WebSocket.WSRoutePath      | /ws            | WS_ROUTE_PATH                 | --ws-route /ws                    |
| WebSocket.BroadcastPath    | /_admin/broadcast | WS_BROADCAST_PATH          | --ws-broadcast /_admin/broadcast |
| WebSocket.BroadcastEnabled | true           | WS_BROADCAST_ENABLED          | --ws-broadcast-enabled=false      |
| WebSocket.MaxConnections   | 0 (unlimited)  | (config only)                 | -                                 |
| WebSocket.MaxRoomConnections | 0 (unlimited) | (config only)                | -                                 |

使用 `config.Get*` 辅助函数（参见 `config/env.go`）或 Go 的 `flag` 包，将这些来源转换为 `AppConfig`，然后调用 `core.New(...)`。

## 开发与测试
- 安装 Go 1.24+（匹配 `go.mod`）。
- 运行测试：`go test ./...`
- 使用 Go 工具链进行格式化和静态检查（`go fmt`、`go vet`）。
