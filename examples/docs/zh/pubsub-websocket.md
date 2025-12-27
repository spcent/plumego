# Pub/Sub 与 WebSocket 模块

**pubsub** 与 WebSocket 帮助在无需外部 Broker 的情况下实现轻量实时消息。

## Pub/Sub 总线
- 使用 `pubsub.New()` 创建并通过 `core.WithPubSub(...)` 注入。
- 支持通配符订阅；开启 `AppConfig.PubSub.Enabled` 可暴露调试页面。
- 约定使用前缀区分来源，如 `in.github.*`、`in.stripe.*`、`ws.broadcast`。
- 订阅回调应保持短小并尊重 `context.Context`；返回错误可在日志/指标中暴露失败。

## WebSocket Hub
- 通过 `core.ConfigureWebSocket()` 或 `ConfigureWebSocketWithOptions` 加载 `core.WebSocketConfig`。
- 默认 `/ws` 路由使用 JWT 认证；广播端点 `/_admin/broadcast` 可用 `BroadcastEnabled` 开关并用共享密钥防护。
- 可调节 `WriteWait`、`ReadWait`、`PingInterval`、worker 数、队列大小以适配客户端行为。
- 与 Pub/Sub 配合，将入站 Webhook 事件直接扇出给在线客户端。

## 示例
```go
bus := pubsub.New()
wsCfg := core.DefaultWebSocketConfig()
wsCfg.Secret = []byte(os.Getenv("WS_SECRET"))
app := core.New(core.WithPubSub(bus))
_, _ = app.ConfigureWebSocketWithOptions(wsCfg)

bus.Subscribe("in.github.*", func(ctx context.Context, evt pubsub.Event) error {
    return bus.Publish(ctx, "ws.broadcast", evt.Payload)
})
```

## 运维建议
- 如需选择性推送，在 payload 中附带频道/用户信息，并在 WebSocket Handler 侧过滤。
- 在订阅回调中使用 `context.WithTimeout`，当下游变慢时尽快返回错误。
- 压测时关注 goroutine 与内存占用，根据结果调整缓冲区再上线。

## 代码位置
- `pubsub/pubsub.go`：进程内总线实现。
- `core/websocket.go`：Hub 配置、JWT 校验与广播处理。
- `examples/reference/main.go`：将入站 Webhook 通过 Pub/Sub 扇出到 WebSocket 的接线示例。
