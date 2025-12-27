# Security 与 Webhook 模块

**security** 与 Webhook 相关包负责入站签名校验和出站投递管理。

## 入站 Webhook
- 通过 `core.WithWebhookIn` 配置密钥与限制（支持 GitHub、Stripe），可设置 `GitHubSecret`、`StripeSecret`、`StripeTolerance`、`MaxBodyBytes`、主题前缀等。
- 请求会进行 HMAC 校验，失败时日志中会出现 `signature mismatch` 提示。
- 通过 Pub/Sub 总线将有效事件发布到 `in.github.*`、`in.stripe.*` 等前缀，供下游消费。

## 出站 Webhook
- 使用 `core.WithWebhookOut` 启用目标管理与投递（触发 Token、基础路径、统计开关、默认分页限制）。
- 组件会挂载 CRUD、触发、重放等 API；启用后将路由注册到应用。
- 默认使用内存存储，可按需替换自定义存储以获得持久化能力。

## 运维建议
- 密钥不应写入仓库，使用 `GITHUB_WEBHOOK_SECRET`、`STRIPE_WEBHOOK_SECRET`、`WEBHOOK_TRIGGER_TOKEN` 等环境变量加载。
- 合理设置 `MaxBodyBytes`，避免超大 Payload 消耗内存。
- 将入站 Webhook、Pub/Sub 与 WebSocket 结合，可实现实时广播。

## 接线示例
```go
app := core.New(
    core.WithWebhookIn(core.WebhookInConfig{
        GitHubSecret:      os.Getenv("GITHUB_WEBHOOK_SECRET"),
        StripeSecret:      os.Getenv("STRIPE_WEBHOOK_SECRET"),
        MaxBodyBytes:      1 << 20,
        StripeTolerance:   5 * time.Minute,
        TopicPrefixGitHub: "in.github.",
        TopicPrefixStripe: "in.stripe.",
    }),
    core.WithWebhookOut(core.WebhookOutConfig{
        TriggerToken: os.Getenv("WEBHOOK_TRIGGER_TOKEN"),
        BasePath:     "/webhooks",
    }),
)
```

## 代码位置
- `security/webhook`：签名校验与辅助方法。
- `core/webhook_in.go`、`core/webhook_out.go`：组件接线与路由注册。
- `env.example`：Webhook 相关环境变量名称。
