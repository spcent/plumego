# Plumego 文档索引（简体中文）

此目录按照 plumego 的代码结构拆分为独立模块文档，便于按需查阅。

## 模块索引
- [Core](core.md)：生命周期编排、配置选项、组件与排障。
- [Router](router.md)：前缀树匹配、路由分组、参数与中间件继承。
- [Middleware](middleware.md)：内置防护、组合规则、自定义模板。
- [Pub/Sub & WebSocket](pubsub-websocket.md)：进程内总线、WebSocket Hub 与实时扇出模式。
- [Metrics & Health](metrics-health.md)：Prometheus/OpenTelemetry 适配与就绪/构建探针。
- [Security & Webhook](security-webhook.md)：入站签名校验与出站投递管理。

## 示例运行方式
文档中的代码均基于 `examples/reference` 演示应用，可直接运行验证：

```bash
go run ./examples/reference
```

## 风格提示
- 保持边界：业务逻辑写在你的 `main` 包，plumego 提供基础设施。
- 在调用 `Boot()` 前完成路由与中间件注册；启动后新增注册会 panic 属于预期。
- 优先通过 `core.With...` 选项或 `env.example` 中的环境变量配置行为。
