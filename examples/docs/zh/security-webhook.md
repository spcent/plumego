# Security 与 Webhook 模块

Plumego 提供入站 webhook 校验、出站投递管理以及轻量鉴权工具。

## 入站 webhook
通过 `core.WithWebhookIn` 挂载 GitHub、Stripe 等受保护的接收端点。

```go
app := core.New(core.WithWebhookIn(core.WebhookInConfig{
    Enabled:           true,
    Pub:               bus, // 可选：将事件发布到进程内总线
    GitHubSecret:      config.GetString("GITHUB_WEBHOOK_SECRET", "dev-github"),
    StripeSecret:      config.GetString("STRIPE_WEBHOOK_SECRET", "whsec_dev"),
    MaxBodyBytes:      1 << 20,
    StripeTolerance:   5 * time.Minute,
    TopicPrefixGitHub: "in.github.",
    TopicPrefixStripe: "in.stripe.",
}))
```

- GitHub：基于共享密钥的 HMAC 签名校验。
- Stripe：签名 + 时间容差校验。
- 可选将事件发布到 Pub/Sub，以解耦处理流程。
- `MaxBodyBytes` 防止超大请求体。

## 出站 webhook
`core.WithWebhookOut` 接入出站投递服务（示例使用内存存储，可通过 `webhookout.Service` 扩展）。

```go
store := webhookout.NewMemStore()
svc := webhookout.NewService(store, webhookout.ConfigFromEnv())
app := core.New(core.WithWebhookOut(core.WebhookOutConfig{
    Enabled:          true,
    Service:          svc,
    TriggerToken:     config.GetString("WEBHOOK_TRIGGER_TOKEN", "dev-token"),
    BasePath:         "/webhooks",
    IncludeStats:     true,
    DefaultPageLimit: 50,
}))
svc.Start(context.Background())
defer svc.Stop()
```

能力概览：
- `BasePath` 下提供 webhook 目标与密钥的 CRUD HTTP 接口。
- 投递具备重试、重放与可选统计暴露。
- Trigger Token 保护变更端点，可按需叠加鉴权/限流中间件。

## 简易鉴权辅助
- `middleware.SimpleAuth("token")`：校验 `Authorization: Bearer <token>`。
- `middleware.APIKey(header, value)`：强制特定头部的 API Key。

适用于管理后台或内部端点（如 metrics、webhook 管理），建议配合 TLS，并通过环境变量轮换密钥。

## 运维提示
- 避免在日志中输出 webhook 签名或密钥。
- 入站处理保持快速；将事件发布到 Pub/Sub 或队列，避免阻塞 webhook 发送方。
- 在启动阶段及早校验密钥配置，确保错误尽早暴露。

## 代码位置
- `core/webhook.go`：入站配置接线。
- `net/webhookout/`：出站投递服务、存储实现与 HTTP 处理器。
- `middleware/auth.go`：Bearer/API-Key 辅助中间件。
- `examples/reference/main.go`：入/出站 webhook 与 Pub/Sub 扇出的完整示例。
