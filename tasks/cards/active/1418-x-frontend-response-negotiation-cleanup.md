# Card 1418

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/frontend
Owned Files:
- x/frontend/frontend.go
- x/frontend/negotiation.go
- x/frontend/response.go
- x/frontend/response_test.go
- x/frontend/security_test.go
Depends On:
- 1417

Goal:
- Isolate frontend response and content negotiation behavior before future maturity evaluation.

Scope:
- Move response interception and status replacement helpers into `response.go`.
- Keep content negotiation in `negotiation.go`.
- Preserve directory safety, range/precompressed behavior, and legacy extensionless request behavior.

Non-goals:
- Do not change default mount behavior.
- Do not alter embedded asset APIs.
- Do not promote `x/frontend` maturity in this card.

Files:
- x/frontend/frontend.go
- x/frontend/negotiation.go
- x/frontend/response.go
- x/frontend/response_test.go
- x/frontend/security_test.go

Tests:
- go test -timeout 20s ./x/frontend
- go vet ./x/frontend
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless response behavior text changes.

Done Definition:
- Response interception and negotiation responsibilities are separated.
- Existing frontend response and security tests pass.
- No public API or serving behavior changes.

Outcome:

