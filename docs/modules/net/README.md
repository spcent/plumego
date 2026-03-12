# Net Module

> **Package Path**: `github.com/spcent/plumego/net` | **Stability**: Medium | **Priority**: P2

Legacy note: webhook and websocket implementations have moved to `x/webhook` and `x/websocket`. This module index now covers the remaining `net/*` packages that are still intentionally rooted here.

## Overview

The `net/` tree now primarily covers discovery, IPC, HTTP helpers, messaging, and queueing.

**Key Features**:
- **Service Discovery**: Static and Consul-based service registry
- **IPC**: Unix/Windows inter-process communication
- **Message Queue**: In-memory message queue with persistence and retry
- **Messaging**: queue-backed notification orchestration
- **HTTP helpers**: outbound client helpers and security utilities

## Subpackages

| Package | Description |
|---------|-------------|
| `net/discovery` | Service discovery (static, Consul) |
| `net/ipc` | Inter-process communication (Unix/Windows) |
| `net/messaging` | queue-backed messaging orchestration |
| `net/mq` | In-memory message queue with persistence |
| `net/http` | HTTP client helpers and security utilities |

Moved extension surfaces:

| Package | New entrypoint |
|---------|----------------|
| `net/webhookin` | `x/webhook` |
| `net/webhookout` | `x/webhook` |
| `net/websocket` | `x/websocket` |

## Quick Start

### Service Discovery

```go
import "github.com/spcent/plumego/x/discovery"

// Static service registry
registry := discovery.NewStatic(
    discovery.Service{Name: "payment-service", Addr: "payment:8080"},
    discovery.Service{Name: "email-service", Addr: "email:8080"},
)

addr, err := registry.Resolve("payment-service")
```

### Message Queue

```go
import "github.com/spcent/plumego/x/mq"

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
- **[Message Queue](mq/README.md)** — In-memory message broker
- **[Webhook Migration](webhookin/README.md)** — moved to `x/webhook`
- **[WebSocket Migration](websocket/README.md)** — moved to `x/websocket`

## Related Documentation

- [Middleware: Proxy](../middleware/proxy.md) — HTTP reverse proxy
- [Security](../security/) — Webhook signature verification
- [AI: SSE](../x/ai/sse.md) — SSE streaming (AI-specific)
