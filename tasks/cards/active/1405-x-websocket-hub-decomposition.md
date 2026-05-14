# Card 1405

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/websocket
Owned Files:
- x/websocket/hub.go
- x/websocket/hub_rooms.go
- x/websocket/hub_broadcast.go
- x/websocket/hub_test.go
- x/websocket/race_test.go
Depends On:
- 1404

Goal:
- Reduce `x/websocket` hub edit radius while preserving the beta public API.

Scope:
- Move room membership helpers into `hub_rooms.go`.
- Move broadcast and queue fan-out helpers into `hub_broadcast.go`.
- Preserve exported hub constructors, connection lifecycle, room semantics, and bounded queue behavior.
- Keep compatibility aliases intact.

Non-goals:
- Do not change public hub types or constructors.
- Do not change handshake, auth, or stream APIs.
- Do not alter queue capacity or close semantics.

Files:
- x/websocket/hub.go
- x/websocket/hub_rooms.go
- x/websocket/hub_broadcast.go
- x/websocket/hub_test.go
- x/websocket/race_test.go

Tests:
- go test -timeout 30s ./x/websocket
- go vet ./x/websocket
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless public hub comments move or change.

Done Definition:
- Hub room and broadcast logic have separate file ownership.
- Existing hub lifecycle, room, queue, and race tests pass.
- No exported API change is introduced.

Outcome:

