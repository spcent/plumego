# Card 0777

Milestone: M-003
Recipe: specs/change-recipes/implementation.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/conn.go`
- `x/websocket/writer.go`
- `x/websocket/errors.go`
- `x/websocket/protocol_compliance_test.go`
- `x/websocket/writer_pump_test.go`

Goal:
- Close protocol and async-send contract gaps in `Conn`.

Scope:
- Validate outbound close status codes and close reason payload size/UTF-8 before writing.
- Copy application data on enqueue so async writes do not alias caller-owned buffers.
- Enforce or normalize ping/pong timing invariants to avoid premature heartbeat closes.
- Keep network write deadlines active for frame writes.

Non-goals:
- Do not add client-side WebSocket semantics.
- Do not add external WebSocket dependencies.

Files:
- `x/websocket/conn.go`
- `x/websocket/writer.go`
- `x/websocket/errors.go`
- `x/websocket/protocol_compliance_test.go`
- `x/websocket/writer_pump_test.go`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for `WriteMessage` and `WriteClose` ownership/error semantics.

Done Definition:
- Outbound close frames cannot violate RFC 6455 close payload rules.
- Tests prove queued sends own their payload bytes.

Outcome:
- Added explicit outbound close-frame validation for close codes, UTF-8 reason
  text, and RFC 6455 control payload length.
- Added send-queue payload copying so async writes own application bytes.
- Enforced ping period and pong wait ordering through public setters.
- Verified with `go test -timeout 20s ./x/websocket/...`, `go test -race
  -timeout 60s ./x/websocket/...`, `go vet ./x/websocket/...`, and `go run
  ./internal/checks/module-manifests`.
