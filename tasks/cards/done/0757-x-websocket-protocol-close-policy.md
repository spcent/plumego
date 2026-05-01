# Card 0757

Milestone: M-003
Recipe: specs/change-recipes/security-hardening.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/conn.go`
- `x/websocket/server.go`
- `x/websocket/errors.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0756

Goal:
- Align protocol and validation failures with stable WebSocket close semantics.

Problem:
The server read loop silently drops invalid text messages and continues. Public write APIs also accept arbitrary opcodes. Stable behavior should either deliver valid data or close with a clear RFC6455 status.

Scope:
- Reject invalid outbound opcodes before enqueueing application messages.
- On invalid text payloads, send the appropriate close frame (`1007`, `1008`, or `1009`) and close.
- On protocol read errors, send `1002` or the nearest safe close code where possible.
- Add focused tests for invalid outbound opcodes and read-loop validation close behavior.

Non-goals:
- Do not add compression, subprotocol negotiation, or extension support.
- Do not change admin broadcast semantics.
- Do not add non-stdlib websocket dependencies.

Files:
- `x/websocket/conn.go`
- `x/websocket/server.go`
- `x/websocket/errors.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for close-code behavior and outbound opcode validation.

Done Definition:
- Public writes reject unsupported data/control opcodes.
- Invalid inbound text validation produces a close frame instead of silent drop.
- Protocol read failures close the connection with a documented close policy.

Outcome:
- Added public write opcode validation so application writes accept text and binary messages only.
- Added close-frame mapping for read/protocol and message validation failures.
- Changed invalid inbound text handling from silent drop to close with `1007`.
- Documented the close-code policy for protocol, payload, policy, and size failures.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go build ./...`
