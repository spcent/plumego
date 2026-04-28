# Card 0688

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/authn/token.go
- security/authn/token_test.go
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- security/jwt/auth_jwt.go
- security/jwt/auth_jwt_test.go
- docs/modules/security/README.md
Depends On: 0687

Goal:
Harden authentication primitives around exact bearer parsing, fixed-token comparison, JWT active-key recovery, and principal/claims ownership.

Scope:
- Require an actual whitespace delimiter after the `Bearer` scheme so values like `BearerX token` are rejected.
- Compare static tokens through fixed-length digest comparison to avoid length-dependent secret comparisons.
- Recover from a stale persisted JWT active-key pointer by creating and persisting a new active key.
- Defensively copy JWT roles, permissions, and claims when converting to principals or storing/retrieving token claims from context.
- Add negative and ownership tests for token parsing, static-token fallback, JWT active-key recovery, and claims aliasing.

Non-goals:
- Do not add session lifecycle or revocation behavior to stable security.
- Do not change the public authn or JWT type names.
- Do not introduce non-stdlib dependencies.

Files:
- `security/authn/token.go`
- `security/authn/token_test.go`
- `security/jwt/jwt.go`
- `security/jwt/jwt_test.go`
- `security/jwt/auth_jwt.go`
- `security/jwt/auth_jwt_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/authn ./security/jwt`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for bearer parsing and defensive-copy ownership semantics.

Done Definition:
- Malformed bearer schemes fail closed and do not authenticate via Authorization header parsing.
- Static token comparison does not compare raw secret lengths directly.
- A manager with a missing persisted active key can issue new tokens after startup.
- Principal and token-claims context accessors do not expose caller-owned mutable slices.
- Targeted security tests and vet pass.

Outcome:
- Tightened bearer token extraction to require the exact `Bearer` scheme followed by a whitespace delimiter and a single token value.
- Changed `StaticToken` credential comparison to compare fixed-length SHA-256 digests with constant-time comparison.
- Added JWT startup recovery for persisted active-key IDs that no longer have a stored signing key.
- Defensively copied JWT role and permission slices in `PrincipalFromClaims`, `WithTokenClaims`, and `TokenClaimsFromContext`.
- Synced the security module primer with bearer parsing, static-token comparison, and JWT copy-ownership guidance.
- Validation run: `go test -race -timeout 60s ./security/authn ./security/jwt`; `go test -timeout 20s ./security/...`; `go vet ./security/...`.
