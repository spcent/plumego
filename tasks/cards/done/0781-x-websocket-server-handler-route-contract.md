# Card 0781

Milestone: M-003
Recipe: specs/change-recipes/implementation.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/websocket.go`
- `x/websocket/server.go`
- `x/websocket/websocket_test.go`
- `x/websocket/ws_test.go`
- `docs/modules/x-websocket/README.md`

Goal:
- Make the top-level `Server` route registration support both generic message handling and built-in room fanout explicitly.

Scope:
- Add a `MessageHandler` hook to `WebSocketConfig`.
- Route `RegisterRoutes` through `ServeWSWithConfig` when a custom handler is provided, and through `ServeRoomFanoutWS` only when no custom handler is configured.
- Make `ServeRoomFanoutWS` reject or document any conflicting `OnMessage` input instead of silently overwriting it.
- Add focused tests for custom handler routing and fanout helper behavior.

Non-goals:
- Do not add application-level delivery acknowledgements.
- Do not change authentication or room policy in this card.
- Do not promote `x/websocket` out of experimental status.

Files:
- `x/websocket/websocket.go`
- `x/websocket/server.go`
- `x/websocket/websocket_test.go`
- `x/websocket/ws_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for `WebSocketConfig.OnMessage` and fanout helper semantics.

Done Definition:
- `New(WebSocketConfig{OnMessage: ...})` registers a generic websocket route.
- The built-in fanout behavior remains available and explicit.

Outcome:
- Added `WebSocketConfig.OnMessage` for custom top-level server route handling.
- Kept default `RegisterRoutes` behavior as explicit room fanout when no custom
  handler is configured.
- Made `ServeRoomFanoutWS` reject conflicting `OnMessage` input instead of
  silently overwriting it.
- Verified with `go test -timeout 20s ./x/websocket/...` and `go vet
  ./x/websocket/...`.
