# Card 0716

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/password/password.go
- security/password/password_strength_test.go
- docs/modules/security/README.md
Depends On: 0715

Goal:
Align password cost validation with the exported `MinimumCost` security contract.

Scope:
- Reject password hash generation costs below `MinimumCost`.
- Reject stored password hashes whose encoded cost is below `MinimumCost`.
- Keep `ErrInvalidCost` and `ErrInvalidHash` classification with `errors.Is`.
- Update password tests and security docs.

Non-goals:
- Do not add password rehash/migration helpers.
- Do not change PBKDF2 algorithm or stored hash format.
- Do not log or return password material.

Files:
- `security/password/password.go`
- `security/password/password_strength_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/password`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for minimum password cost behavior.

Done Definition:
- `HashPasswordWithCost` rejects costs below `MinimumCost`.
- `CheckPassword` rejects stored hashes with costs below `MinimumCost` as invalid hashes.
- Password tests and security docs reflect the enforced minimum.

Outcome:
- Enforced `MinimumCost` in shared password cost validation.
- Kept generated hashes and stored hash verification bounded by `MinimumCost` and `MaximumCost`.
- Wrapped stored-hash low-cost failures so callers can classify them with both `ErrInvalidHash` and `ErrInvalidCost` through `errors.Is`.
- Added focused tests for below-minimum generation and stored hash verification.
- Synced the security module primer with the enforced cost bounds.
- Validation run: `go test -race -timeout 60s ./security/password`; `go test -timeout 20s ./security/...`; `go vet ./security/...`.
