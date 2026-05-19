# Card 1291

Milestone:
Recipe: specs/change-recipes/extension-stability.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files: x/websocket/writer.go, x/websocket/hub.go, x/websocket/writer_test.go, x/websocket/hub_test.go
Depends On: 0765

Goal:
Avoid unnecessary large payload copies when outbound queues are already full or configured to drop.

Scope:
- Defer single-connection payload ownership copy until a send path can accept the message.
- Avoid broadcast payload cloning when no destination accepts the message.
- Preserve the stable ownership contract that accepted sends are insulated from caller slice mutation.
- Add focused tests for dropped/full-queue paths.

Non-goals:
- Change the documented accepted-send ownership contract.
- Add zero-copy send semantics.
- Change message fragmentation behavior.

Files:
- x/websocket/writer.go
- x/websocket/hub.go
- x/websocket/writer_test.go
- x/websocket/hub_test.go

Tests:
- go test -timeout 20s ./x/websocket/...

Docs Sync:
- None expected unless ownership wording changes.

Done Definition:
- Rejected/drop sends do not copy large payloads unnecessarily.
- Accepted sends remain caller-slice safe.
- Module tests pass.

Outcome:
Deferred outbound payload snapshots until a connection or hub queue has capacity
to accept the message. Full `SendDrop` queues and full broadcast job queues now
avoid copying large caller payloads while accepted sends still snapshot bytes.
Added focused allocation tests for the full-queue paths.
