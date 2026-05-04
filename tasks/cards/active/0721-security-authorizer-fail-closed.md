# Card 0721

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: security
Owned Files:
- security/jwt/auth_jwt.go
- security/jwt/auth_jwt_test.go
- security/jwt/jwt_test.go
- docs/modules/security/README.md
- docs/stable-api/snapshots/security-head.snapshot
Depends On: 0720

Goal:
Make stable JWT authorizer helpers fail closed when no explicit policy or permission target is configured.

Scope:
- Make an empty `AuthZPolicy` deny by default unless an explicit allow-empty flag is set.
- Make `PermissionAuthorizer` deny empty action/resource checks unless explicitly allowed.
- Add compatibility flags for callers that intentionally want empty checks to pass.
- Update focused tests, docs, and the security API snapshot.

Non-goals:
- Do not add business-specific authorization policy.
- Do not add tenant session lifecycle or revocation logic.
- Do not change principal or token claim shapes beyond authorizer config fields.

Files:
- `security/jwt/auth_jwt.go`
- `security/jwt/auth_jwt_test.go`
- `security/jwt/jwt_test.go`
- `docs/modules/security/README.md`
- `docs/stable-api/snapshots/security-head.snapshot`

Tests:
- `go test -race -timeout 60s ./security/jwt`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for fail-closed authorizer defaults and explicit compatibility flags.

Done Definition:
- Empty policy and empty permission targets deny by default.
- Existing intentional allow-empty tests use explicit flags.
- Targeted security tests, vet, and the security API snapshot are updated.

Outcome:
- Pending.
