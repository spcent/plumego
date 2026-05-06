# x/websocket Public API Inventory

Module: `x/websocket`

Status: inventory current as of 2026-05-05

Source snapshot:
`docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`

This inventory classifies the current development-head exported surface for
freeze review. It does not promote the module, replace release API snapshots, or
represent owner sign-off.

## Snapshot Summary

- Total exported symbols in current-head snapshot: 150
- Consts: 16
- Functions: 23
- Types: 25
- Vars: 38
- Methods: 48

## Keep For Stable Candidate

These symbols are part of the intended websocket transport surface and should
remain stable candidates if release evidence and owner sign-off are later
provided.

- Server wiring: `New`, `DefaultWebSocketConfig`, `RouteRegistrar`, `Server`,
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
- `NewHS256TokenAuth`, `HS256TokenAuth`,
  `HS256TokenAuth.AuthenticateToken`
- `SecurityConfig`, `SecurityMetrics`, `SecurityEvent`,
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

- `x/websocket/module.yaml` remains `experimental`.
- `RouteRegistrar` is now part of the explicit server wiring candidate surface.
- `SecurityConfig` no longer carries Hub runtime fields; queue-full behavior and
  connection-rate limits remain on `HubConfig` and `WebSocketConfig`.
- Security event configuration uses `EnableSecurityEvents`; metric counters are
  always collected through `Hub.Metrics()`.
- No exported symbols were added by the latest runtime cleanup; the current-head
  snapshot remains 150 exported symbols.
- Runtime cleanup tightened route setup validation, outbound data/close protocol
  validation, queued payload ownership, finite socket write deadlines, broadcast
  input validation, bounded-reader documentation, and best-effort security event
  shutdown semantics.
- API contract note: `Conn`/`NewConnE` are server-side primitives that read
  masked client frames and write unmasked server frames. `ReadMessageStream` is
  a bounded reader over buffered frames, not a low-memory or zero-copy streaming
  API.
- Stable promotion remains blocked until release refs, release API snapshots,
  and `realtime` owner sign-off exist.
