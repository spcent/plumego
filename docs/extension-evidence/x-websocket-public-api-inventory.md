# x/websocket Public API Inventory

Module: `x/websocket`

Status: inventory current as of 2026-05-06

Source snapshot:
`docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`

This inventory classifies the exported surface that was reviewed before beta
promotion. The beta promotion source of truth is
`docs/extension-evidence/x-websocket.md` plus
`specs/extension-beta-evidence.yaml`; this file remains a supporting inventory.

## Snapshot Summary

- Total exported symbols in current-head snapshot: 150
- Consts: 16
- Functions: 23
- Types: 25
- Vars: 38
- Methods: 48

## Keep For Beta Surface

These symbols are part of the reviewed websocket transport surface covered by
the beta evidence.

- Server wiring: `New`, `DefaultWebSocketConfig`, `Server`,
  `WebSocketConfig`, `Server.RegisterRoutes`, `Server.Shutdown`,
  `Server.Health`, `Server.Hub`
- Handler wiring: `ServeWSWithConfig`, `ServeRoomFanoutWS`, `ServerConfig`,
  `Message`, `MessageHandler`
- Hub lifecycle and fanout: `NewHubE`, `NewHubWithConfigE`, `Hub`,
  `HubConfig`, `HubMetrics`, `BroadcastResult`, `Hub.Stop`, `Hub.Shutdown`,
  `Hub.TryJoin`, `Hub.CanJoin`, `Hub.Leave`, `Hub.RemoveConn`,
  `Hub.RangeConns`, `Hub.BroadcastRoom`, `Hub.BroadcastAll`,
  `Hub.TryBroadcastRoom`, `Hub.TryBroadcastAll`, `Hub.Metrics`,
  `Hub.GetRooms`, `Hub.GetRoomCount`, `Hub.GetRoomRegistrationCount`
- Connection API: `NewConnE`, `Conn`, `Conn.Close`, `Conn.IsClosed`,
  `Conn.ReadMessage`, `Conn.ReadMessageStream`, `Conn.WriteMessage`,
  `Conn.WriteMessageContext`, `Conn.WriteText`, `Conn.WriteBinary`,
  `Conn.WriteJSON`, `Conn.WriteClose`, `Conn.SetReadLimit`,
  `Conn.SetPingPeriod`, `Conn.SetPongWait`, `Conn.GetLastPong`,
  `Conn.SetMetadata`, `Conn.GetMetadata`, `Conn.DeleteMetadata`,
  `Conn.RangeMetadata`
- Protocol constants: `OpcodeText`, `OpcodeBinary`, close status constants,
  `DefaultSendQueueSize`, `MaxRoomNameLength`, `RoomPasswordHeader`
- Send behavior: `SendBehavior`, `SendBlock`, `SendDrop`, `SendClose`
- Validation: `MessageValidationConfig`, `DefaultMessageValidationConfig`,
  `ValidateTextMessage`, `ValidateRoomName`, `ValidateWebSocketKey`
- Auth extension points: `RoomAuthorizer`, `TokenAuthenticator`,
  `UserInfo`, `ExtractUserInfo`
- Sentinel behavior errors: exported `Err*` values in `module.yaml`

## Keep With Explicit Stable Scope

These symbols are acceptable only if the owner confirms the stable module should
include built-in simple security helpers in addition to transport primitives.
They are implemented and tested, but they widen the public surface beyond the
minimal websocket transport core.

- `NewSimpleRoomAuth`, `SimpleRoomAuth`, `SimpleRoomAuth.SetRoomPassword`,
  `SimpleRoomAuth.AuthorizeRoom`
- `NewSecureRoomAuth`, `SecureRoomAuth`,
  `SecureRoomAuth.SetRoomPassword`, `SecureRoomAuth.AuthenticateToken`,
  `SecureRoomAuth.MaxMessageSize`, `SecureRoomAuth.GetMetrics`,
  `SecureRoomAuth.ResetMetrics`
- `NewSimpleHS256TokenAuth`, `NewHS256TokenAuth`, `SimpleHS256TokenAuth`,
  `SimpleHS256TokenAuth.AuthenticateToken`
- `SecurityConfig`, `SecurityMetrics`,
  `ValidateRoomPassword`, `ValidateSecurityConfig`
- `GenerateSecureSecret`

## Review Before Stable

These symbols are useful today but should receive an owner decision before a
stable promise because they look more like helper or diagnostic API than core
websocket transport.

- Error constructors and structs: `NewCloseError`, `CloseError`,
  `NewSecurityError`, `SecurityError`, `NewValidationError`,
  `ValidationError`
- Utility helpers: `SanitizeForLogging`, `IsTemporary`

## Current Decision

- `x/websocket/module.yaml` is `beta`.
- Server route registration is explicit through `Server.RegisterRoutes`.
- `SecurityConfig` no longer carries Hub runtime fields; queue-full behavior and
  connection-rate limits remain on `HubConfig` and `WebSocketConfig`.
- Security event configuration uses `EnableSecurityEvents`; metric counters are
  always collected through `Hub.Metrics()`.
- No exported symbols were added by the latest runtime cleanup follow-up; the
  current-head snapshot remains 150 exported symbols and was refreshed on
  2026-05-06 because unexported fields inside exported implementation structs
  changed.
- Runtime cleanup tightened route setup validation, outbound data/close protocol
  validation, queued payload ownership, finite socket write deadlines, broadcast
  input validation, bounded-reader documentation, and best-effort security event
  shutdown semantics.
- Follow-up cleanup added bounded `OnMessage` callback dispatch, visible admin
  broadcast runtime outcomes, route-conflict preflight for snapshot-capable
  registrars, absolute queued write deadlines, no-copy rejected send paths, and
  stricter disabled-security-event runtime semantics without changing the
  exported symbol count.
- API contract note: `Conn`/`NewConnE` are server-side primitives that read
  masked client frames and write unmasked server frames. `ReadMessageStream` is
  a bounded reader over buffered frames, not a low-memory or zero-copy streaming
  API.
- Beta promotion is complete. Release refs `d2c25c3` and `ec70358`, matching
  release API snapshots, and `realtime` owner sign-off are recorded in
  `docs/extension-evidence/x-websocket.md`.
