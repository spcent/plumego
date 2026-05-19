# Card 1307

Milestone: M-003
Recipe: specs/change-recipes/bugfix.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/conn.go`
- `x/websocket/writer.go`
- `x/websocket/server.go`
- `x/websocket/writer_pump_test.go`
- `x/websocket/websocket_extended_test.go`
Depends On: 0767

Goal:
- Bound network writes and make close-frame semantics match the API contract.

Problem:
`SendTimeout` controls enqueue behavior, not socket writes. `writeFrame` has no write deadline, so slow clients can block writer goroutines. `WriteClose` says it performs a proper closing handshake, but it writes a close frame and then immediately closes TCP.

Scope:
- Add explicit write deadline configuration for frame writes.
- Apply write deadlines in `writeFrame` or the writer pump without breaking tests.
- Clarify or implement graceful close semantics for `WriteClose`.
- Add slow-writer and close-frame behavior tests.

Non-goals:
- Do not introduce a third-party websocket library.
- Do not change read-limit semantics.
- Do not hide write errors.

Files:
- `x/websocket/conn.go`
- `x/websocket/writer.go`
- `x/websocket/server.go`
- `x/websocket/writer_pump_test.go`
- `x/websocket/websocket_extended_test.go`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for send timeout, write deadline, and close behavior.

Done Definition:
- Slow or blocked network writes are bounded by a documented deadline.
- `WriteClose` implementation and Go doc use the same semantics.
- Tests cover write timeout and close behavior.

Outcome:
- Added configurable frame write deadlines through `WebSocketConfig`,
  `ServerConfig`, and `Conn.SetWriteTimeout`.
- Applied write deadlines inside `writeFrame`.
- Updated `WriteClose` documentation to describe best-effort close-frame
  delivery followed by TCP close.
- Added write deadline coverage for `WriteClose`.
