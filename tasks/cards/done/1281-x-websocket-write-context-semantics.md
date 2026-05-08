# Card 1281

Milestone:
Recipe: specs/change-recipes/extension-stability.yaml
Priority: P0
State: active
Primary Module: x/websocket
Owned Files: x/websocket/writer.go, x/websocket/conn.go, x/websocket/writer_test.go, docs/modules/x-websocket/README.md
Depends On:

Goal:
Make `WriteMessageContext` return values and deadlines match stable send semantics.

Scope:
- Remove the fast-path race where a successfully enqueued message can return `ErrConnClosed`.
- Carry absolute context deadline information through queued writes where feasible.
- Ensure socket write deadline policy is clear for queued and fragmented writes.
- Add focused tests for close races and deadline propagation.

Non-goals:
- Implement per-byte streaming writes.
- Change public opcode constants.
- Add external websocket dependencies.

Files:
- x/websocket/writer.go
- x/websocket/conn.go
- x/websocket/writer_test.go
- docs/modules/x-websocket/README.md

Tests:
- go test -timeout 20s ./x/websocket/...

Docs Sync:
- Document context deadline and send timeout interaction.

Done Definition:
- A nil error from `WriteMessageContext` means the message was accepted for delivery.
- A close race after acceptance does not produce a misleading retry signal.
- Deadline behavior is tested and documented.

Outcome:
Removed the misleading post-enqueue close check from the fast path and carried
absolute context deadlines on queued outbound messages. Writer execution now
counts queue wait time against context deadlines and skips expired queued writes.
Added focused tests and updated module docs.
