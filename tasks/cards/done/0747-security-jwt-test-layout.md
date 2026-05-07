# Card 0747

Milestone:
Recipe: specs/change-recipes/refactor.yaml
Priority: P2
State: done
Primary Module: security
Owned Files:
- security/jwt/jwt_test.go
- security/jwt/jwt_keys_test.go
- security/jwt/jwt_verify_test.go
- security/jwt/jwt_claims_test.go
- security/jwt/jwt_rotation_test.go

Goal:
Split oversized JWT tests by behavior area without changing assertions.

Scope:
- Move key lifecycle, verification negative, claims/context, and rotation tests into focused files.
- Keep package, helper names, and assertions unchanged.
- Preserve existing negative matrix tests.

Non-goals:
- Do not change production code.
- Do not rewrite assertions opportunistically.
- Do not delete coverage.

Files:
- security/jwt/jwt_test.go
- security/jwt/jwt_keys_test.go
- security/jwt/jwt_verify_test.go
- security/jwt/jwt_claims_test.go
- security/jwt/jwt_rotation_test.go

Tests:
- go test -timeout 20s ./security/jwt
- go vet ./security/jwt
- go run ./internal/checks/dependency-rules

Docs Sync:
- Not required for behavior-neutral test refactor.

Done Definition:
- `jwt_test.go` is materially smaller.
- Test coverage remains in `go test ./security/jwt`.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Split key lifecycle/config tests into `jwt_keys_test.go`.
- Split rotation/concurrency tests into `jwt_rotation_test.go`.
- Split verification negative matrix tests into `jwt_verify_test.go`.
- Split claims/context/auth adapter tests into `jwt_claims_test.go`.
- Reduced `jwt_test.go` to shared helpers, basic generation tests, and benchmarks.

Validation:
- `go test -timeout 20s ./security/jwt`
- `go vet ./security/jwt`
- `go run ./internal/checks/dependency-rules`
