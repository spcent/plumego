# 1033 - x/websocket handshake auth and metrics

Status: done
Priority: P0
Primary module: `x/websocket`

## Goal

Keep handshake security failure ordering and rejection metrics consistent.

## Scope

- Authenticate tokens before exposing room/hub capacity failures.
- Count `CanJoin` capacity/rate rejections from handshake paths.
- Validate admin broadcast `room` query with `ValidateRoomName`.
- Add focused tests for auth-before-capacity, rejection metrics, and invalid
  admin room names.

## Non-goals

- Changing token algorithms.
- Changing route registration shape.
- Changing broadcast payload schema.

## Files

- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/server_config_test.go`
- `x/websocket/websocket_test.go`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document only if status codes or admin broadcast behavior changes.

## Done Definition

- Unauthorized clients do not learn capacity state.
- Handshake capacity failures increment rejection metrics.
- Admin broadcast rejects invalid room names.
- Validation passes.

## Outcome

- Moved token authentication before room password and capacity checks so
  unauthorized clients do not learn room/hub capacity state.
- Guarded bearer-token requests with nil `TokenAuth` to return structured
  invalid-token errors instead of panicking.
- Counted handshake capacity failures in `RejectedTotal`.
- Validated admin broadcast `room` query values with `ValidateRoomName`.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
