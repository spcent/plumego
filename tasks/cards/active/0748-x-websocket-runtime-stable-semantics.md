# 0748 - x/websocket runtime stable semantics

Status: active
Priority: P0
Primary module: `x/websocket`

## Goal

Close the remaining behavior-level gaps from the websocket stable-hardening
audit that are not already covered by cards 0742-0747.

## Scope

- Prevent user-provided `SecurityEventHandler` from blocking `Hub.Stop` or
  `Hub.Shutdown`.
- Make `Conn.SetReadLimit(0)` match the config-layer contract by restoring the
  default read limit instead of rejecting every non-empty frame.
- Add regression tests for blocking security handlers during stop and zero read
  limit behavior.

## Non-goals

- New event exporter or worker pool.
- Changing default read limit size.
- Promoting `x/websocket` beyond experimental.

## Files

- `x/websocket/conn.go`
- `x/websocket/hub.go`
- `x/websocket/protocol_test.go`
- `x/websocket/hub_lifecycle_test.go`
- `docs/modules/x-websocket/README.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document `SetReadLimit(0)` and async security event handler shutdown
semantics.

## Done Definition

- A stuck security event handler cannot block hub stop/shutdown.
- `SetReadLimit(0)` restores the default read limit.
- Validation passes.
