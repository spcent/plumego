# Card 0741

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: x/websocket
Owned Files:
- `x/websocket/websocket.go`
- `x/websocket/server.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`
Depends On: 0740

Goal:
- Make WebSocket route registration and handshake validation fail visibly and RFC6455-compliant.

Problem:
The handshake validates method, upgrade, and `Sec-WebSocket-Key`, but does not require `Sec-WebSocket-Version: 13`. `RegisterRoutes` also assumes a valid registrar, route path, and hub, so assembly mistakes can panic or silently install invalid behavior instead of failing during wiring.

Scope:
- Require `Sec-WebSocket-Version: 13` during handshake and reject missing or unsupported versions.
- Change route registration to return an error for nil registrar, nil hub, empty `WSRoutePath`, empty `BroadcastPath` when enabled, or invalid handler configuration.
- Update all callers for the new registration signature; callers must handle or explicitly discard returned errors.
- Keep route wiring one method, one path, one handler per line after the signature change.
- Follow `AGENTS.md §7.1` for exported registration signature changes.

Non-goals:
- Do not change frame parsing beyond the handshake-version requirement.
- Do not add alternate WebSocket protocol versions.
- Do not alter core/router behavior.

Files:
- `x/websocket/websocket.go`
- `x/websocket/server.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for registration error handling and handshake requirements.

Done Definition:
- Missing or non-13 `Sec-WebSocket-Version` is rejected in tests.
- Invalid route registration inputs return explicit errors before route installation.
- Every caller of the changed registration API is updated in the same commit.
- `go build ./...` proves no silent discarded registration errors remain.
