# Card 0690

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/authn/authn.go
- security/authn/authn_test.go
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- security/password/password.go
- security/password/password_test.go
- security/password/password_strength_test.go
- docs/modules/security/README.md
Depends On: 0689

Goal:
Close remaining mutable ownership and resource-boundary gaps in auth principals, JWT signing keys, and password hash verification.

Scope:
- Defensively copy principal roles, scopes, and claims when storing to and reading from context.
- Defensively copy JWT signing key material when caching or returning signing keys from rotation.
- Add password hash cost upper bound to avoid attacker-controlled verification work.
- Validate stored password salt and derived-hash lengths before PBKDF2 verification.
- Add ownership and negative-path tests for principal context, rotated signing keys, and password hash bounds.

Non-goals:
- Do not add session lifecycle, revocation, tenant policy, or application authorization to stable security.
- Do not change the password hash wire format.
- Do not remove public fields from stable structs.

Files:
- `security/authn/authn.go`
- `security/authn/authn_test.go`
- `security/jwt/jwt.go`
- `security/jwt/jwt_test.go`
- `security/password/password.go`
- `security/password/password_test.go`
- `security/password/password_strength_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/authn ./security/jwt ./security/password`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for copy-ownership and password bounds semantics.

Done Definition:
- Principal context helpers do not expose caller-owned mutable slices or maps.
- Mutating a returned rotated JWT signing key cannot mutate manager-held key material.
- Password hashes with oversized cost, invalid salt length, or invalid hash length fail before expensive verification.
- Targeted security tests and vet pass.

Outcome:
- Added defensive copy behavior for `authn.WithPrincipal` and `authn.PrincipalFromContext`.
- Added defensive signing-key copies when JWT managers cache or return rotated keys.
- Added `password.MaximumCost` and cost validation for hashing and stored-hash verification.
- Validated stored password salt and hash lengths before running PBKDF2 verification.
- Synced the security module primer with principal copy ownership and password bounds guidance.
- Validation run: `go test -race -timeout 60s ./security/authn ./security/jwt ./security/password`; `go test -timeout 20s ./security/...`; `go vet ./security/...`.
