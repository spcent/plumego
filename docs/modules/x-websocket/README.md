# x/websocket

## Purpose

`x/websocket` owns websocket server helpers and explicit route registration for websocket transport.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen
- Beta candidate once the extension stability policy's two-release API freeze
  evidence is available. Current blocker: no repository release history proves
  two consecutive minor releases without exported `x/websocket` API changes.

## Use this module when

- the task is websocket transport behavior
- the change is connection lifecycle or hub behavior

## Do not use this module for

- app bootstrap
- generic HTTP routing outside websocket transport

## First files to read

- `x/websocket/module.yaml`
- `x/websocket/websocket.go`
- `reference/standard-service` when checking canonical bootstrap shape

## Public entrypoints

- Server wiring: `New`, `DefaultWebSocketConfig`, `Server`, `WebSocketConfig`
- Handler wiring: `ServeWSWithConfig`, `ServeRoomFanoutWS`, `ServerConfig`,
  `Message`, `MessageHandler`
- Hub lifecycle: `NewHubE`, `NewHubWithConfigE`, `Hub`, `HubConfig`,
  `HubMetrics`
- Connection API: `NewConnE`, `Conn`, `SendBehavior`, close/operation constants
- Auth/security helpers: `RoomAuthorizer`, `TokenAuthenticator`,
  `NewSimpleRoomAuth`, `NewHS256TokenAuth`, `NewSecureRoomAuth`,
  `SecurityConfig`, `SecurityMetrics`, `SecurityEvent`
- Validation helpers and errors: `ValidateWebSocketKey`, `ValidateTextMessage`,
  `ValidateRoomName`, `ValidateRoomPassword`, `ValidateSecurityConfig`,
  `SanitizeForLogging`, exported sentinel errors, and exported error structs

## Main risks when changing this module

- websocket auth regression
- broadcast behavior regression
- connection lifecycle regression

## Boundary rules

- keep websocket setup explicit and out of `core`; do not add hidden goroutines or global state at import time
- keep transport concerns (`ServeWSWithConfig`, `ServeRoomFanoutWS`) inside `x/websocket`; do not push connection-level logic into stable roots or middleware
- keep auth and broadcast gates reviewable and testable in isolation
- handle room-password setup errors explicitly; do not hide hash failures behind log-only behavior
- keep security metrics instance-scoped (`SecureRoomAuth.GetMetrics`, `Hub.Metrics`) instead of reintroducing global wrappers
- treat `x/websocket` as the app-facing websocket transport surface; app-level session management belongs in the calling handler

## Handler contract

`ServeWSWithConfig` is the low-level transport handler. It completes the
handshake, joins the configured room, reads complete inbound messages, validates
text payloads, and passes each accepted message to `ServerConfig.OnMessage`.
It does not broadcast client messages by default.

Use `ServeRoomFanoutWS` when the application wants built-in room fanout behavior
where each accepted client message is broadcast back to the same room.
Room authorization, token authentication, anonymous access, and query-token
support are separate `ServerConfig` choices.

`DefaultWebSocketConfig` keeps the admin broadcast route disabled. Applications
that enable it must configure a separate `BroadcastSecret` of at least 32 bytes.
The broadcast endpoint reads that secret from `Authorization: Bearer ...`, not
from URL query parameters, and caps request bodies with
`BroadcastMaxBodyBytes` (default 1 MiB).

Browser handshakes with an `Origin` header require explicit
`AllowedOrigins` configuration. Non-browser clients without `Origin` skip the
origin check; use `[]string{"*"}` only for development or intentionally public
endpoints.

Room names must be 1-128 ASCII characters using letters, digits, `-`, `_`, `.`,
or `:`. Room passwords are read from the `X-Room-Password` header; URL query
passwords are rejected.
Handshake validation requires `Sec-WebSocket-Version: 13`.

`Conn.WriteClose` sends a best-effort close frame and then closes TCP; it does
not wait for a peer close frame. `ReadMessageStream` returns a bounded reader,
not a zero-copy stream: continuation frames are read lazily, but frame payloads
are still buffered in memory. Read limits apply to the complete message,
including all continuation frames; oversized pooled buffers are discarded rather
than retained.

Hub metrics are always collected and exposed through `Hub.Metrics()`. Security
events are opt-in through `HubConfig.EnableSecurityMetrics`; applications can
consume them with `HubConfig.SecurityEventHandler`. Event producers never block
on that handler; the security monitor goroutine invokes it and drops later
events if the internal buffer fills. Hub debug logging uses `HubConfig.Logger`
when provided and is no-op by default.
Use `TryBroadcastRoom` or `TryBroadcastAll` when a caller needs accepted and
dropped send counts; `BroadcastRoom` and `BroadcastAll` remain fire-and-forget
wrappers.

Security helpers clone caller-provided JWT secrets before storing them and
reject secrets shorter than 32 bytes. `NewHS256TokenAuth` is a lightweight
HS256 verifier for compact bearer tokens: it validates the signature and an
optional integer `exp` claim, but it is not an OIDC/JWT policy engine for
issuer, audience, `nbf`, or `iat` enforcement.

## Current test coverage

- connection configuration (read limit, ping period, pong wait)
- `Hub` lifecycle: `Stop` idempotency, `Shutdown` (empty and with connections, context cancellation), `TryJoin`/`Leave`/`RemoveConn` lifecycle, `RangeConns` iteration and early return
- capacity errors: `ErrHubFull`, `ErrRoomFull`, `ErrHubStopped` from `TryJoin`/`CanJoin` after stop or at limit
- broadcast: `BroadcastRoom`, `BroadcastAll` (positive path and no-op after stop), race-condition coverage under concurrent goroutines
- security: `ValidateSecurityConfig`, `ValidateWebSocketKey`, `ValidateRoomPassword`, `SecureRoomAuth`, security metrics, connection limit enforcement
- validation: text message sanitization and control-character handling
- server setup: `ServeRoomFanoutWS` (method-not-allowed, bad-request, bad-room-password), `ServeWSWithConfig` invalid-config rejection, config normalization

## Beta readiness

`x/websocket` satisfies the current coverage and boundary portions of
`docs/EXTENSION_STABILITY_POLICY.md`: hub lifecycle, shutdown, capacity errors,
broadcast no-op behavior, validation, security checks, and server setup failure
paths have focused tests.

The module remains `experimental` until the release-history criterion is
verifiable. Promotion to `beta` requires evidence that exported `x/websocket`
symbols have not changed for two consecutive minor releases, plus owner
sign-off recorded with the promotion card.

## Canonical change shape

- keep websocket setup explicit and out of `core`
- keep auth and broadcast gates reviewable
- keep handshake failures on stable structured error codes for method, upgrade, key, origin, room, token, join, hijack, and server-configuration failures
- handle room-password setup errors explicitly; do not hide hash failures behind log-only behavior
- keep security metrics instance-scoped (`SecureRoomAuth.GetMetrics`, `Hub.Metrics`) instead of reintroducing global wrappers
- treat `x/websocket` as the app-facing websocket transport surface
