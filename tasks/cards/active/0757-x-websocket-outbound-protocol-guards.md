# 0757 - x/websocket outbound protocol guards

Status: active
Priority: P0
Primary module: `x/websocket`

## Goal

Reject invalid outbound websocket protocol inputs before writing frames.

## Scope

- Validate public data-send opcodes so only text and binary use
  `WriteMessage`/broadcast APIs.
- Validate close frame status code, reason UTF-8, and 125-byte payload limit.
- Add focused negative tests.

## Non-goals

- Adding public ping/pong APIs.
- Changing inbound frame parsing.
- Changing message validation policy.

## Files

- `x/websocket/conn.go`
- `x/websocket/writer.go`
- `x/websocket/hub.go`
- `x/websocket/errors.go`
- `x/websocket/writer_pump_test.go`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document that public data-send APIs accept only text and binary opcodes.

## Done Definition

- Invalid data opcodes are rejected before enqueue/broadcast.
- Invalid close frames are rejected before network writes.
- Validation passes.
