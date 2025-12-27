# Pub/Sub 与 WebSocket 模块

**pubsub** 与 WebSocket 辅助提供无外部依赖的轻量实时消息能力。

## Pub/Sub 总线
- 使用 `pubsub.New()` 创建并通过 `core.WithPubSub(...)` 注入，webhook 与 WebSocket Hub 可共享同一总线。
- 主题支持通配：`in.github.*`、`in.stripe.*`、`ws.broadcast`。
- 订阅时使用支持上下文的处理函数，返回错误可在日志与指标中暴露失败。
- 通过 `AppConfig.PubSub.Enabled`（或 `PUBSUB_DEBUG_ENABLED=true`）开启调试快照 UI，用于简单观测。

```go
bus := pubsub.New()
_ = bus.Subscribe("orders.*", func(ctx context.Context, evt pubsub.Event) error {
    fmt.Printf("topic=%s payload=%s\n", evt.Topic, string(evt.Payload))
    return nil
})
_ = bus.Publish(context.Background(), "orders.created", []byte(`{"id":1}`))
```

## WebSocket Hub
- 通过 `core.ConfigureWebSocket()` 或 `core.ConfigureWebSocketWithOptions(core.WebSocketConfig)` 配置。
- 默认 `/ws` 路由需要使用 `WebSocketConfig.Secret` 签发的 JWT，令牌可由你的应用或 CLI 颁发。
- 可选广播端点（默认 `/_admin/broadcast`）使用相同密钥保护，`BroadcastEnabled` 可关闭。
- 根据客户端行为调整时间窗口（`WriteWait`、`ReadWait`、`PingInterval`）与缓冲（`SendQueueSize`、`HubWorkers`）。

```go
bus := pubsub.New()
app := core.New(core.WithPubSub(bus))
wsCfg := core.DefaultWebSocketConfig()
wsCfg.Secret = []byte(os.Getenv("WS_SECRET"))
_, _ = app.ConfigureWebSocketWithOptions(wsCfg)

// 将入站事件广播给连接的客户端。
bus.Subscribe("in.github.*", func(ctx context.Context, evt pubsub.Event) error {
    return bus.Publish(ctx, core.WebSocketBroadcastTopic, evt.Payload)
})
```

## 组合步骤
1. 通过 `core.WithPubSub` 注入总线。
2. 配置 WebSocket Hub，客户端按约定鉴权连接。
3. 将业务或 webhook 事件发布到客户端订阅的主题。
4. 使用广播端点或自定义订阅者将事件推送到 WebSocket 连接。

## 运维建议
- 在订阅函数内使用 `context.WithTimeout` 防止下游阻塞导致 worker 堵塞。
- 通过主题命名空间（`in.*`、`out.*`、`ws.*`）保持监控和过滤的清晰性。
- 压测 Hub 缓冲；若客户端高频突发或重连，可增加 `SendQueueSize` 与 `HubWorkers`。

## 代码位置
- `pubsub/pubsub.go`：进程内总线实现。
- `core/websocket.go`：Hub 配置、JWT 校验、广播处理和主题常量。
- `examples/reference/main.go`：入站 webhook 到 Pub/Sub 与 WebSocket 广播的接线示例。
