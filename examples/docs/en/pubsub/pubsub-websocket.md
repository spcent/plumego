# Pub/Sub and WebSocket modules

`pubsub` and WebSocket helpers provide lightweight realtime messaging without external brokers.

## Pub/Sub bus
Create an in-process bus with `pubsub.New()`.

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

To share events with other modules (for example webhook-in), pass the same `bus` instance into those component constructors.

## WebSocket hub
Mount explicitly via `x/websocket`.

```go
app := core.New(core.WithAddr(":8080"))
wsCfg := xwebsocket.DefaultWebSocketConfig()
wsCfg.Secret = []byte(os.Getenv("WS_SECRET"))
comp, err := xwebsocket.NewComponent(wsCfg, false, app.Logger())
if err != nil {
    log.Fatal(err)
}
_ = app.MountComponent(comp)
hub := comp.Hub()

// Forward pubsub events to websocket clients.
go func() {
    for msg := range sub.C() {
        payload, _ := json.Marshal(msg)
        hub.BroadcastAll(ws.OpcodeText, payload)
    }
}()
```

## Integration steps
1. Build one shared `pubsub.New()` bus.
2. Pass it into components that produce events (webhook, domain services).
3. Subscribe once and fan out to WebSocket clients.
4. Keep topic namespaces explicit (`in.*`, `domain.*`, `out.*`).

## Where to look in repo
- `pubsub/pubsub.go`: in-process bus implementation.
- `x/websocket/websocket.go`: app-level websocket component.
- `x/websocket/hub.go`: hub broadcast internals.
- `reference/standard-service/internal/app`: production-style wiring.
