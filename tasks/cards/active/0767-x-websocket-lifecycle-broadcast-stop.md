# Card 0767

Milestone: M-003
Recipe: specs/change-recipes/bugfix.yaml
Priority: P0
State: todo
Primary Module: x/websocket
Owned Files:
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/hub_lifecycle_test.go`
- `x/websocket/hub_stop_test.go`
- `x/websocket/race_test.go`
Depends On: 0766

Goal:
- Make hub shutdown and broadcast behavior deterministic under concurrent stop.

Problem:
`Shutdown(ctx)` does not define nil context behavior and may leave partial state after cancellation. Broadcast calls only check `stopped` at entry, so a concurrent `Stop` can race with snapshot/dispatch and enqueue to a queue with no active workers.

Scope:
- Define and implement `Shutdown(nil)` behavior.
- Define partial shutdown semantics when context is cancelled.
- Prevent broadcast enqueue after workers have been stopped.
- Ensure stopped hubs return observable broadcast results instead of ambiguous zero values where needed.
- Add concurrent stop/broadcast regression tests.

Non-goals:
- Do not redesign the hub as a generic event bus.
- Do not add external synchronization dependencies.
- Do not change auth behavior.

Files:
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/hub_lifecycle_test.go`
- `x/websocket/hub_stop_test.go`
- `x/websocket/race_test.go`

Tests:
- `go test -race -timeout 60s ./x/websocket/...`
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required if shutdown or stopped-broadcast semantics change.

Done Definition:
- `Shutdown(nil)` cannot panic.
- Stop/broadcast races cannot report delivery for jobs that no worker can consume.
- Cancellation behavior is documented and covered by tests.

Outcome:
-
