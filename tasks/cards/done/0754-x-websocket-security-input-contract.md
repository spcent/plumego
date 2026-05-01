# Card 0754

Milestone: M-003
Recipe: specs/change-recipes/security-hardening.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/auth.go`
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0739, 0740, 0741

Goal:
- Make the WebSocket security input contract fail-closed enough for a future stable API.

Problem:
Secrets are retained as caller-owned byte slices, `AllowedOrigins: ["*"]` bypasses the explicit `AllowAllOrigins` contract, malformed JWT temporal claims are ignored, and documentation examples still show secrets that are too short for the current policy.

Scope:
- Defensively copy JWT and broadcast secrets at construction boundaries.
- Remove `AllowedOrigins: ["*"]` as an implicit allow-all origin bypass; require `AllowAllOrigins`.
- Treat present but malformed JWT `exp` claims as invalid tokens.
- Keep query token transport disabled by default.
- Update examples and docs to use valid 32-byte-plus secrets.

Non-goals:
- Do not add OIDC/JWKS or non-stdlib JWT dependencies.
- Do not change admin broadcast authorization shape beyond copied secrets.
- Do not promote module maturity.

Files:
- `x/websocket/auth.go`
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for origin allow-all, secret requirements, and JWT verifier caveats.

Done Definition:
- Mutating caller-provided secret slices after construction cannot change verification behavior.
- `AllowedOrigins: ["*"]` does not allow browser origins unless `AllowAllOrigins` is also true.
- Malformed `exp` claims fail closed.
- Documentation examples match current secret policy.

Outcome:
- Defensively copied JWT and broadcast secrets at `NewSimpleRoomAuth` and `New` construction boundaries.
- Removed `AllowedOrigins: ["*"]` as an implicit allow-all origin bypass; callers must set `AllowAllOrigins`.
- Rejected malformed non-numeric JWT `exp` claims as `ErrInvalidToken`.
- Updated package/module examples to use secrets that satisfy the 32-byte policy.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go build ./...`
