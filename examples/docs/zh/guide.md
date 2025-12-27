# Plumego 模块级使用指南（简体中文）

> 本文遵循现代开发文档的结构化风格，对 plumego 的核心模块进行分层解读。每个模块均提供概念说明、配置要点、最佳实践、错误排查清单以及示例代码，便于团队直接复制到项目模板中。全文所有案例均基于 `examples/reference` 内置的示例服务，可通过 `go run` 快速运行验证。

## Core：生命周期与应用编排

Core 模块负责应用的全局生命周期、配置注入与启动顺序，是构建任何 plumego 服务的基石。为了覆盖常见上线场景，本节从配置策略、调试开关、启动流程、并发安全性以及可观察性挂载五个维度展开。

### 概念与职责

- **显式构造**：使用 `core.New` 明确传入监听地址、调试模式、Pub/Sub 总线、指标收集器、链路追踪器等依赖，避免隐式全局变量带来的状态漂移。
- **可组合能力**：Core 不负责业务逻辑，而是聚合 Router、Middleware、Metrics、Webhook 等模块，提供清晰的组装点。
- **启动即校验**：`Boot()` 会在真正监听端口前完成路由冻结、依赖检查和中间件链构建，启动失败会返回错误而非静默忽略。

### 配置要点

1. **调试模式**：`core.WithDebug()` 会放宽部分安全限制并增加日志细节，仅建议在开发环境开启。
2. **地址与优雅退出**：默认监听 `:8080`，可通过环境变量或 `core.WithAddr` 覆盖；Core 内置 `http.Server` 的优雅关停逻辑，确保在 SIGTERM 到达时完成正在处理的请求。
3. **Webhook 输入**：使用 `core.WithWebhookIn` 配置 GitHub、Stripe 等入站 webhook，需显式传入密钥、容忍时间窗口与体积上限，防止超大请求耗尽内存。
4. **Webhook 输出**：`core.WithWebhookOut` 允许同时启用出站 webhook 管理，与内置的触发端点组合即可实现事件扇出。
5. **可观察性挂载**：配合 `metrics.NewPrometheusCollector` 与 `metrics.NewOpenTelemetryTracer`，Core 会自动将请求指标注入日志中间件，减少手写埋点。

### 启动顺序与线程安全

- Core 在构造时就会拷贝必要的指针，避免外部修改配置影响运行时行为；
- 在 `Boot()` 前可以安全注册路由与中间件，`Boot()` 调用后 Router 会被冻结，新增路由将触发 panic，确保路由表一致；
- 长时间运行的 goroutine（如 webhook 出站调度器）建议在 `Boot()` 之前 `Start()` 并在 `defer` 中 `Stop()`，保持与 HTTP 生命周期对齐。

### 常见排查清单

- **端口被占用**：确认 `WithAddr` 设置未冲突；在容器内使用 `netstat -tnlp` 诊断；
- **环境变量缺失**：对于必填的 webhook 密钥，未配置会导致签名校验失败；可在日志中搜索 `signature mismatch` 关键字；
- **调试输出过多**：生产环境务必移除 `WithDebug()`，并确保日志级别由外部配置控制。

### 示例

```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
    core.WithPubSub(pubsub.New()),
    core.WithMetricsCollector(metrics.NewPrometheusCollector("plumego_example")),
)
app.EnableRecovery()
app.EnableLogging()
// 注册完所有路由后，统一启动
if err := app.Boot(); err != nil {
    log.Fatalf("server stopped: %v", err)
}
```

### 进阶实践与案例

- **多实例一致性**：在微服务场景下可以通过配置中心下发 `WithAddr`、超时和中间件开关，所有实例使用同一段初始化代码，降低配置漂移风险。
- **蓝绿发布**：利用 `Boot()` 前的路由冻结特性，可以在新版本中先准备好完整路由树，待探针通过后切流，避免半初始化状态收到流量。
- **优雅停机**：在收到退出信号时，先调用出站 webhook 的 `Stop()` 停止重试，再关闭 HTTP 服务，最后关闭 Pub/Sub，确保事件不丢失。
- **配置回退**：保持 `envOr` 之类的包裹函数简洁，并在日志中输出有效值，便于排查回退后的真实运行参数。

## Router：声明式路由与参数匹配

Router 模块提供类似前缀树的高性能匹配能力，支持静态路径、命名参数与通配符。为了保证每个接口在文档、代码与监控中的一致性，本节聚焦路由组织、分组与调试工具。

### 路由定义策略

- **REST 风格优先**：建议使用 `GET /resource/:id`、`POST /resource` 等约定式路径，便于客户端缓存与监控聚合；
- **参数命名**：命名参数以 `:id` 形式出现，值会注入 `context` 供 handler 读取；通配符 `*filepath` 适用于静态文件或多层级文档。
- **分组复用**：通过 `Group` 复用前缀和中间件，例如 `api := router.Group("/api")`，在子路由中只需关注实际业务路径。

### 路由调试与冻结

- **查看已注册路由**：Router 会记录 `routes` 映射，配合日志可以输出所有 Method-Path 组合；
- **冻结机制**：当 `Boot()` 被调用后，Router 进入冻结状态，再次注册会 panic，用于捕获初始化顺序错误；
- **链式中间件**：`Group` 会继承父中间件，子路由可以追加更多拦截器，形成可预测的执行顺序。

### 性能与可维护性

- 路由节点使用前缀树存储，避免线性遍历；
- 在高并发场景下推荐将动态路径控制在两级以内，减少匹配深度；
- 对于需要灰度或限流的接口，可在 Group 层引入自定义中间件，而不是在 handler 内手写条件分支。

### 示例

```go
api := app.Router().Group("/api")
api.Get("/users/:id", userHandler)
api.Post("/users", createUserHandler)
assets := app.Router().Group("/docs")
assets.Get("/*filepath", docsHandler)
```

### 进阶实践与案例

- **可观测路径命名**：在路由层统一约定路径格式与操作动词（如 `/v1/users/:id/disable`），使监控面板能按动词或资源聚合。
- **分组边界**：将内部管理接口与公开 API 放入不同的 Group，并在 Group 层添加认证或 IP 白名单中间件，减少重复代码。
- **调试导出**：在开发环境添加专用路由 `/debug/routes`，遍历 `routes` 映射输出，帮助新人快速理解接口面貌。
- **性能 Profiling**：热点路径尽量保持静态段较多，避免在第一段就使用通配符；必要时可以将重度计算的 handler 拆分为预处理和后处理两个路由。

## Middleware：链式管控与跨切面能力

Middleware 模块提供了恢复、日志、CORS、超时、压缩等跨切面能力。本节从执行顺序、编排原则、安全注意事项和自定义模式四个角度展开，保证中间件既可插拔又可观测。

### 执行顺序

- 中间件在路由注册时绑定，执行顺序为 **全局 → 分组 → 路由专属**；
- 恢复中间件应尽量靠前，以捕获后续环节的 panic；
- 日志、指标中间件通常放在靠外层，以覆盖整个调用链。

### 自定义模式

1. **遵循签名**：函数类型 `middleware.Middleware` 需要接收并返回 `http.Handler`，可参照现有实现；
2. **避免全局变量**：使用闭包捕获配置，或在构造函数中准备所需依赖，防止并发读写问题；
3. **错误响应一致性**：保持统一的 JSON 或文本格式，确保上层监控可以通过状态码与响应体识别问题来源。

### 安全注意事项

- CORS 配置需限制可信 Origin，避免默认放开；
- Body 限制、超时限制应针对上传接口开启，防止大体积请求耗尽内存；
- 在日志中避免输出敏感字段，可利用自定义中间件对 `Authorization` 头做掩码。

### 示例

```go
app.EnableRecovery()
app.EnableLogging()
app.EnableCORS()
// 自定义限流
app.Use(middleware.NewConcurrencyLimiter(100))
```

### 进阶实践与案例

- **组合策略**：在分组中堆叠限流、鉴权、审计中间件，借助继承关系避免重复注册；对 WebSocket、Webhook 等特殊路径可使用独立分组隔离配置。
- **错误语义**：在自定义中间件里返回统一的错误结构体（含 trace id、错误码、用户提示），配合日志即可实现统一的错误观测面板。
- **防御性编程**：在中间件中使用 `context.WithTimeout` 控制下游处理时长，并在超时时写入特定响应头，便于客户端重试或降级。
- **安全审计**：在访问控制中间件中记录角色、资源、操作类型，结合 Pub/Sub 推送到审计系统，满足企业合规要求。

## Pub/Sub 与 WebSocket：事件驱动与实时传输

plumego 内置的 Pub/Sub 与 WebSocket 组件适合搭建轻量级事件系统。本节涵盖主题命名、消息扇出、连接安全、心跳策略以及与 webhook 的联动方式。

### 设计要点

- **主题前缀**：建议使用模块化前缀，如 `in.github.*`、`in.stripe.*`、`ws.broadcast`，便于订阅过滤；
- **并发安全**：内存总线实现已处理锁粒度问题，但仍应避免长时间阻塞的订阅处理器；
- **心跳与回收**：WebSocket 默认提供心跳，生产环境应根据客户端需求调整 `PingInterval` 与 `WriteWait`。

### 与 Webhook 的结合

入站 webhook 事件可直接发布到总线，再由 WebSocket 订阅者实时消费，实现多终端同步；出站 webhook 可以作为另一个消费者，将内部事件推送到外部系统。

### 示例

```go
bus := pubsub.New()
wsCfg := core.DefaultWebSocketConfig()
wsCfg.Secret = []byte(envOr("WS_SECRET", "dev-secret"))
_, _ = app.ConfigureWebSocketWithOptions(wsCfg)
// 订阅事件并广播
bus.Subscribe("in.github.*", func(ctx context.Context, evt pubsub.Event) error {
    return bus.Publish(ctx, "ws.broadcast", evt.Payload)
})
```

### 进阶实践与案例

- **回溯与重放**：在订阅函数内部增加幂等校验（如事件 id 去重），并根据业务需求将关键事件写入持久化存储，支持事后重放。
- **广播与单播**：通过在事件 payload 中携带频道或用户 id，在 WebSocket 层实现选择性推送，避免一股脑广播导致带宽浪费。
- **超时控制**：为订阅回调设置 `context.WithTimeout`，当外部服务变慢时及时返回错误，让上游监控捕获异常并触发告警。
- **压测策略**：在压测环境中模拟高频 webhook 输入，观察内存占用与 goroutine 数量，确保订阅处理器不会成为瓶颈。

## Metrics 与 Health：可观察性默认开启

可观察性是运维友好的关键。plumego 将 Prometheus、OpenTelemetry 以及标准健康探针整合在统一接口，帮助团队快速接入监控平台。

### 指标采集

- 使用 `metrics.NewPrometheusCollector(namespace)` 生成采集器，并在路由中暴露 `/metrics` 供 Prometheus 抓取；
- 中间件会自动记录请求时延、状态码分布、panic 次数等关键指标，无需额外代码；
- 在多实例部署时，可通过 namespace 或实例标签区分来源。

### 链路追踪

- `metrics.NewOpenTelemetryTracer(serviceName)` 提供 tracer，配合日志中间件将 trace id 注入日志；
- 建议在入口处（如 webhook handler）创建新的 span，并在业务逻辑中传递 `context.Context`，确保链路完整。

### 健康探针

- `health.ReadinessHandler()` 用于就绪检测，应覆盖依赖服务（数据库、缓存、消息队列）的可用性；
- `health.BuildInfoHandler()` 可以暴露版本与编译信息，便于部署对比；
- 结合 Kubernetes 的 `readinessProbe` 和 `livenessProbe`，可在异常时自动重启。

### 进阶实践与案例

- **指标分层**：按照入口、业务、外部依赖三个层次设计指标前缀，区分延迟来源；通过标签标记路由、调用方与地区，方便维度切片。
- **错误预算**：结合状态码分布与 P99 延迟，制定错误预算与 SLO；在预算接近阈值时可自动降低新功能灰度流量。
- **探针扩展**：为 `/health/ready` 增加数据库和缓存探测，使用 context 超时包裹，避免探针阻塞影响实际请求处理。
- **链路追踪规范**：在团队内统一 trace id 的传递键名与日志格式，便于跨语言服务串联调用链。

## Security 与 Webhook：安全边界与签名校验

安全模块围绕 webhook 入站、出站以及请求签名展开。本节提供默认安全策略、错误处理与审计建议，帮助团队在快速迭代时保持合规。

### 入站安全

- GitHub/Stripe 等 webhook 校验需要配置密钥与允许的时间偏差，防止重放攻击；
- `MaxBodyBytes` 限制请求体积，避免恶意放大；
- 建议开启 `EnableLogging` 并结合审计日志，仅记录必要的元数据（事件类型、签名是否通过），避免存储原始 payload。

### 出站安全

- 出站 webhook 通过 `TriggerToken` 控制触发端点访问；
- 建议在 `webhookout.Config` 中限制并发重试次数与退避策略，防止外部系统故障时耗尽资源；
- 使用持久化存储时，可替换 `NewMemStore` 为自定义实现。

### 审计与合规

- 结合日志追踪字段（trace id、request id）记录敏感操作；
- 对于需要合规的场景，可在 Pub/Sub 层增加事件白名单，拒绝未知类型；
- 建议定期轮换 webhook 密钥，并在配置中预留过渡期同时接受新旧签名。

### 进阶实践与案例

- **签名演练**：在开发环境准备伪造请求，确认签名校验的错误日志与响应码符合预期；同时验证时间戳回放攻击是否被拒绝。
- **故障注入**：对出站 webhook 的目标地址进行故障注入（如 5xx、超时），观察退避与重试是否符合设定，并确保触发端点不会导致任务堆积。
- **密钥管理**：将 webhook 密钥存放于专用的密钥管理服务，应用启动时通过环境变量或挂载文件读取，避免写入镜像。
- **访问分级**：区分公共、合作方、内部三类 webhook，使用不同的触发 token 和路由前缀，减少误调用风险。

## Frontend 与 静态资源：内嵌 UI 与文档浏览

Frontend 模块提供将静态文件嵌入二进制并通过统一前缀提供服务的能力。本节将演示如何挂载 UI、托管 Markdown 文档（如本指南），并与路由中其他模块协同。

### 资源嵌入与托管

- 使用 `//go:embed` 将 `ui/*` 或文档目录打包进二进制，无需额外部署步骤；
- 调用 `frontend.RegisterFS(router, http.FS(staticFS), frontend.WithPrefix("/"))` 将资源暴露为静态站点；
- 当需要同时托管 UI 与 API 时，可将静态前缀设为 `/`，API 统一置于 `/api` 下，避免路径冲突。

### 文档导航

- 本示例新增 `/docs` 路由，会自动列出语言与模块导航，并将 Markdown 渲染为简洁的 HTML；
- 支持 `*/` 通配符路径，便于直接访问 `/docs/zh/core` 或 `/docs/en/router`；
- 文档渲染会对文本进行 HTML 转义，并提供基础样式，保证可读性与安全性。

### 运营建议

- 若需自定义主题，可在文档目录加入额外 CSS，并在渲染模板中引用；
- 对于大体积文件建议开启压缩中间件，或将资源托管到 CDN；
- 在多语言场景下，保持文件命名一致便于自动导航与链接分享。

### 进阶实践与案例

- **文档构建流程**：通过 CI 将 Markdown 质量检查（标题级别、链接可用性）纳入管道，保证 `/docs` 展示的内容随版本同步发布。
- **导航定制**：利用新路由自动生成语言与模块导航，结合自定义样式可以轻量替代独立文档站点，适合私有化部署或内网环境。
- **演示同步**：在嵌入的 UI 中加入 WebSocket 客户端示例，与文档描述相互印证，帮助开发者快速验证实时能力。
- **缓存策略**：为静态资源设置合理的缓存头，并在文档更新时通过文件名哈希或版本前缀避免旧缓存干扰。

> 至此，plumego 的核心模块指南已覆盖从生命周期管理到静态资源托管的全链路实践。团队可根据自身需求裁剪模块、调整中间件顺序，并在 `/docs` 内维护与代码同步的中文文档，确保新成员能快速上手、及时排查并稳定上线。
