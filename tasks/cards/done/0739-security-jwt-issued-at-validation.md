# Card 0739

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- docs/modules/security/README.md

Goal:
Make JWT time claim validation reject issued-at values that are in the future beyond configured clock skew.

Scope:
- Add an `iat > now + skew` negative check in `VerifyToken`.
- Add a focused verification test for a signed token with future `iat` and current `nbf`.
- Sync security docs with the full time-claim contract.

Non-goals:
- Do not change signing algorithms or key rotation behavior.
- Do not change public error sentinel names.
- Do not split JWT files in this card.

Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/jwt
- go vet ./security/jwt
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update security module docs for JWT time-claim checks.

Done Definition:
- Future `iat` claims fail verification.
- Existing time-claim tests still pass.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Added JWT verification rejection for issued-at values beyond `now + ClockSkew`.
- Added a signed-token regression case with future `iat` and current `nbf`.
- Synced security module docs with the full time-claim validation contract.

Validation:
- `gofmt -w security/jwt/jwt.go security/jwt/jwt_test.go`
- `go test -timeout 20s ./security/jwt`
- `go vet ./security/jwt`
- `go run ./internal/checks/dependency-rules`
