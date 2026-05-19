# Card 1171

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: done
Primary Module: router
Owned Files: router/registration.go, router/static.go, router/static_test.go, tasks/cards/active/README.md
Depends On: 0754-router-static-prefix-cleanup

Goal:
Deduplicate static registration lifecycle preflight with `AddRoute` lifecycle
logic.

Scope:
- Extract a shared internal lifecycle check for route registration readiness and
  frozen state.
- Use it from `AddRoute` and static preflight so lifecycle wording stays in one
  place.
- Preserve static lifecycle error precedence before filesystem or input work.
- Keep tests from 0744 and 0745 passing.

Non-goals:
- Changing registration public APIs.
- Changing static request serving.
- Widening router responsibilities.

Files:
- router/registration.go
- router/static.go
- router/static_test.go
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required unless wording changes.

Done Definition:
- Lifecycle error string construction for route registration is shared.
- Static still preflights lifecycle before filesystem/input validation.
- Router targeted tests, race tests, and vet pass.

Outcome:
- Extracted shared route registration readiness and frozen-state error helpers.
- Reused the lifecycle preflight from static registration so `Static` and
  `StaticFS` keep lifecycle precedence before filesystem or file-system input
  validation.
- Kept `AddRoute` frozen checks under the write lock while sharing canonical
  lifecycle error construction.

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
