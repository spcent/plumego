# Pub/Sub 与 WebSocket 模块

`pubsub` 与 WebSocket 辅助提供无外部依赖的轻量实时消息能力。

## Pub/Sub 总线
使用 `pubsub.New()` 创建进程内总线。

```go
bus := pubsub.New()
sub, err := bus.SubscribePattern("orders.*", pubsub.DefaultSubOptions())
if err != nil {
    log.Fatal(err)
}
defer sub.Cancel()

_ = bus.Publish("orders.created", pubsub.Message{
    ID:    "evt-1",
    Topic: "orders.created",
    Type:  "order.created",
    Time:  time.Now(),
    Data:  map[string]any{"id": 1},
})
```

若其他模块（如 webhook 入站组件）也要发布事件，请共享同一个 `bus` 实例并显式注入。

## WebSocket Hub
通过 `app.ConfigureWebSocket()` 或 `app.ConfigureWebSocketWithOptions(...)` 配置。

```go
app := core.New(core.WithAddr(":8080"))
wsCfg := core.DefaultWebSocketConfig()
wsCfg.Secret = []byte(os.Getenv("WS_SECRET"))
hub, err := app.ConfigureWebSocketWithOptions(wsCfg)
if err != nil {
    log.Fatal(err)
}

// 将 pubsub 事件广播给 websocket 客户端。
go func() {
    for msg := range sub.C() {
        payload, _ := json.Marshal(msg)
        hub.BroadcastAll(ws.OpcodeText, payload)
    }
}()
```

## 组合步骤
1. 创建一个共享的 `pubsub.New()` 总线。
2. 将总线注入产生日志事件的组件（webhook/业务服务）。
3. 统一订阅并扇出到 WebSocket 客户端。
4. 保持主题命名空间清晰（`in.*`、`domain.*`、`out.*`）。

## 代码位置
- `pubsub/pubsub.go`：进程内总线实现。
- `core/websocket_wrapper.go`：应用层 websocket 配置。
- `net/websocket/hub.go`：广播与连接管理实现。
- `examples/reference/internal/app`：接近生产的接线示例。
