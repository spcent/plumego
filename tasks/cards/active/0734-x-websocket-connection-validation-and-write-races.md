# Card 0734

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: x/websocket
Owned Files:
- `x/websocket/conn.go`
- `x/websocket/writer.go`
- `x/websocket/server.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0733

Goal:
- Remove panic-prone connection configuration paths and close/write races from the WebSocket connection layer.

Problem:
`NewConn` accepts invalid public inputs that can panic or create unusable connections. `SetPingPeriod` and `SetPongWait` can make writer goroutines panic through zero or negative ticker durations. `WriteMessageContext` can enqueue messages after `Close` because `sendQueue` remains open and the close path races with the send select.

Scope:
- Add explicit validation for public connection construction and mutable timing settings.
- Introduce an error-returning constructor or validation helper if preserving existing API shape requires compatibility.
- Make writes after close reliably return `ErrConnectionClosed`.
- Add focused tests for negative queue size, nil connection, invalid send behavior, invalid ping/pong durations, and write-after-close.

Non-goals:
- Do not redesign the frame writer or worker pool.
- Do not change Hub broadcast queue semantics except as needed to observe write errors.
- Do not remove exported symbols without following `AGENTS.md` symbol-change protocol.

Files:
- `x/websocket/conn.go`
- `x/websocket/writer.go`
- `x/websocket/server.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required if constructor behavior, errors, or timing setter behavior changes.

Done Definition:
- Invalid public connection config returns explicit errors or is rejected before goroutine startup.
- Invalid ping/pong durations cannot panic.
- Writes after close are deterministic and covered by race-safe tests.

Outcome:
-
