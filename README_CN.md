# Plumego — 仅基于golang标准库的 Web 工具包

![plumego logo](docs/branding/logo-v1/final/svg/plumego-primary-horizontal-light.svg)

[![Go 版本](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![版本](https://img.shields.io/badge/version-v1.0.0--rc.1-blue)](https://github.com/spcent/plumego/releases)
[![许可证](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Plumego 是一个小型 Go HTTP 工具包，完全基于标准库实现，同时覆盖路由、中间件、优雅关闭、WebSocket 辅助工具、Webhook 管道以及静态前端托管。它设计为嵌入到你自己的 `main` 包中，而不是作为一个独立的框架二进制文件运行。

`core` 包是稳定的主入口；顶层 `plumego` 包提供常用类型与选项的便捷 re-export。

## 亮点
- **路由器支持分组和参数**：基于 Trie 的匹配器，支持 `/:param` 段、路由冻结，以及每路由/分组的中件栈。
- **中间件链**：日志、恢复、gzip、CORS、超时（默认缓冲上限 10 MiB）、限流、并发限制、请求体大小限制、安全头，以及认证辅助工具，全部包装标准 `http.Handler`。
- **安全辅助**：JWT + 密码工具、安全头策略、输入安全校验与基础防滥用组件，便于进行安全基线加固。
- **集成扩展**：提供 `database/sql`、Redis 缓存与消息队列的轻量适配器/扩展点（`net/mq` 增加可靠任务队列能力，仍处于实验状态）。
- **幂等工具**：提供 `store/idempotency` 的 KV/SQL 幂等存储接口。
- **结构化日志钩子**：接入自定义日志器，并通过中间件钩子收集指标/链路追踪。
- **优雅生命周期**：环境变量加载、连接排水、就绪标志，以及可选的 TLS/HTTP2 配置，带有合理默认值。
- **可选服务**：内置带认证的 WebSocket 中心、进程内 Pub/Sub（带调试快照）、入站/出站 Webhook 路由器，以及从磁盘或嵌入资源提供静态前端。
- **任务调度**：通过 `scheduler` 包提供进程内 cron、延迟任务与可重试任务。

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
    "context"
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
    plumelog "github.com/spcent/plumego/log"
    "github.com/spcent/plumego/middleware/observability"
    "github.com/spcent/plumego/middleware/recovery"
)

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithDebug(),
        core.WithDevTools(),
        core.WithLogger(plumelog.NewGLogger()),
    )

    if err := app.Use(
        observability.RequestID(),
        recovery.Recovery(app.Logger()),
    ); err != nil {
        log.Fatalf("register middleware: %v", err)
    }

    if err := app.AddRoute(http.MethodGet, "/ping", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte("pong"))
    })); err != nil {
        log.Fatalf("register route: %v", err)
    }

    if err := app.Prepare(); err != nil {
        log.Fatalf("prepare app: %v", err)
    }
    if err := app.Start(context.Background()); err != nil {
        log.Fatalf("start runtime: %v", err)
    }

    srv, err := app.Server()
    if err != nil {
        log.Fatalf("build server: %v", err)
    }
    defer app.Shutdown(context.Background())

    log.Fatal(srv.ListenAndServe())
}
```

`plumego.App` 也实现了 `http.Handler`，可以直接挂载到标准库的服务器中：

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/contract"
    "github.com/spcent/plumego"
)

func main() {
    app := plumego.New(plumego.WithAddr(":8080"))

    app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        _ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{
            "status": "ok",
        }, nil)
    })

    log.Println("server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", app))
}
```

## 配置基础
- 环境变量应在 `main` 包中显式加载。`core.WithEnvPath` 仅记录路径，供需要该信息的组件使用，例如 devtools 热重载。
- `core.New(...)` 默认使用 `NoOpLogger`。如果希望有请求日志或运行期日志，请显式注入 `core.WithLogger(...)`。
- 常用变量：`AUTH_TOKEN`（ops 组件默认鉴权配置）、`WS_SECRET`（WebSocket JWT 签名密钥，至少 32 字节）、`WEBHOOK_TRIGGER_TOKEN`、`GITHUB_WEBHOOK_SECRET` 和 `STRIPE_WEBHOOK_SECRET`（详见 `env.example`）。
- 应用默认包括 10485760 字节（10 MiB）请求体限制、256 并发请求限制（带队列）、HTTP 读/写超时，以及 5000ms（5 秒）优雅关闭窗口。可通过 `core.With...` 选项覆盖。
- 安全基线建议通过 `app.Use(...)` 显式组合，例如 `middleware/security.SecurityHeaders(...)` 与 `middleware/ratelimit.AbuseGuard(...)`。
- 调试模式与 devtools 已拆分：`core.WithDebug()` 只开启调试行为，`core.WithDevTools()` 才会挂载调试路由组件。
- Debug 模式（`core.WithDebug`）默认开启 `/_debug` 调试端点（路由表、Middleware、配置快照、指标、pprof、手动重载）、友好 JSON 错误输出，以及 `.env` 热加载。这些端点仅用于本地开发或受保护环境，生产环境应关闭或加访问控制。

## 关键组件
- **路由器**：使用 `Get`、`Post` 等标准库风格方法注册处理器（`func(w http.ResponseWriter, r *http.Request)`）。分组允许附加共享中间件，静态前端可以通过 `frontend.RegisterFromDir` 挂载，并支持缓存/回退选项（`frontend.WithCacheControl`、`frontend.WithIndexCacheControl`、`frontend.WithFallback`、`frontend.WithHeaders`）。
- **中间件**：在启动前使用 `app.Use(...)` 显式链式添加，并保持传输层职责。推荐的可观测性顺序是 `middleware/observability.RequestID`、`middleware/observability.Tracing`、`middleware/observability.HTTPMetrics`、`middleware/observability.AccessLog`，之后再接 `middleware/recovery.Recovery(logger)`。
- **多租户（实验）**：提供租户隔离、配额管理、策略控制和数据库过滤能力，API 仍处于实验阶段，可能变更。详见[多租户](#多租户)章节。
- **运维/管理端点**：可选的受保护运维 API，包含队列状态/重放、回执查询、通道健康、租户配额等能力。通过 `core/components/ops` 挂载，并使用令牌或自定义中间件保护；当 `AllowInsecure` 为 false（默认）且未配置鉴权时会拒绝访问。
- **Contract 工具**：使用 `contract.WriteError` 输出统一错误结构，使用 `contract.WriteResponse` / `Ctx.Response` 输出带 trace id 的标准 JSON 响应。
- **WebSocket 中心**：`ConfigureWebSocket()` 挂载受 JWT 保护的 `/ws` 端点，以及可选的广播端点（受共享密钥保护）。通过 `WebSocketConfig` 自定义工作线程数和队列大小。
- **Pub/Sub + Webhook**：提供 `pubsub.PubSub` 实现以启用 Webhook 分发。出站 Webhook 管理包括目标 CRUD、交付重放和触发令牌；入站接收器处理 GitHub/Stripe 签名，并提供通用 HMAC 验证、重放保护与 IP 白名单。
- **健康检查 + 就绪**：生命周期钩子在启动/关闭期间标记就绪状态，构建元数据（`Version`、`Commit`、`BuildTime`）可通过 ldflags 注入。

## 多租户

Plumego 为 SaaS 应用提供实验性的多租户支持，包括租户隔离、配额管理和策略控制。

### 功能特性

- **租户配置**：灵活的存储后端（内存、数据库 + LRU 缓存）
- **限流能力**：租户级令牌桶（每秒请求数 + 突发控制）
- **配额管理**：租户级用量限制（分钟/小时/日/月窗口），固定窗口执行
- **策略控制**：租户级模型和工具白名单
- **路由策略缓存**：租户级路由策略支持缓存封装
- **数据库隔离**：通过 `TenantDB` 包装器自动为所有 SQL 查询添加租户过滤
- **中间件栈**：租户解析 → 限流 → 配额检查 → 策略执行
- **审计钩子**：可选的回调接口，用于监控配额/策略违规

### 快速配置

```go
import (
    "context"
    "database/sql"
    "log"
    "net/http"
    "time"

    "github.com/spcent/plumego/contract"
    "github.com/spcent/plumego/core"
    tenantmw "github.com/spcent/plumego/middleware/tenant"
    storedb "github.com/spcent/plumego/store/db"
    "github.com/spcent/plumego/tenant"
    tenantpolicy "github.com/spcent/plumego/tenant/middleware"
)

func setupTenantApp(database *sql.DB) *core.App {
    // 创建带缓存的租户配置管理器
    tenantMgr := storedb.NewDBTenantConfigManager(
        database,
        storedb.WithTenantCache(1000, 5*time.Minute),
    )

    // 创建租户中间件需要的管理器
    quotaMgr := tenant.NewWindowQuotaManager(tenantMgr, tenant.NewInMemoryQuotaStore())
    policyEval := tenant.NewConfigPolicyEvaluator(tenantMgr)
    rateLimiter := tenant.NewTokenBucketRateLimiter(
        &tenant.RateLimitConfigProviderFromConfig{Manager: tenantMgr},
    )

    // 租户感知数据库包装器（自动追加 tenant 过滤）
    tenantDB := storedb.NewTenantDB(database)

    app := core.New(core.WithAddr(":8080"))
    api := app.Router().Group("/api")

    // canonical：显式中间件链
    api.Use(tenantmw.TenantResolver(tenantmw.TenantResolverOptions{
        HeaderName: "X-Tenant-ID",
    }))
    api.Use(tenantmw.TenantRateLimit(tenantmw.TenantRateLimitOptions{
        Limiter: rateLimiter,
    }))
    api.Use(tenantpolicy.TenantQuota(tenantpolicy.TenantQuotaOptions{
        Manager: quotaMgr,
        Hooks: tenant.Hooks{
            OnQuota: func(ctx context.Context, decision tenant.QuotaDecision) {
                if !decision.Allowed {
                    log.Printf("租户 %s 超过配额", decision.TenantID)
                }
            },
        },
    }))
    api.Use(tenantpolicy.TenantPolicy(tenantpolicy.TenantPolicyOptions{
        Evaluator: policyEval,
    }))

    api.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rows, err := tenantDB.QueryFromContext(
            r.Context(),
            "SELECT id, email FROM users WHERE active = ?",
            true,
        )
        if err != nil {
            contract.WriteError(w, r, contract.APIError{
                Status:   http.StatusInternalServerError,
                Code:     "db_query_failed",
                Message:  "query failed",
                Category: contract.CategoryServer,
            })
            return
        }
        defer rows.Close()
        _ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{"status": "ok"}, nil)
    }))

    return app
}
```

如果你在代码中管理租户配置（或实现自定义配置提供者），可以这样设置多窗口配额：

```go
tenantMgr := tenant.NewInMemoryConfigManager()
tenantMgr.SetTenantConfig(tenant.Config{
    TenantID: "tenant-id",
    Quota: tenant.QuotaConfig{
        Limits: []tenant.QuotaLimit{
            {Window: tenant.QuotaWindowDay, Requests: 200000},
            {Window: tenant.QuotaWindowMonth, Tokens: 10_000_000},
        },
    },
    Policy: tenant.PolicyConfig{
        AllowedModels: []string{"gpt-4o-mini"},
        AllowedTools:  []string{"search"},
    },
    RateLimit: tenant.RateLimitConfig{
        RequestsPerSecond: 50,
        Burst:             100,
    },
})

// 路由策略缓存（可选）
routePolicyStore := tenant.NewInMemoryRoutePolicyStore()
_ = routePolicyStore.SetRoutePolicy(context.Background(), tenant.RoutePolicy{
    TenantID: "tenant-id",
    Strategy: "weighted",
    Payload:  []byte(`{"rules":[{"provider":"a","weight":70},{"provider":"b","weight":30}]}`),
})
routePolicyCache := tenant.NewInMemoryRoutePolicyCache(1000, 5*time.Minute)
routePolicyProvider := tenant.NewCachedRoutePolicyProvider(routePolicyStore, routePolicyCache)

// 可选：复用同一份配置创建租户限流器
rateLimiter := tenant.NewTokenBucketRateLimiter(
    &tenant.RateLimitConfigProviderFromConfig{Manager: tenantMgr},
)
_ = routePolicyProvider
_ = rateLimiter
```

### 自动查询过滤

`TenantDB` 包装器会自动为所有查询添加租户 ID 过滤：

```go
// 你的查询
rows, err := tenantDB.QueryFromContext(ctx,
    "SELECT * FROM users WHERE active = ?", true)

// 自动转换为
"SELECT * FROM users WHERE tenant_id = ? AND active = ?"
// tenant_id 来自上下文
```

这样可防止跨租户数据泄露，并通过移除手动租户过滤简化业务逻辑。

### 示例应用

参见 `examples/multi-tenant-saas/` 完整示例，包含：
- 租户 CRUD 操作的管理 API
- 租户范围的业务 API
- 带 retry-after 头的配额执行
- 模型/工具的策略验证
- 租户级请求分析

运行示例：
```bash
cd examples/multi-tenant-saas
go run .
```

### 生产环境注意事项

- **性能**：使用带 LRU 缓存的数据库配置管理器（支持 1000+ 租户）
- **安全**：将基于 Header 的租户 ID 替换为签名 JWT token
- **监控**：启用配额/策略钩子进行指标采集
- **扩展**：多实例部署于负载均衡器后，共享数据库

## 后台 Runner
使用最小生命周期接口注册后台任务：

```go
type Runner interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

app.Register(myRunner)
```

Runner 会在 HTTP server 启动前启动，并在优雅关闭时停止。

## 任务调度
使用 `scheduler` 包提供进程内 cron 与延迟任务：

```go
sch := scheduler.New(scheduler.WithWorkers(2))
sch.Start()
defer sch.Stop(context.Background())

sch.AddCron("cleanup", "0 * * * *", func(ctx context.Context) error {
    // 每小时任务
    return nil
})

sch.Delay("one-off", 5*time.Second, func(ctx context.Context) error {
    return nil
})
```

可选增强包括管理 Handler（`scheduler.NewAdminHandler`）以及可插拔持久化（`scheduler.WithStore`，支持内存或 KV）。
还可以通过 `scheduler.WithPanicHandler` 与 `scheduler.WithMetricsSink` 接入异常处理与指标汇报。
作业状态快照提供统一状态机（`queued`、`scheduled`、`running`、`retrying`、`failed`、`canceled`、`completed`），并包含 `StateUpdated` 时间戳；`JobQuery` 支持通过 `States` 字段按状态过滤。
Admin Handler 在 `/scheduler/jobs` 支持 `state` 查询参数（可重复）按作业状态过滤。

## 认证契约
Plumego 通过 `contract` 中的接口将认证、授权、会话校验分离，推荐用中间件组合：

```go
app.Use(auth.Authenticate(jwtManager.Authenticator(jwt.TokenTypeAccess)))
app.Use(auth.SessionCheck(sessionStore, sessionValidator))
app.Use(auth.Authorize(jwt.PolicyAuthorizer{Policy: jwt.AuthZPolicy{AnyRole: []string{"admin"}}}, "", ""))

protected := middleware.Apply(
	http.HandlerFunc(adminHandler),
	auth.Authenticate(jwtManager.Authenticator(jwt.TokenTypeAccess)),
	auth.SessionCheck(sessionStore, sessionValidator),
	auth.Authorize(jwt.PolicyAuthorizer{Policy: jwt.AuthZPolicy{AnyRole: []string{"admin"}}}, "", ""),
)
```

`security/jwt` 提供契约适配器（`jwtManager.Authenticator`、`jwt.PolicyAuthorizer`、`jwt.PermissionAuthorizer`），以保持存储与策略实现的解耦。

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
healthManager, err := health.NewHealthManager(health.HealthCheckConfig{})
if err != nil {
    log.Fatal(err)
}

app := core.New(core.WithHealthManager(healthManager))
app.Get("/health/ready", health.ReadinessHandler(healthManager).ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)
```

`ReadinessHandler` 会返回传入 `HealthManager` 的就绪状态（ready 为 true 时返回 200，否则 503）。当通过 `core.WithHealthManager` 挂载后，core 生命周期会自动更新 ready/not-ready 状态。

## 可观测性适配器
无需自行编写适配器，即可将日志中间件接入指标/链路追踪后端：

- `metrics.NewPrometheusCollector(namespace)` 实现 `observability.MetricsCollector`，并通过 `collector.Handler()` 暴露 `/metrics` 处理程序。
- `metrics.NewOpenTelemetryTracer(name)` 实现 `observability.Tracer`，发出带有 HTTP 元数据的 span。

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
使用 `config.LoadEnv` 加载环境变量，或绑定命令行标志；`config.ConfigManager` 也提供 `LoadBestEffort` 用于跳过可选配置源失败，并提供 `ReloadWithValidation` 做事务式热加载；配置键在读取时会规范化为小写的 snake_case，因此 CamelCase 和 UPPER_SNAKE 会映射到同一值；带 `_MS` 后缀的环境变量单位为毫秒；使用下表实现可预测的部署。

| AppConfig 字段             | 默认值          | 环境变量                       | Flag 示例                          |
|----------------------------|-----------------|--------------------------------|------------------------------------|
| Addr                       | :8080          | APP_ADDR                      | --addr :8080                      |
| EnvFile                    | .env           | APP_ENV_FILE                  | --env-file .env                   |
| Debug                      | false          | APP_DEBUG                     | --debug                           |
| ShutdownTimeout            | 5000ms         | APP_SHUTDOWN_TIMEOUT_MS       | --shutdown-timeout 5000ms         |
| ReadTimeout                | 30000ms        | APP_READ_TIMEOUT_MS           | --read-timeout 30000ms            |
| ReadHeaderTimeout          | 5000ms         | APP_READ_HEADER_TIMEOUT_MS    | --read-header-timeout 5000ms      |
| WriteTimeout               | 30000ms        | APP_WRITE_TIMEOUT_MS          | --write-timeout 30000ms           |
| IdleTimeout                | 60000ms        | APP_IDLE_TIMEOUT_MS           | --idle-timeout 60000ms            |
| MaxHeaderBytes             | 1048576        | APP_MAX_HEADER_BYTES          | --max-header-bytes 1048576        |
| EnableHTTP2                | true           | APP_ENABLE_HTTP2              | --http2=false                     |
| DrainInterval              | 500ms          | APP_DRAIN_INTERVAL_MS         | --drain-interval 500ms            |
| MaxBodyBytes               | 10485760       | APP_MAX_BODY_BYTES            | --max-body-bytes 10485760         |
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
| WebhookIn.MaxBodyBytes     | 1048576        | WEBHOOK_MAX_BODY_BYTES        | --webhook-max-body 1048576        |
| WebhookIn.StripeTolerance  | 300000ms       | WEBHOOK_STRIPE_TOLERANCE_MS   | --stripe-tolerance 300000ms       |
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

## 开发服务器与仪表盘

`plumego` CLI 包含一个强大的开发服务器，它本身就是使用 plumego 框架构建的。它提供热重载、实时监控和 Web 仪表盘，大大提升开发体验。

仪表盘**默认启用** - 只需运行 `plumego dev` 即可开始使用。

**定位差异与生产建议**
- `core.WithDebug` 会暴露应用级 `/_debug` 端点，属于应用自身调试接口，生产环境应关闭或加访问控制。
- `plumego dev` 仪表盘是本地开发工具，运行独立的仪表盘服务，不建议在生产环境对外暴露。
- 仪表盘可能读取应用的 `/_debug` 端点用于路由/配置/指标/pprof 展示，因此仅建议在本地或受控环境启用。

### 快速开始

```bash
plumego dev
# 仪表盘：http://localhost:9999
# 你的应用：http://localhost:8080
```

### 仪表盘功能

每个 `plumego dev` 会话都包含：

- **实时日志**：流式传输应用程序的 stdout/stderr，支持过滤
- **路由浏览器**：自动发现并展示应用程序的所有 HTTP 路由
- **指标仪表盘**：监控运行时间、PID、健康状态和性能
- **构建管理**：查看构建输出并手动触发重新构建
- **应用控制**：从 UI 中启动、停止和重启应用程序
- **热重载**：文件更改时自动重新构建和重启（< 5 秒）

### 自定义配置

```bash
# 自定义应用端口
plumego dev --addr :3000

# 自定义仪表盘端口
plumego dev --dashboard-addr :8888

# 自定义监听模式
plumego dev --watch "**/*.go,**/*.yaml"

# 调整热重载灵敏度
plumego dev --debounce 1s
```

完整文档请参见 `cmd/plumego/DEV_SERVER.md`。

## 文档
规范文档入口与优先级顺序：`docs/README.md`。

详细文档请参见 `examples/docs` 目录：
- `examples/docs/en/` - 英文文档
- `examples/docs/zh/` - 中文文档
