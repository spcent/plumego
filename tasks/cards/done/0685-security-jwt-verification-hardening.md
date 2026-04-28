# Card 0685

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- security/jwt/auth_jwt_test.go
- docs/modules/security/README.md
Depends On:

Goal:
Harden JWT verification so semantic checks fail closed and signing-key rotation cannot mutate manager state without the manager write lock.

Scope:
- Fix the unlocked `ensureRotationUnsafe` call from `VerifyToken`.
- Normalize default JWT algorithm handling so an omitted algorithm does not fail later in key generation.
- Require configured issuer and audience claims to be present and exact during verification.
- Require verified tokens to carry a subject.
- Remove misleading JWT docs that reference non-existent manager lifecycle APIs.
- Add negative tests for missing issuer, missing audience, missing subject, and verify-time rotation.

Non-goals:
- Do not add session revocation, token-version invalidation, tenant policy, or revocation stores to stable security.
- Do not change the public token payload shape.
- Do not introduce non-stdlib JWT dependencies.

Files:
- `security/jwt/jwt.go`
- `security/jwt/jwt_test.go`
- `security/jwt/auth_jwt_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/jwt`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required if JWT verification semantics or module guidance text changes.

Done Definition:
- `VerifyToken` does not call state-mutating rotation code without holding the manager write lock.
- Tokens missing configured issuer, configured audience, or subject fail closed.
- Empty `JWTConfig.Algorithm` is normalized consistently before key generation.
- Targeted security tests and vet pass.

Outcome:
- Normalized empty JWT algorithm config to HS256 before validation and key generation.
- Wrapped verify-time rotation in the JWT manager write lock.
- Required configured issuer, configured audience, and subject during token verification.
- Added regression coverage for missing semantic claims and concurrent verify-time rotation.
- Synced the security module primer with the fail-closed JWT claim contract.
- Validation run: `go test -race -timeout 60s ./security/jwt`; `go test -timeout 20s ./security/...`; `go vet ./security/...`.
