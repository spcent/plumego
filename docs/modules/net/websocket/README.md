# WebSocket Hub

> **Package**: `github.com/spcent/plumego/net/websocket`

## Overview

`net/websocket` provides a standard-library-only WebSocket server implementation with:

- RFC6455 frame handling (text/binary/control frames)
- Room-based broadcast hub with worker pool
- Bounded per-connection send queues (`SendBlock` / `SendDrop` / `SendClose`)
- JWT + room-password authentication
- Optional origin allow-list validation
- Message validation and connection metadata helpers

## Quick Start

```go
package main

import (
    "net/http"
    "time"

    "github.com/spcent/plumego/net/websocket"
)

func main() {
    hub := websocket.NewHubWithConfig(websocket.HubConfig{
        WorkerCount:        4,
        JobQueueSize:       1024,
        MaxConnections:     10000,
        MaxRoomConnections: 1000,
    })
    defer hub.Stop()

    auth := websocket.NewSimpleRoomAuth([]byte("your-32-byte-secret"))
    auth.SetRoomPassword("chat", "chat-room-password")

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        websocket.ServeWSWithConfig(w, r, websocket.ServerConfig{
            Hub:            hub,
            Auth:           auth,
            QueueSize:      256,
            SendTimeout:    200 * time.Millisecond,
            SendBehavior:   websocket.SendBlock,
            AllowedOrigins: []string{"https://app.example.com"},
            ReadLimit:      1 << 20, // 1MB
        })
    })

    _ = http.ListenAndServe(":8080", nil)
}
```

## Authentication

Supported by `RoomAuthenticator`:

- `NewSimpleRoomAuth(secret)`
- `NewSecureRoomAuth(secret, SecurityConfig)`

Handshake input:

- room: `?room=chat`
- room password: `?room_password=...`
- token: `Authorization: Bearer <token>` or `?token=...`

## Server Config

`ServeWSWithConfig` uses `ServerConfig`:

- `Hub`, `Auth` (required)
- `QueueSize`, `SendTimeout`, `SendBehavior`
- `AllowedOrigins` (empty means allow all)
- `ReadLimit` (max inbound frame payload size)
- `MessageValidation` (text-message validation rules)

When `ReadLimit` is set, text-message validation max length is clamped to the same upper bound.

## Hub Semantics

`MaxConnections` and `GetTotalCount()` are based on **room registrations**.

- One connection in one room = 1
- One connection joined to 3 rooms = 3

Use `BroadcastRoom(room, op, data)` for room broadcast and `BroadcastAll(op, data)` for global broadcast.

After `hub.Stop()`, new joins are rejected with `ErrHubStopped`.

## Conn Helpers

- `WriteText`, `WriteBinary`, `WriteJSON`, `WriteMessage`
- `ReadMessage`, `ReadMessageStream`
- `SetReadLimit`, `SetPingPeriod`, `SetPongWait`
- `SetMetadata`, `GetMetadata`, `DeleteMetadata`, `RangeMetadata`

## Notes

- `UpgradeClient` is deprecated and intentionally not implemented.
- For strict CSRF/origin control, prefer `ServeWSWithConfig` over `ServeWSWithAuth`.
