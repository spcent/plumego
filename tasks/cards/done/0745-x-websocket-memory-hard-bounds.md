# 0745 - x/websocket memory hard bounds

Status: done
Priority: P0
Primary module: `x/websocket`

## Goal

Close remaining memory-retention and oversized allocation gaps.

## Scope

- Add an upper hard cap for `SetReadLimit` / `ReadLimit`-derived inbound
  payload allocation.
- Cap retained `connListPool` slice capacity.
- Add tests for excessive read limits and large broadcast snapshot pool reuse.

## Non-goals

- True zero-copy streaming.
- Changing default read limit.
- New dependencies.

## Files

- `x/websocket/conn.go`
- `x/websocket/server.go`
- `x/websocket/hub.go`
- `x/websocket/protocol_test.go`
- `x/websocket/hub_lifecycle_test.go`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document the hard inbound message cap.

## Done Definition

- Unreasonably large read limits are rejected.
- Large broadcast snapshot slices are not retained in the pool.
- Validation passes.

## Outcome

- Added a 64 MiB hard cap for connection, server, top-level, and
  auth-derived read limits.
- Prevented oversized hub connection snapshot slices from being retained in the
  pool after large broadcasts.
- Documented the default and hard inbound message bounds.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
