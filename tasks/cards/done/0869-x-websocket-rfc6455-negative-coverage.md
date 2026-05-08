# 0869 - x/websocket RFC6455 negative coverage

Status: done
Priority: P0
Primary module: `x/websocket`

## Problem

Frame parsing still accepts or ignores several invalid RFC6455 states such as
RSV bits, unknown opcodes, invalid continuation order, non-minimal length
encoding, and malformed close payloads.

## Scope

- Reject RSV bits, unknown opcodes, invalid continuation ordering, and
  non-minimal length encodings.
- Validate close frame payload length, status codes, and UTF-8 reason text.
- Add focused negative tests for each protocol condition.
- Keep the implementation dependency-free.

## Out of Scope

- Replacing the WebSocket implementation with a third-party library.
- Full fuzz harness adoption.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/dependency-rules`

## Outcome

- Rejected RSV bits, unknown opcodes, continuation frames before a data
  message, and non-minimal payload length encodings.
- Validated close frame payload length, status codes, and UTF-8 reason text.
- Added focused negative protocol tests for each repaired RFC6455 path.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go run ./internal/checks/dependency-rules`
