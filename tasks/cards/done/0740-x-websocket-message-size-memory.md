# 0740 - x/websocket message size and memory bounds

Status: done
Priority: P0
Primary module: `x/websocket`

## Goal

Enforce full-message size limits across fragmented messages and avoid retaining
large pooled buffers.

## Scope

- Track cumulative message size across continuation frames.
- Reject fragmented messages that exceed the effective read/message limit.
- Add buffer-pool capacity caps so large buffers are not retained.
- Add focused tests for fragmented over-limit messages and pool-safe behavior.

## Non-goals

- True zero-copy streaming.
- Changing public payload schemas.

## Files

- `x/websocket/stream.go`
- `x/websocket/server.go`
- `x/websocket/conn.go`
- `x/websocket/protocol_test.go`
- `x/websocket/websocket_extended_test.go`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Clarify that limits apply to complete messages, including fragmented messages.

## Done Definition

- Fragmented messages cannot bypass configured limits.
- Pooled buffers above the cap are discarded instead of retained.
- Validation passes.

## Outcome

- Added cumulative message-size accounting across continuation frames so
  fragmented messages cannot exceed the effective read limit.
- Added a 64 KiB cap for retained pooled message buffers; oversized buffers are
  discarded for both complete-message reads and stream readers.
- Updated server copy paths to return buffers through the capped pool helper.
- Documented that read limits apply to complete fragmented messages and that
  oversized pooled buffers are not retained.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
- `go run ./internal/checks/module-manifests`
