# Net Module

> **Package Path**: `github.com/spcent/plumego/net` | **Stability**: Medium | **Priority**: P2

## Overview

The `net/` package provides network utilities for Plumego applications: service discovery, inbound/outbound webhooks, WebSocket hubs, inter-process communication, and in-memory message queuing.

**Key Features**:
- **Service Discovery**: Static and Consul-based service registry
- **Inbound Webhooks**: Verified webhook receivers (GitHub, Stripe, custom)
- **Outbound Webhooks**: Reliable webhook delivery with retries and dead-letter queuing
- **WebSocket Hub**: JWT-authenticated WebSocket connections with broadcasting
- **IPC**: Unix/Windows inter-process communication
- **Message Queue**: In-memory message queue with persistence and retry

## Subpackages

| Package | Description |
|---------|-------------|
| `net/discovery` | Service discovery (static, Consul) |
| `net/webhookin` | Inbound webhook receivers with signature verification |
| `net/webhookout` | Outbound webhook delivery with retry |
| `net/websocket` | WebSocket hub with JWT authentication |
| `net/ipc` | Inter-process communication (Unix/Windows) |
| `net/mq` | In-memory message queue with persistence |
| `net/http` | HTTP client helpers and security utilities |

## Quick Start

### Service Discovery

```go
import "github.com/spcent/plumego/net/discovery"

// Static service registry
registry := discovery.NewStatic(
    discovery.Service{Name: "payment-service", Addr: "payment:8080"},
    discovery.Service{Name: "email-service", Addr: "email:8080"},
)

addr, err := registry.Resolve("payment-service")
```

### Inbound Webhooks

```go
import "github.com/spcent/plumego/net/webhookin"

// GitHub webhooks with HMAC verification
handler := webhookin.NewGitHub(
    webhookin.WithSecret(os.Getenv("GITHUB_WEBHOOK_SECRET")),
    webhookin.OnPush(handlePush),
    webhookin.OnPullRequest(handlePR),
)
app.Post("/webhooks/github", handler)
```

### Outbound Webhooks

```go
import "github.com/spcent/plumego/net/webhookout"

sender := webhookout.NewService(webhookout.Config{
    Workers:   8,
    QueueSize: 2048,
    Timeout:   10 * time.Second,
})
sender.Start()

sender.Send(webhookout.Webhook{
    URL:     "https://example.com/webhook",
    Payload: eventData,
    Secret:  "webhook-signing-secret",
})
```

### WebSocket

```go
import (
    "net/http"
    "os"
    "time"

    "github.com/spcent/plumego/net/websocket"
)

hub := websocket.NewHubWithConfig(websocket.HubConfig{
    WorkerCount:    4,
    JobQueueSize:   1024,
    MaxConnections: 10000,
})
defer hub.Stop()

auth := websocket.NewSimpleRoomAuth([]byte(os.Getenv("WS_SECRET")))

app.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
    websocket.ServeWSWithConfig(w, r, websocket.ServerConfig{
        Hub:            hub,
        Auth:           auth,
        QueueSize:      256,
        SendTimeout:    200 * time.Millisecond,
        SendBehavior:   websocket.SendBlock,
        AllowedOrigins: []string{"https://app.example.com"},
    })
})

hub.BroadcastRoom("channel", websocket.OpcodeText, message)
```

### Message Queue

```go
import "github.com/spcent/plumego/net/mq"

broker := mq.NewBroker()
broker.Start()

// Producer
broker.Publish("order.created", orderEvent)

// Consumer
broker.Subscribe("order.created", func(msg *mq.Message) error {
    return processOrder(msg)
})
```

## Module Documentation

- **[Service Discovery](discovery/README.md)** — Static and Consul-based discovery
- **[Inbound Webhooks](webhookin/README.md)** — GitHub, Stripe, custom receivers
- **[Outbound Webhooks](webhookout/README.md)** — Reliable webhook delivery
- **[WebSocket Hub](websocket/README.md)** — Real-time WebSocket connections
- **[Message Queue](mq/README.md)** — In-memory message broker

## Related Documentation

- [Middleware: Proxy](../middleware/proxy.md) — HTTP reverse proxy
- [Security](../security/) — Webhook signature verification
- [AI: SSE](../x/ai/sse.md) — SSE streaming (AI-specific)
