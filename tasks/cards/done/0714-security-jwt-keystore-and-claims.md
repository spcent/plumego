# Card 0714

Milestone:
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P0
State: done
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- specs/dependency-rules.yaml
- docs/modules/security/README.md
- docs/stable-api/snapshots/security-head.snapshot
Depends On:

Goal:
Remove the concrete `store/kv` dependency from stable JWT primitives and make token verification fail closed on missing temporal claims without adding session or revocation behavior.

Scope:
- Replace the concrete `*store/kv.KVStore` dependency in `security/jwt` with a narrow JWT key store interface.
- Keep existing source-level callers that pass `*store/kv.KVStore` working through interface satisfaction.
- Keep token generation responsible for key rotation; keep verification as a read-only validation path.
- Reject verified JWT payloads that omit required temporal claims.
- Update focused JWT tests, security docs, and the stable API snapshot.

Non-goals:
- Do not add revocation, token version invalidation, or tenant session behavior.
- Do not introduce a third-party JWT library or a new dependency.
- Do not move durable key providers into stable `security`.

Files:
- `security/jwt/jwt.go`
- `security/jwt/jwt_test.go`
- `specs/dependency-rules.yaml`
- `docs/modules/security/README.md`
- `docs/stable-api/snapshots/security-head.snapshot`

Tests:
- `go test -race -timeout 60s ./security/jwt`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for JWT key-store and verification semantics.

Done Definition:
- `security/jwt` no longer imports `store/kv`.
- `NewJWTManager` accepts a minimal key-store interface.
- Token verification does not rotate or persist signing keys.
- JWTs without issued-at, not-before, or expiration claims are rejected.
- Targeted security tests, vet, and the security API snapshot are updated.

Outcome:
- Replaced the concrete `*store/kv.KVStore` dependency in `security/jwt` with the exported minimal `KeyStore` interface.
- Removed the dependency-rules exception that allowed stable `security` to import `store/kv`.
- Kept existing callers source-compatible because `*store/kv.KVStore` satisfies the new interface.
- Made JWT verification read-only with respect to key rotation; automatic rotation remains on token issuance and explicit rotation.
- Rejected JWTs that omit `iat`, `nbf`, or `exp` temporal claims.
- Synced the security module primer and stable security API snapshot.
- Validation run: `go test -race -timeout 60s ./security/jwt`; `go test -timeout 20s ./security/...`; `go vet ./security/...`; `go run ./internal/checks/dependency-rules`.
