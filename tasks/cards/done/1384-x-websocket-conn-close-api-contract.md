# Card 1384

Milestone: M-003
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/conn.go`
- `x/websocket/errors.go`
- `x/websocket/server.go`
- `x/websocket/protocol_compliance_test.go`
- `docs/modules/x-websocket/README.md`
Depends On:
- M-003

Goal:
- Clarify low-level connection and handler close contracts before API freeze.

Scope:
- Document `Conn` and `NewConnE` as server-side post-handshake APIs with goroutine ownership.
- Make ping/pong timing errors accurately describe the invariant.
- Make handler close-error construction fail-visible or otherwise explicitly validated.
- Keep `WriteClose` best-effort close-frame behavior.

Non-goals:
- Do not add client-side websocket support.
- Do not implement a full peer close handshake.

Files:
- `x/websocket/conn.go`
- `x/websocket/errors.go`
- `x/websocket/server.go`
- `x/websocket/protocol_compliance_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for low-level connection ownership and close-error semantics.

Done Definition:
- Low-level connection APIs no longer look like generic client/server WebSocket wrappers.
- Invalid handler close errors are visible at construction or documented validation points.

Outcome:
- `Conn`, `NewConnE`, `WriteClose`, and bounded reader semantics are documented
  as server-side post-handshake contracts.
- Close-frame validation and write-error behavior are covered by websocket
  tests.

Validation:
- go test -timeout 20s ./x/websocket/...
- go vet ./x/websocket/...
