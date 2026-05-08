# Card 0776

Milestone: M-003
Recipe: specs/change-recipes/implementation.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/hub.go`
- `x/websocket/hub_lifecycle_test.go`
- `x/websocket/hub_stop_test.go`
- `x/websocket/race_test.go`

Goal:
- Make `Hub.Stop` and `Hub.Shutdown` terminal lifecycle semantics clear, bounded, and observable.

Scope:
- Define whether stopped hubs retain room registrations.
- Ensure cancelled shutdown leaves the hub in a documented terminal state.
- Keep worker drain bounded and document the timeout policy.
- Add focused lifecycle tests for stop, shutdown, cancellation, and metrics.

Non-goals:
- Do not implement full WebSocket closing handshake during process shutdown.
- Do not change `Conn.Close` transport ownership.

Files:
- `x/websocket/hub.go`
- `x/websocket/hub_lifecycle_test.go`
- `x/websocket/hub_stop_test.go`
- `x/websocket/race_test.go`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`

Docs Sync:
- Required if lifecycle comments change exported behavior.

Done Definition:
- Stop/shutdown behavior is deterministic and covered by tests.
- No stopped hub retains misleading registration metrics.

Outcome:
- Defined `Hub.Stop` as a terminal state that stops workers and clears room
  registrations.
- Ensured cancelled `Shutdown` stops the hub and clears registration metrics.
- Added tests for stop metric cleanup and cancelled shutdown terminal state.
- Verified with `go test -timeout 20s ./x/websocket/...` and `go test -race
  -timeout 60s ./x/websocket/...`.
