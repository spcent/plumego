# x/websocket

## Purpose

`x/websocket` owns websocket server helpers and explicit route registration for websocket transport.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

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
- `NewHub`
- `NewHubWithConfig`
- `ServeWSWithAuth`
- `ServeWSWithConfig`

## Main risks when changing this module

- websocket auth regression
- broadcast behavior regression
- connection lifecycle regression

## Boundary rules

- keep websocket setup explicit and out of `core`; do not add hidden goroutines or global state at import time
- keep transport concerns (`ServeWSWithAuth`, `ServeWSWithConfig`) inside `x/websocket`; do not push connection-level logic into stable roots or middleware
- keep auth and broadcast gates reviewable and testable in isolation
- handle room-password setup errors explicitly; do not hide hash failures behind log-only behavior
- keep security metrics instance-scoped (`SecureRoomAuth.GetMetrics`, `Hub.Metrics`) instead of reintroducing global wrappers
- treat `x/websocket` as the app-facing websocket transport surface; app-level session management belongs in the calling handler

## Current test coverage

- connection configuration (read limit, ping period, pong wait)
- `Hub` lifecycle: `Stop` idempotency, `Shutdown` (empty and with connections, context cancellation), `Join`/`TryJoin`/`Leave`/`RemoveConn` lifecycle, `RangeConns` iteration and early return
- capacity errors: `ErrHubFull`, `ErrRoomFull`, `ErrHubStopped` from `TryJoin`/`CanJoin` after stop or at limit
- broadcast: `BroadcastRoom`, `BroadcastAll` (positive path and no-op after stop), race-condition coverage under concurrent goroutines
- security: `ValidateSecurityConfig`, `ValidateWebSocketKey`, `ValidateRoomPassword`, `SecureRoomAuth`, security metrics, connection limit enforcement
- validation: text message sanitization, dangerous-pattern detection, control-character handling
- server setup: `ServeWSWithAuth` (method-not-allowed, bad-request, bad-room-password), `ServeWSWithConfig` invalid-config rejection, config normalization

## Canonical change shape

- keep websocket setup explicit and out of `core`
- keep auth and broadcast gates reviewable
- handle room-password setup errors explicitly; do not hide hash failures behind log-only behavior
- keep security metrics instance-scoped (`SecureRoomAuth.GetMetrics`, `Hub.Metrics`) instead of reintroducing global wrappers
- treat `x/websocket` as the app-facing websocket transport surface
