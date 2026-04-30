# Card 0735

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: x/websocket
Owned Files:
- `x/websocket/conn.go`
- `x/websocket/stream.go`
- `x/websocket/errors.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0734

Goal:
- Tighten WebSocket frame parsing to reject malformed RFC6455 inputs before the package can be considered stable.

Problem:
The frame reader handles masking, payload length, and basic control frame checks, but it does not fully validate RSV bits, reserved opcodes, minimal length encoding, close payload shape, or continuation sequencing in the low-level reader path.

Scope:
- Reject non-zero RSV bits unless an explicitly supported extension exists.
- Reject reserved or unknown opcodes.
- Reject non-minimal payload length encodings.
- Validate close frame payload length and status code shape.
- Enforce continuation sequencing consistently between `ReadMessage`, `ReadMessageStream`, and frame-level reads.
- Add negative tests for malformed frames and valid baseline frames.

Non-goals:
- Do not implement compression or extension negotiation.
- Do not add third-party WebSocket libraries.
- Do not change high-level broadcast behavior.

Files:
- `x/websocket/conn.go`
- `x/websocket/stream.go`
- `x/websocket/errors.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required if protocol rejection behavior or error names change.

Done Definition:
- Malformed RSV, opcode, length, close payload, and continuation cases are rejected by tests.
- Valid text, binary, ping, pong, close, and fragmented messages still pass.
- No protocol-compliance change weakens read limits or masking enforcement.

Outcome:
-
