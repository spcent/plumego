# Card 0775

Milestone: M-003
Recipe: specs/change-recipes/implementation.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/hub_lifecycle_test.go`
- `x/websocket/websocket_test.go`
- `x/websocket/module.yaml`

Goal:
- Make broadcast validation and fanout result semantics clear and consistent.

Scope:
- Let public `Hub` broadcast/read APIs use the configured room-name validator when one is supplied.
- Make `BroadcastResult` distinguish invalid input and stopped hub from empty/no-target broadcasts.
- Clarify that broadcast counts represent enqueue results, not network delivery.
- Keep admin broadcast and public hub broadcast behavior consistent.

Non-goals:
- Do not add durable delivery acknowledgements.
- Do not turn the hub into a generic event bus.

Files:
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/hub_lifecycle_test.go`
- `x/websocket/websocket_test.go`
- `x/websocket/module.yaml`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for `BroadcastResult` public field semantics.

Done Definition:
- Custom room validators work through the `Server` and `Hub` broadcast paths.
- Invalid/stopped broadcasts are observable in the result.

Outcome:
- Added `Invalid` and `Stopped` result flags and included them in
  `BroadcastResult.Rejected`.
- Added `HubConfig.RoomNameValidator` and applied it to public hub room APIs.
- Passed top-level `WebSocketConfig.RoomNameValidator` into the hub runtime.
- Verified with `go test -timeout 20s ./x/websocket/...` and `go vet
  ./x/websocket/...`.
