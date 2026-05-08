# Card 1261

Milestone: M-003
Recipe: specs/change-recipes/api-contract.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
- `x/websocket/module.yaml`
Depends On: 0762

Goal:
- Separate pure WebSocket transport serving from the built-in room fanout helper.

Problem:
`ServeWSWithConfig` currently upgrades, authenticates, joins a room, and then broadcasts every client message back to the room. That mixes transport lifecycle with chat/fanout product behavior.

Scope:
- Change the transport entrypoint to delegate message handling through explicit callbacks or a small handler interface.
- Move the current room fanout loop into an explicitly named helper such as `ServeRoomFanoutWS` or an equivalent stable-candidate API.
- Keep route registration explicit and readable.
- Update tests to cover both custom message handling and room fanout helper behavior.

Non-goals:
- Do not add a generic event bus.
- Do not move websocket setup into stable roots.
- Do not add non-stdlib dependencies.

Files:
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
- `x/websocket/module.yaml`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for the new transport-vs-fanout contract and examples.

Done Definition:
- The low-level serve path does not hard-code room fanout behavior.
- The old fanout behavior remains available only through an explicitly named helper.
- Docs show both custom message handling and room fanout wiring.

Outcome:
- Added `Message` and `MessageHandler` so `ServeWSWithConfig` delegates
  validated client messages instead of hard-coding room fanout.
- Added `ServeRoomFanoutWS` for the previous room broadcast behavior.
- Updated `Server.RegisterRoutes` and integration tests to use the explicit
  fanout helper.
- Added custom handler coverage for the low-level transport serve path.
