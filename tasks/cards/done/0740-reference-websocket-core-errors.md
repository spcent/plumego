# 0740 - reference WebSocket Core Errors

State: done
Priority: P1
Primary Module: core

## Goal

Bring the WebSocket reference app into line with the core error-handling
contract for middleware registration and shutdown.

## Scope

- Handle `core.Use` errors in the WebSocket reference constructor.
- Explicitly handle or discard `Core.Shutdown` errors in the WebSocket shutdown
  defer.
- Preserve middleware order and WebSocket shutdown behavior.

## Non-goals

- Do not change WebSocket extension behavior.
- Do not alter route registration.
- Do not add new dependencies.

## Files

- `reference/with-websocket/internal/app/app.go`

## Tests

- `go test -timeout 20s ./reference/...`
- `go run ./internal/checks/reference-layout`

## Docs Sync

Not required.

## Done Definition

- WebSocket reference app no longer silently ignores core middleware or shutdown
  errors.
- Reference tests and layout checks pass.

## Outcome

- Updated WebSocket reference constructor to return wrapped `core.Use` errors.
- Made deferred `Core.Shutdown` discard explicit, preserving existing WebSocket
  shutdown behavior.
- Verified with `go test -timeout 20s ./reference/...` and
  `go run ./internal/checks/reference-layout`.
