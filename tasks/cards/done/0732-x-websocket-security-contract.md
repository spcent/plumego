# Card 0732

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/auth.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On:

Goal:
- Make the WebSocket security contract explicit and fail-closed enough for a future stable surface.

Problem:
`ServeWSWithConfig` currently verifies JWT only when a token is present, `ServeWSWithAuth` defaults to allow-all origins, and the admin broadcast endpoint reads an unbounded body guarded by the same raw secret used elsewhere. Package docs also imply stronger authentication than the implementation enforces.

Scope:
- Add explicit configuration for whether a connection requires JWT authentication.
- Preserve intentional unauthenticated development behavior only behind an obvious opt-in.
- Replace implicit allow-all origin behavior with explicit configuration or clearly named unsafe helper behavior.
- Add a bounded request body limit and validation path for the admin broadcast endpoint.
- Keep room password behavior explicit and compatible where possible.
- Update tests for missing token, invalid token, allow-all origin, denied origin, and oversized broadcast bodies.

Non-goals:
- Do not add non-standard-library dependencies.
- Do not introduce tenant policy, business channel authorization, or role-based access control.
- Do not promote `x/websocket` to beta or stable in this card.
- Do not change unrelated Hub broadcast semantics.

Files:
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/auth.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required. Document the final authentication, origin, token transport, and admin broadcast body-limit semantics.

Done Definition:
- Missing JWT behavior is controlled by an explicit option and covered by tests.
- Origin allow-all behavior is impossible to enable accidentally or is documented as an unsafe convenience.
- Admin broadcast rejects oversized bodies before reading them fully.
- Docs describe implemented behavior without overstating JWT enforcement.

Outcome:
- Added explicit JWT-required defaults with `AllowUnauthenticated` as the opt-in room-password-only mode.
- Made origin allow-all explicit through `AllowAllOrigins`; `ServeWSWithAuth` remains the compatibility helper with that opt-in set.
- Added `BroadcastMaxBytes` and oversized-body rejection for the admin broadcast endpoint.
- Updated WebSocket module docs and tests for missing token, explicit origin allow, and oversized broadcast bodies.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go test -race -timeout 60s ./x/websocket/...`
  - `go vet ./x/websocket/...`
