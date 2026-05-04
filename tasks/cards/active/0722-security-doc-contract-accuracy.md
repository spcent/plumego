# Card 0722

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/input/input.go
- docs/modules/security/README.md
Depends On: 0721

Goal:
Align security package documentation and examples with the actual stable contract.

Scope:
- Remove unsupported API-token claims from JWT package docs unless backed by code.
- Adjust JWT examples to emphasize the `KeyStore` interface instead of coupling examples to a concrete store package.
- Tighten sanitizer comments so `SanitizeHTML` and `SanitizeSQL` read as lossy defense-in-depth helpers, not complete sanitizers.
- Update the security module README where needed.

Non-goals:
- Do not rename or remove exported sanitizer helpers.
- Do not introduce a third-party sanitizer.
- Do not change token or validation behavior.

Files:
- `security/jwt/jwt.go`
- `security/input/input.go`
- `docs/modules/security/README.md`

Tests:
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- This is a docs and comments sync card.

Done Definition:
- Public examples no longer imply a concrete store dependency for JWT.
- Sanitizer docs explicitly warn about lossy and incomplete behavior.
- Security tests, vet, and dependency boundary checks pass.

Outcome:
- Pending.
