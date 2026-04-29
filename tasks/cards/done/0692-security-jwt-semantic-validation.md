# Card 0692

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- docs/modules/security/README.md
Depends On: 0691

Goal:
Complete JWT semantic validation around caller context, JOSE header shape, payload key identity, and persisted signing key material.

Scope:
- Honor canceled contexts in token generation and verification before expensive work.
- Require the JOSE header `typ` to be `JWT`.
- Require payload `kid` to match the verified header `kid`.
- Validate persisted signing keys during manager startup before caching them.
- Add negative tests for canceled context, bad `typ`, mismatched payload `kid`, and malformed persisted signing keys.

Non-goals:
- Do not add token revocation or session lifecycle behavior.
- Do not change the JWT public payload fields.
- Do not introduce a third-party JWT library.

Files:
- `security/jwt/jwt.go`
- `security/jwt/jwt_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/jwt`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for JWT validation semantics.

Done Definition:
- Canceled contexts stop JWT generation and verification.
- JWTs with non-`JWT` `typ` or mismatched payload/header key IDs are rejected.
- Persisted signing keys with missing IDs, unsupported algorithms, or invalid key material are rejected at startup.
- Targeted security tests and vet pass.

Outcome:
- Added canceled-context checks to JWT token generation and verification.
- Required JOSE header `typ` to equal `JWT` and header/payload key IDs to match.
- Added startup validation for persisted signing key IDs, algorithms, and key material lengths.
- Added negative tests for canceled contexts, invalid header type, mismatched payload key ID, and malformed persisted signing keys.
- Synced the security module primer with JWT semantic validation guidance.
- Validation run: `go test -race -timeout 60s ./security/jwt`; `go test -timeout 20s ./security/...`; `go vet ./security/...`.
