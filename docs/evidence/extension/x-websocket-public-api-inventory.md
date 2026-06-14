# x/websocket Public API Inventory

Module: `x/websocket`

Inventory status: updated 2026-06-14 (alias removal sync: NewHubE/NewHubWithConfigE/NewConnE removed, method names corrected)

Source snapshot:
`docs/evidence/extension/snapshots/first-batch/x-websocket-head.snapshot`

This inventory classifies the exported surface that was reviewed before `beta`
promotion. The promotion source of truth is
`docs/evidence/extension/x-websocket.md` plus
`specs/extension-beta-evidence.yaml`; this file remains a supporting inventory.

## Snapshot Summary

- Total exported symbols in current-head snapshot: 150
- Consts: 16
- Functions: 23
- Types: 25
- Vars: 38
- Methods: 48

## Included In Reviewed `beta` Surface

These symbols are part of the reviewed websocket transport surface covered by
the recorded `beta` evidence.

- Server wiring: `New`, `DefaultWebSocketConfig`, `Server`,
  `WebSocketConfig`, `Server.RegisterRoutes`, `Server.Shutdown`,
  `Server.Health`, `Server.Hub`
- Handler wiring: `ServeWSWithConfig`, `ServeRoomFanoutWS`, `ServerConfig`,
  `Message`, `MessageHandler`
- Hub lifecycle and fanout: `NewHub`, `NewHubWithConfig`, `Hub`,
  `HubConfig`, `HubMetrics`, `BroadcastResult`, `Hub.Stop`, `Hub.Shutdown`,
  `Hub.TryJoin`, `Hub.CanJoin`, `Hub.Leave`, `Hub.RemoveConn`,
  `Hub.RangeConns`, `Hub.BroadcastRoom`, `Hub.BroadcastAll`,
  `Hub.TryBroadcastRoom`, `Hub.TryBroadcastAll`, `Hub.Metrics`,
  `Hub.GetRooms`, `Hub.GetRoomCount`, `Hub.GetRoomRegistrationCount`
- Connection API: `NewConn`, `Conn`, `Conn.Close`, `Conn.IsClosed`,
  `Conn.ReadMessage`, `Conn.ReadMessageReader`, `Conn.WriteMessage`,
  `Conn.WriteMessageContext`, `Conn.WriteText`, `Conn.WriteBinary`,
  `Conn.WriteJSON`, `Conn.WriteClose`, `Conn.SetReadLimit`,
  `Conn.SetPingPeriod`, `Conn.SetPongWait`, `Conn.GetLastPong`,
  `Conn.SetMetadata`, `Conn.GetMetadata`, `Conn.DeleteMetadata`,
  `Conn.RangeMetadata`
- Protocol constants: `OpcodeText`, `OpcodeBinary`, close status constants,
  `DefaultSendQueueSize`, `MaxRoomNameLength`
- Send behavior: `SendBehavior`, `SendBlock`, `SendDrop`, `SendClose`
- Validation: `MessageValidationConfig`, `DefaultMessageValidationConfig`,
  `ValidateTextMessage`, `ValidateRoomName`, `ValidateWebSocketKey`
- Auth extension points: `RoomAuthorizer`, `TokenAuthenticator`,
  `UserInfo`, `ExtractUserInfo`
- Sentinel behavior errors: exported `Err*` values in `module.yaml`

## Included Only With Explicit Scope Confirmation

These symbols are acceptable only if the owner confirms that the stable module
should include built-in simple security helpers in addition to transport
primitives. They are implemented and tested, but they widen the public surface
beyond the minimal websocket transport core.

- `NewSimpleRoomAuth`, `SimpleRoomAuth`, `SimpleRoomAuth.SetRoomPassword`,
  `SimpleRoomAuth.CheckRoomPassword`
- `NewSecureRoomAuth`, `SecureRoomAuth`,
  `SecureRoomAuth.SetRoomPassword`, `SecureRoomAuth.CheckRoomPassword`,
  `SecureRoomAuth.VerifyJWT`, `SecureRoomAuth.MaxMessageSize`,
  `SecureRoomAuth.GetMetrics`, `SecureRoomAuth.ResetMetrics`
- `NewSimpleHS256TokenAuth`, `SimpleHS256TokenAuth`,
  `SimpleHS256TokenAuth.VerifyJWT`
- `SecurityConfig`, `SecurityMetrics`,
  `ValidateRoomPassword`, `ValidateSecurityConfig`
- `GenerateSecureSecret`

## Needs Scope Decision Before Stable Promise

These symbols are useful today but should receive an owner decision before any
stable promise because they look more like helper or diagnostic API than core
websocket transport.

- Error constructors and structs: `NewCloseError`, `CloseError`,
  `NewValidationError`, `ValidationError`
- Utility helpers: `SanitizeForLogging`

## Promotion Posture

- `x/websocket/module.yaml` is `beta`.
- Blocker state: none in the promotion evidence record.
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
- API contract note: `Conn`/`NewConn` are server-side primitives that read
  masked client frames and write unmasked server frames. `ReadMessageReader` is
  a bounded reader over buffered frames, not a low-memory or zero-copy streaming
  API.
- `beta` promotion is complete. Release refs `d2c25c3` and `ec70358`, matching
  release API snapshots, and `realtime` owner sign-off are recorded in
  `docs/evidence/extension/x-websocket.md`.
