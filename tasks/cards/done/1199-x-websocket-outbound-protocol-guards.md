# 1199 - x/websocket outbound protocol guards

Status: done
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

## Outcome

- Added data opcode validation for public queued sends and result-returning hub
  broadcast APIs.
- Added write-side close frame validation for close status, reason UTF-8, and
  control payload size.
- Added negative tests for invalid data opcodes and invalid close payloads.
- Documented public data-send opcode and close-frame validation behavior.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
