# Card 0735

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/registration.go, router/router_contract_test.go, router/router_conflict_test.go, docs/modules/router/README.md
Depends On: 0734-router-reverse-url-empty-param-contract

Goal:
Remove ambiguous last-wins route parameter behavior by rejecting duplicate
parameter names within one route pattern.

Scope:
- Reject repeated `:param` or `*param` names in a single route pattern.
- Update tests that currently assert last-wins behavior.
- Keep same-name params across different routes unchanged.
- Document the unique parameter-name contract.

Non-goals:
- Changing parameter name character validation.
- Changing wildcard terminal validation.
- Changing `Param` lookup API.

Files:
- router/registration.go
- router/router_contract_test.go
- router/router_conflict_test.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Duplicate param names within a route fail registration.
- Router tests reflect the unique-name contract.
- Router tests, race tests, and vet pass.

Outcome:
- Route registration now rejects duplicate parameter names within a fully
  composed route pattern.
- Updated former last-wins tests to assert registration failures for duplicate
  root and grouped params.
- Updated router docs to require unique route parameter names.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
