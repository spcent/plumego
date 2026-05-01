# Card 0739

Milestone:
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/server.go`
- `x/websocket/auth.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`
Depends On:

Goal:
- Finalize the client-facing WebSocket authentication contract as a fail-closed stable API.

Problem:
The current surface still has compatibility behavior that is unsafe or unclear for stable use: `ServeWSWithAuth` hides allow-all origin behavior, JWT verification is intentionally lightweight but not named as such, token lookup still accepts `?token=`, and `NewSimpleRoomAuth` accepts weak secrets when called directly.

Scope:
- Delete the misleading compatibility helper `ServeWSWithAuth`; expose only explicit config-based serving paths.
- Add explicit token transport policy; query-string token transport must be disabled by default and allowed only through an obvious opt-in such as `AllowQueryToken`.
- Rename or document the built-in JWT verifier as a simple HS256 helper, or add a validation config that covers `nbf`, `iat`, issuer, audience, and required claims.
- Replace `NewSimpleRoomAuth` with an error-returning constructor that validates secret strength, or move weak behavior behind an explicitly dev-only helper.
- Update all internal tests and examples that rely on deleted compatibility symbols.
- Follow `AGENTS.md §7.1` for every exported symbol removal or rename.

Non-goals:
- Do not add OIDC, JWKS fetching, or external JWT dependencies.
- Do not introduce tenant or role policy into `x/websocket`.
- Do not change admin broadcast authorization in this card.

Files:
- `x/websocket/server.go`
- `x/websocket/auth.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for auth defaults, token transport, JWT scope, deleted helpers, and migration notes.

Done Definition:
- Unsafe allow-all helper paths are deleted, not deprecated.
- Query token auth is fail-closed by default and tested.
- JWT semantics are either fully validated by config or explicitly named as simple HS256-only behavior.
- Room secret construction rejects weak production secrets.
- Old exported names are absent after `rg -n --glob '*.go' 'ServeWSWithAuth|NewSimpleRoomAuth' .`, unless a renamed dev helper is intentionally documented.

Outcome:
- Deleted `ServeWSWithAuth` and migrated callers to explicit `ServeWSWithConfig`.
- Added `AllowQueryToken`; query-string JWT transport is disabled by default and covered by tests.
- Changed `NewSimpleRoomAuth` to return an error and reject weak JWT secrets.
- Clarified the built-in verifier as a compact HS256 helper rather than a full OIDC/JWT policy.
- Updated docs and module manifest for the new client auth contract.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go build ./...`
