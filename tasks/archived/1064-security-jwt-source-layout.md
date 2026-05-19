# Card 1064

Milestone:
Recipe: specs/change-recipes/refactor.yaml
Priority: P2
State: done
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/jwt/config.go
- security/jwt/keys.go
- security/jwt/signing.go
- security/jwt/verify.go
- security/jwt/claims.go

Goal:
Split the JWT implementation by responsibility without changing behavior.

Scope:
- Move config/defaults, claim types, key lifecycle helpers, signing helpers, and verification helpers into focused files.
- Keep exported names, behavior, and package unchanged.
- Preserve gofmt and package-local tests.

Non-goals:
- Do not change JWT semantics in this card.
- Do not split tests in this card.
- Do not introduce new dependencies.

Files:
- security/jwt/jwt.go
- security/jwt/config.go
- security/jwt/keys.go
- security/jwt/signing.go
- security/jwt/verify.go
- security/jwt/claims.go

Tests:
- go test -timeout 20s ./security/jwt
- go vet ./security/jwt
- go run ./internal/checks/dependency-rules

Docs Sync:
- Not required for behavior-neutral refactor.

Done Definition:
- `jwt.go` no longer owns all JWT responsibilities.
- Public API and tests remain unchanged.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Split JWT config/defaults, claim/key interfaces, key lifecycle, signing, and verification into focused files.
- Left `jwt.go` focused on manager construction, time source, and token generation flow.
- Preserved exported API and package behavior.

Validation:
- `go test -timeout 20s ./security/jwt`
- `go vet ./security/jwt`
- `go run ./internal/checks/dependency-rules`
