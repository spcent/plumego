# Card 1340

Milestone: M-003
Recipe: specs/change-recipes/implementation.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/auth.go`
- `x/websocket/server.go`
- `x/websocket/server_config_test.go`
- `x/websocket/websocket_extended_test.go`
- `x/websocket/module.yaml`

Goal:
- Make authentication and room authorization ordering explicit enough for stable production use.

Scope:
- Verify bearer/query tokens before room authorization when a token is required or supplied.
- Add a request-aware room authorization hook that can inspect request context and authenticated user claims.
- Preserve the simple room-password helper as a basic helper, with docs that nil room auth allows rooms.
- Keep query token support behind the existing explicit opt-in.

Non-goals:
- Do not implement OIDC, issuer, audience, or required-claims policy in the built-in HS256 helper.
- Do not change origin defaults in this card.

Files:
- `x/websocket/auth.go`
- `x/websocket/server.go`
- `x/websocket/server_config_test.go`
- `x/websocket/websocket_extended_test.go`
- `x/websocket/module.yaml`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for new public auth hook semantics.

Done Definition:
- Token verification establishes `Conn.UserInfo` before context-aware room authorization runs.
- Tests cover claims-aware room authorization and legacy room-password behavior.

Outcome:
- Added `RoomAuthorization` and `RoomRequestAuthorizer` for request-aware,
  claims-aware room policy.
- Reordered handshake auth so supplied/required tokens are verified before room
  password or room policy checks.
- Preserved legacy `RoomAuthorizer.CheckRoomPassword` behavior for simple room
  password helpers.
- Verified with `go test -timeout 20s ./x/websocket/...`, `go vet
  ./x/websocket/...`, and `go run ./internal/checks/module-manifests`.
