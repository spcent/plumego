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

- `New`
- `DefaultWebSocketConfig`
- `NewConnE`
- `NewHub`
- `NewHubWithConfigE`
- `ServeWSWithConfig`

## Main risks when changing this module

- websocket auth regression
- broadcast behavior regression
- connection lifecycle regression

## Boundary rules

- keep websocket setup explicit and out of `core`; do not add hidden goroutines or global state at import time
- keep transport concerns (`ServeWSWithConfig`) inside `x/websocket`; do not push connection-level logic into stable roots or middleware
- keep auth and broadcast gates reviewable and testable in isolation
- require JWT by default in `ServeWSWithConfig`; set `AllowUnauthenticated` only for room-password-only development or trusted internal flows
- treat origin allow-all as an explicit opt-in through `AllowAllOrigins`
- treat query-string JWT transport as disabled by default; set `AllowQueryToken` only for trusted non-browser clients that cannot send headers
- keep `DefaultWebSocketConfig` free of environment reads; callers must pass secrets explicitly
- keep admin broadcast disabled by default; enable it only with a dedicated `BroadcastSecret` or `BroadcastAuthorizer`
- keep admin broadcast request bodies bounded with `BroadcastMaxBytes`
- require `Sec-WebSocket-Version: 13` during handshake
- treat `RegisterRoutes` errors as assembly failures; nil registrars, nil hubs, empty websocket paths, and empty enabled broadcast paths must fail visibly
- perform the real Hub join before writing `101 Switching Protocols`; if capacity changes after the pre-check, return a normal HTTP error before upgrade
- treat `Hub.Shutdown` as a hard connection close path, not a WebSocket close-frame handshake
- use `NewConnE` for connection construction with explicit validation errors
- use `NewHubWithConfigE` for hub configuration with explicit validation errors
- reject non-positive ping/pong durations at setter boundaries
- reject malformed RFC6455 frames: non-zero RSV bits, reserved opcodes, non-minimal payload lengths, malformed close payloads, and invalid continuation ordering
- treat inbound reads as bounded whole-message reads: `ReadLimit` applies to the total fragmented message, and `ReadMessageStream` reads continuation frames incrementally but does not provide an unbounded streaming bypass
- treat `TryJoin` as the only public join path; all joins enforce capacity and closed-state checks
- treat `HubMetrics.ActiveConnections` as unique connections and `HubMetrics.RoomRegistrations` as connection-room registrations
- pass `HubConfig.Logger` for hub logs; the default hub logger discards output
- handle room-password setup errors explicitly; do not hide hash failures behind log-only behavior
- keep security metrics instance-scoped (`SecureRoomAuth.GetMetrics`, `Hub.Metrics`) instead of reintroducing global wrappers
- treat `x/websocket` as the app-facing websocket transport surface; app-level session management belongs in the calling handler

## Current test coverage

- connection configuration (read limit, ping period, pong wait)
- connection construction validation and write-after-close behavior
- hub construction validation, caller-owned logging, and metrics count semantics
- RFC6455 frame parsing negatives for RSV bits, reserved opcodes, non-minimal lengths, close payloads, masking, and continuation ordering
- fragmented and unfragmented read-limit enforcement at and above configured limits
- `Hub` lifecycle: `Stop` idempotency, `Shutdown` (empty and with hard-closed connections, context cancellation), `TryJoin`/`Leave`/`RemoveConn` lifecycle, `RangeConns` iteration and early return
- capacity errors: `ErrHubFull`, `ErrRoomFull`, `ErrHubStopped` from `TryJoin`/`CanJoin` after stop or at limit
- broadcast: `BroadcastRoom`, `BroadcastAll` (positive path and no-op after stop), race-condition coverage under concurrent goroutines
- security: `ValidateSecurityConfig`, `ValidateWebSocketKey`, `ValidateRoomPassword`, `SecureRoomAuth`, security metrics, connection limit enforcement
- validation: text message sanitization, dangerous-pattern detection, control-character handling
- server setup: `ServeWSWithConfig` method-not-allowed, bad-request, version-13 requirement, bad-room-password, invalid-config rejection, missing-token rejection, query-token rejection by default, explicit origin allow behavior, config normalization
- route registration: nil registrar, nil hub, empty websocket path, duplicate routes, and empty enabled broadcast path
- admin broadcast: disabled-by-default behavior, dedicated secret or authorizer validation, JWT-secret rejection, empty body behavior, and oversized-body rejection

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
- keep JWT-required, unauthenticated, and origin allow-all behavior explicit in configuration
- keep query-string JWT transport disabled unless `AllowQueryToken` is explicitly set
- keep `DefaultWebSocketConfig` deterministic; read environment variables in application wiring before filling config
- keep admin broadcast separately authorized with `BroadcastSecret` or `BroadcastAuthorizer`; never reuse the JWT `Secret`
- bound admin broadcast request bodies before reading them
- require RFC6455 version 13 in the HTTP upgrade request
- keep route registration fail-visible and handle returned errors at every call site
- keep capacity denial before WebSocket upgrade, including post-hijack capacity races
- document shutdown as hard-close unless a future card adds non-blocking close-frame delivery
- keep connection constructors and mutable timing setters fail-visible instead of panic-prone
- keep full config constructors and join paths error-returning; do not reintroduce compatibility-only bypass helpers
- keep metrics names precise enough to distinguish unique connections from room registrations
- keep protocol parsing strict without adding compression or extension negotiation implicitly
- keep large-message behavior bounded; do not describe `ReadMessageStream` as unbounded or zero-copy streaming
- keep handshake failures on stable structured error codes for method, upgrade, key, origin, room, token, join, hijack, and server-configuration failures
- handle room-password setup errors explicitly; do not hide hash failures behind log-only behavior
- keep security metrics instance-scoped (`SecureRoomAuth.GetMetrics`, `Hub.Metrics`) instead of reintroducing global wrappers
- treat `x/websocket` as the app-facing websocket transport surface
