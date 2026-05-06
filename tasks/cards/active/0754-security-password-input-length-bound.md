# Card 0754

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: security
Owned Files:
- security/password/password.go
- security/password/password_test.go
- docs/modules/security/README.md

Goal:
Bound password hashing and verification input size to avoid attacker-controlled CPU amplification.

Scope:
- Add a stable sentinel error for passwords that exceed the hashing/checking input bound.
- Enforce the bound in `HashPasswordWithCost` and `CheckPassword`.
- Add tests for hash and check paths.
- Sync security docs.

Non-goals:
- Do not change hash format.
- Do not change `PasswordStrengthConfig`.
- Do not add external password hashing dependencies.

Files:
- security/password/password.go
- security/password/password_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/password
- go vet ./security/password
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required for new sentinel and security behavior.

Done Definition:
- Overlong plaintext passwords fail before PBKDF2 work in hash and check paths.
- Existing valid hashes continue to verify.
- Targeted tests, vet, and dependency checks pass.
