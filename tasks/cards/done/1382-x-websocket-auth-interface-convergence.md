# Card 1382

Milestone: M-003
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/auth.go`
- `x/websocket/security.go`
- `x/websocket/server.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On:
- M-003

Goal:
- Converge auth interfaces before stable by removing JWT/password-specific method names from the server-facing contracts.

Scope:
- Rename the server-facing token contract from `VerifyJWT` to a token-neutral method.
- Make room authorization use one request-aware interface.
- Replace `RoomAuthorization.QueryTokenOK` with explicit token-source data.
- Keep simple HS256 and room-password helpers as lightweight built-ins with clear comments.
- Update all in-repo callers and tests in the same commit.

Non-goals:
- Do not implement full OIDC, issuer, audience, or required-claims policy.
- Do not change the password hashing package.

Files:
- `x/websocket/auth.go`
- `x/websocket/security.go`
- `x/websocket/server.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for public interface rename and token-source semantics.

Done Definition:
- Old server-facing auth method names are removed from production code.
- Custom room policy no longer needs to implement a dummy password method.

Outcome:
- Server-facing auth uses `TokenAuthenticator` and `RoomAuthorizer`.
- Built-in HS256 token and room-password helpers remain explicit lightweight
  helpers.
- README and manifest list the converged auth entrypoints.

Validation:
- go test -timeout 20s ./x/websocket/...
- go vet ./x/websocket/...
- go build ./...
