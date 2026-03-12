# Security 与 Webhook 模块

Plumego 通过显式 webhook 组件支持入站校验与出站投递管理。

## 入站 webhook（GitHub / Stripe）
通过 `core.WithComponent(...)` 挂载入站组件：

```go
bus := pubsub.New()

app := core.New(
    core.WithComponent(webhook.NewWebhookInComponent(webhook.WebhookInConfig{
        Enabled:           true,
        Pub:               bus,
        GitHubSecret:      os.Getenv("GITHUB_WEBHOOK_SECRET"),
        StripeSecret:      os.Getenv("STRIPE_WEBHOOK_SECRET"),
        MaxBodyBytes:      1 << 20,
        StripeTolerance:   5 * time.Minute,
        TopicPrefixGitHub: "in.github.",
        TopicPrefixStripe: "in.stripe.",
    }, bus, nil)),
)
```

默认入站端点：
- `POST /webhooks/github`
- `POST /webhooks/stripe`

## 出站 webhook 管理
通过组件挂载出站管理：

```go
store := webhook.NewMemStore()
svc := webhook.NewService(store, webhook.ConfigFromEnv())

app := core.New(
    core.WithComponent(webhook.NewWebhookOutComponent(webhook.WebhookOutConfig{
        Enabled:          true,
        Service:          svc,
        TriggerToken:     os.Getenv("WEBHOOK_TRIGGER_TOKEN"),
        BasePath:         "/webhooks",
        IncludeStats:     true,
        DefaultPageLimit: 50,
    })),
)
```

出站组件会在 `BasePath` 下暴露目标管理、投递列表/详情、重放和触发接口。

## 通用签名校验
自定义 provider 可直接使用 `x/webhook`：

```go
result, err := webhook.VerifyHMAC(r, webhook.HMACConfig{
    Secret:   []byte(os.Getenv("WEBHOOK_SECRET")),
    Header:   "X-Signature",
    Prefix:   "sha256=",
    Encoding: webhook.EncodingHex,
})
if err != nil {
    http.Error(w, "invalid signature", webhook.HTTPStatus(err))
    return
}
_ = result.Body
```

## 安全中间件辅助
显式使用中间件子包：

- `auth.SimpleAuth(token)`
- `security.SecurityHeaders(nil)`
- `ratelimit.AbuseGuard(...)`

## 运维建议
- 不记录原始签名、webhook 密钥或私钥。
- 入站处理尽量短，复杂逻辑发布到 Pub/Sub 异步处理。
- 启动期校验关键密钥配置，失败即退出。
