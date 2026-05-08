# 1211 - x/websocket send ownership deadlines

Status: done
Priority: P0
Primary module: `x/websocket`

## Goal

Make outbound send behavior deterministic around caller-owned slices, close
races, and slow socket writes.

## Scope

- Ensure queued outbound messages own their payload bytes.
- Make `WriteMessageContext` fast path observe connection close.
- Ensure hub worker writes always use a finite write deadline.
- Add focused tests for payload ownership and close-race behavior where
  practical.

## Non-goals

- Introducing streaming outbound writes.
- Changing send queue sizing defaults.
- Changing broadcast async delivery semantics.

## Files

- `x/websocket/writer.go`
- `x/websocket/conn.go`
- `x/websocket/hub.go`
- `x/websocket/writer_pump_test.go`
- `docs/modules/x-websocket/README.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document outbound payload ownership and write-deadline semantics.

## Done Definition

- Queued outbound sends cannot observe caller-side slice mutation.
- Close races do not return successful enqueue after close is visible.
- Worker writes have bounded socket deadlines.
- Validation passes.

## Outcome

- Queued outbound messages now snapshot caller payload bytes before enqueue.
- `WriteMessageContext` observes a closed connection channel on the fast path.
- Outbound queued messages carry a write timeout derived from context deadline,
  configured send timeout, or the default finite hub write timeout.
- Socket frame writes now set a finite deadline even when `SendTimeout` is zero.
- Added tests for payload ownership, close-visible fast path, context-derived
  write timeout, and default write deadlines.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
