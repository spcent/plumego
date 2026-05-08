# Card 1270

Milestone: M-003
Recipe: specs/change-recipes/security-hardening.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/auth.go`
- `x/websocket/security.go`
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
Depends On: 0763

Goal:
- Split token authentication, room authorization, and anonymous access into explicit policies.

Problem:
The current `RoomAuthenticator` combines room password checks and JWT verification. `New` also requires a JWT secret even when unauthenticated connections are explicitly allowed, which makes anonymous and room-password-only use cases unclear.

Scope:
- Introduce separate token authentication and room authorization contracts.
- Allow explicitly unauthenticated mode without requiring a JWT secret.
- Rename or reshape the built-in HS256 verifier so it is clearly a simple helper, not a full OIDC/JWT policy engine.
- Preserve fail-closed defaults for authenticated configurations.
- Update tests for JWT-required, anonymous, room-authorized, and custom-auth flows.

Non-goals:
- Do not add OIDC, JWKS, or third-party JWT dependencies.
- Do not weaken origin checks.
- Do not keep deprecated compatibility wrappers.

Files:
- `x/websocket/auth.go`
- `x/websocket/security.go`
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for authentication modes, JWT caveats, and migration notes.

Done Definition:
- Anonymous mode can be configured without a JWT secret.
- Token authentication and room authorization are separate and testable.
- The built-in HS256 helper documents that it validates only signature and supported temporal claims.

Outcome:
- Replaced the combined `RoomAuthenticator` contract with separate
  `TokenAuthenticator` and `RoomAuthorizer` contracts.
- Added `SimpleHS256TokenAuth` for the built-in compact HS256 verifier and made
  `SimpleRoomAuth` room-password-only.
- Updated `ServerConfig` and `WebSocketConfig` to carry token and room policies
  separately.
- Allowed explicit unauthenticated `New` configurations without requiring a JWT
  secret.
- Updated tests and docs for split auth behavior.
