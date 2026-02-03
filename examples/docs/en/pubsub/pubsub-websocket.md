# Pub/Sub and WebSocket modules

The **pubsub** package and WebSocket helpers deliver lightweight real-time messaging without external brokers.

## Pub/Sub bus
- Create with `pubsub.New()` and inject via `core.WithPubSub(...)` so other modules (webhooks, WebSocket hub) share the same bus.
- Topics support wildcards: `in.github.*`, `in.stripe.*`, `ws.broadcast`.
- Subscribe with context-aware handlers; return errors to surface failures in logs and metrics.
- Enable the debug snapshot UI by setting `AppConfig.PubSub.Enabled` (or `PUBSUB_DEBUG_ENABLED=true`), which exposes a simple introspection page.

```go
bus := pubsub.New()
_ = bus.Subscribe("orders.*", func(ctx context.Context, evt pubsub.Event) error {
    fmt.Printf("topic=%s payload=%s\n", evt.Topic, string(evt.Payload))
    return nil
})
_ = bus.Publish(context.Background(), "orders.created", []byte(`{"id":1}`))
```

## WebSocket hub
- Configure through `core.ConfigureWebSocket()` or `core.ConfigureWebSocketWithOptions(core.WebSocketConfig)`.
- Default route `/ws` expects a JWT signed with `WebSocketConfig.Secret`. Issue tokens from your app or a CLI.
- Optional broadcast endpoint (default `/_admin/broadcast`) is guarded by the same secret; disable via `BroadcastEnabled`.
- Tune timing (`WriteWait`, `ReadWait`, `PingInterval`) and buffers (`SendQueueSize`, `HubWorkers`) to match client behavior.

```go
bus := pubsub.New()
app := core.New(core.WithPubSub(bus))
wsCfg := core.DefaultWebSocketConfig()
wsCfg.Secret = []byte(os.Getenv("WS_SECRET"))
_, _ = app.ConfigureWebSocketWithOptions(wsCfg)

// Fan out inbound events to connected clients.
bus.Subscribe("in.github.*", func(ctx context.Context, evt pubsub.Event) error {
    return bus.Publish(ctx, core.WebSocketBroadcastTopic, evt.Payload)
})
```

## Putting Pub/Sub and WebSocket together
1. Inject the bus into the app with `core.WithPubSub`.
2. Configure the WebSocket hub so clients can connect and authenticate.
3. Publish domain events (from webhook receivers or HTTP handlers) to topic names your clients subscribe to.
4. Use the broadcast endpoint or a custom subscriber to push those events to WebSocket connections.

## Operational tips
- Use `context.WithTimeout` inside subscribers to avoid stuck workers when downstream systems slow down.
- Namespacing topics (`in.*`, `out.*`, `ws.*`) keeps dashboards and filters organized.
- Load-test hub buffers; increase `SendQueueSize` and `HubWorkers` if clients burst data or reconnect frequently.

## Where to look in the repo
- `pubsub/pubsub.go`: in-process bus implementation.
- `core/websocket.go`: hub configuration, JWT validation, broadcast handling, topic constants.
- `examples/reference/main.go`: wiring inbound webhooks to Pub/Sub and WebSocket broadcasts.
