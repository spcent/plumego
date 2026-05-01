# Card 0756

Milestone: M-003
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P0
State: active
Primary Module: x/websocket
Owned Files:
- `x/websocket/conn.go`
- `x/websocket/stream.go`
- `x/websocket/errors.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0755

Goal:
- Make connection-level lifecycle APIs fail-visible and avoid misleading reader/close behavior.

Problem:
`SetReadLimit` accepts nonsensical values without returning an error, `WriteClose` discards close-frame write errors, and `ReadMessageReader.Close` returns the reader to the pool even if the caller has not consumed a fragmented message.

Scope:
- Change read-limit mutation to return validation errors and update all callers.
- Return close-frame write errors from `WriteClose`.
- Define early reader close behavior: either drain safely or close the parent connection when a message is not fully consumed.
- Add tests for invalid read limits, close-frame write failures, and early reader close behavior.
- Follow `AGENTS.md §7.1` for exported signature changes.

Non-goals:
- Do not implement unbounded streaming.
- Do not add websocket compression or extensions.
- Do not change hub shutdown behavior in this card.

Files:
- `x/websocket/conn.go`
- `x/websocket/stream.go`
- `x/websocket/errors.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for read limits, close-frame errors, and reader close semantics.

Done Definition:
- Invalid read limits fail visibly.
- `WriteClose` returns the frame write failure when the close frame cannot be sent.
- Closing an unfinished message reader cannot leave the connection in an ambiguous continuation state.

Outcome:
-
