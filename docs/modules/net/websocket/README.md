# WebSocket Migration

`net/websocket` has been moved into [`x/websocket`](../x-websocket/README.md).

Use `github.com/spcent/plumego/x/websocket` for:
- `Hub`, `HubConfig`, `NewHub`, and `NewHubWithConfig`
- `NewSimpleRoomAuth` and `NewSecureRoomAuth`
- `ServeWSWithAuth` and `ServeWSWithConfig`
- `NewComponent` and `DefaultWebSocketConfig`

This page is kept only as a migration marker for historical links.

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
