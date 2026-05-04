# Card 0719

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- docs/modules/security/README.md
- docs/stable-api/snapshots/security-head.snapshot
Depends On: 0718

Goal:
Prevent JWT key-management APIs from returning private signing material and tighten JWT duration config validation.

Scope:
- Replace public key-rotation return data with signing-key metadata that omits HMAC secrets and Ed25519 private keys.
- Keep private signing material internal to JWT signing and persisted key storage.
- Reject negative `RotationInterval` and negative `ClockSkew` in `JWTConfig.Validate`.
- Update focused JWT tests, security docs, and the security API snapshot.

Non-goals:
- Do not change token wire format.
- Do not add token revocation or session lifecycle behavior.
- Do not add a new key provider or dependency.

Files:
- `security/jwt/jwt.go`
- `security/jwt/jwt_test.go`
- `docs/modules/security/README.md`
- `docs/stable-api/snapshots/security-head.snapshot`

Tests:
- `go test -race -timeout 60s ./security/jwt`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for JWT key-rotation return semantics and config validation.

Done Definition:
- `RotateKey` no longer exposes secret/private key bytes.
- Negative rotation interval and clock skew are rejected.
- Existing in-repo callers and tests are updated.
- Targeted security tests, vet, and the security API snapshot are updated.

Outcome:
- Pending.
