# Card 0744

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: security
Owned Files:
- security/abuse/limiter.go
- security/abuse/limiter_test.go
- docs/modules/security/README.md

Goal:
Give abuse limiter callers an explicit fail-closed path for empty limiter keys and remove panic-style examples.

Scope:
- Add an error-returning limiter check for callers that require a non-empty key.
- Keep `Allow` compatibility behavior that maps empty keys to `unknown`.
- Add tests for the strict empty-key path.
- Replace panic-style package examples with return-error style guidance.

Non-goals:
- Do not change token bucket math.
- Do not remove `Allow`.
- Do not change middleware wiring.

Files:
- security/abuse/limiter.go
- security/abuse/limiter_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/abuse
- go vet ./security/abuse
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required for new strict-key behavior and examples.

Done Definition:
- Callers can reject empty limiter keys explicitly.
- Existing `Allow` behavior remains compatible.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Added `ErrInvalidKey` and `Limiter.AllowKey` for fail-closed non-empty key checks.
- Preserved `Allow` compatibility behavior that maps empty keys to `unknown`.
- Replaced panic-style package examples with return-error style guidance.
- Added strict empty-key tests and docs.

Validation:
- `gofmt -w security/abuse/limiter.go security/abuse/limiter_test.go`
- `go test -timeout 20s ./security/abuse`
- `go vet ./security/abuse`
- `go run ./internal/checks/dependency-rules`
