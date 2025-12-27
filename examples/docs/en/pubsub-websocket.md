# Pub/Sub and WebSocket modules

The **pubsub** and WebSocket helpers deliver lightweight real-time messaging without external brokers.

## Pub/Sub bus
- Construct with `pubsub.New()` and inject via `core.WithPubSub(...)`.
- Publish/subscribe with topic filters (wildcards allowed) and optional debug UI if `PubSub.Enabled` is set in `AppConfig`.
- Use prefixed topics (e.g., `in.github.*`, `in.stripe.*`, `ws.broadcast`) to separate sources.
- Keep subscription handlers short and respect `context.Context`; return errors to surface failures in logs/metrics.

## WebSocket hub
- Configure via `core.ConfigureWebSocket()` or `core.ConfigureWebSocketWithOptions` using `core.WebSocketConfig`.
- Default route `/ws` requires JWT auth; broadcast endpoint `/_admin/broadcast` can be toggled with `BroadcastEnabled` and secured with a shared secret.
- Tune `WriteWait`, `ReadWait`, `PingInterval`, worker counts, and queue sizes to match client behavior.
- Pair with the Pub/Sub bus to fan out inbound webhook events to connected clients.

## Usage example
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

## Operational tips
- For selective delivery, include channel/user identifiers in payloads and filter inside the WebSocket handler.
- Use `context.WithTimeout` inside subscribers to fail fast when downstream systems are slow.
- Monitor goroutine and memory usage during load tests; adjust buffer sizes before production rollout.

## Where to look in the repo
- `pubsub/pubsub.go` for the in-process bus implementation.
- `core/websocket.go` for hub configuration, JWT validation, and broadcast handling.
- `examples/reference/main.go` for wiring inbound webhooks to Pub/Sub and WebSocket broadcasts.
