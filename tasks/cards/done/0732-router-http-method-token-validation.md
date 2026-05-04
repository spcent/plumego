# Card 0732

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: done
Primary Module: router
Owned Files: router/registration.go, router/router_contract_test.go, docs/modules/router/README.md
Depends On: 0731-router-release-readiness-coverage

Goal:
Reject invalid HTTP method names using token-level validation while preserving
custom extension methods.

Scope:
- Replace whitespace-only method validation with HTTP token character checks.
- Keep `router.MethodAny` accepted as the reserved fallback sentinel.
- Add negative tests for control characters, separators, and non-token method
  names.
- Update router docs for token method validation.

Non-goals:
- Restricting methods to only standard `net/http` constants.
- Changing `core.App` route helper names.
- Changing MethodAny fallback behavior.

Files:
- router/registration.go
- router/router_contract_test.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Invalid HTTP token methods fail registration.
- Standard methods, custom token methods, and MethodAny still register.
- Router tests, race tests, and vet pass.

Outcome:
- Replaced whitespace-only method validation with HTTP token character
  validation.
- Kept custom token methods and `router.MethodAny` accepted.
- Added negative coverage for separators, control characters, and non-ASCII
  method names.
- Updated router docs for token method validation.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
