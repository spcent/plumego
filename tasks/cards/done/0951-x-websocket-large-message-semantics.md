# Card 0951

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/stream.go`
- `x/websocket/server.go`
- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0735

Goal:
- Make large-message and streaming behavior honest, bounded, and testable.

Problem:
`ReadMessageStream` is named like a streaming API, but fragmented messages are accumulated in memory before being exposed to the caller. The default server read loop then copies each full message into memory again before validation and broadcast.

Scope:
- Choose and implement one stable semantic: true incremental streaming, or bounded whole-message reads with clearer naming and documentation.
- Ensure the server read loop applies message limits before unbounded accumulation.
- Keep `ReadLimit` and `MessageValidation` behavior consistent across fragmented and unfragmented messages.
- Add tests for fragmented messages at, below, and above configured limits.

Non-goals:
- Do not implement backpressure across remote clients or persistent storage.
- Do not add compression support.
- Do not redesign Hub worker queues.

Files:
- `x/websocket/stream.go`
- `x/websocket/server.go`
- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required. Document whether reads are whole-message bounded reads or true streaming reads.

Done Definition:
- Large fragmented messages cannot bypass configured limits.
- API names and docs match actual buffering behavior.
- Tests cover both fragmented and unfragmented limit enforcement.

Outcome:
- Chose bounded whole-message semantics for the current stable-readiness path.
- Enforced `ReadLimit` across the total fragmented message payload in `ReadMessageStream`/`ReadMessage`.
- Documented that `ReadMessageStream` reads continuation frames incrementally but is not an unbounded streaming bypass.
- Added fragmented and unfragmented tests at and above configured read limits.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go test -race -timeout 60s ./x/websocket/...`
  - `go vet ./x/websocket/...`
