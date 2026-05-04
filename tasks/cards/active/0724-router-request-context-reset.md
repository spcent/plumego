# Card 0724

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: active
Primary Module: router
Owned Files: router/dispatch.go, router/router_contract_test.go
Depends On: 0723-router-lifecycle-contract-docs

Goal:
Prevent stale route metadata from leaking through reused request contexts.

Scope:
- Ensure each matched request overwrites route params, route pattern, and route
  name consistently.
- Clear `RouteName` when the current route has no name.
- Add focused regression coverage for stale context metadata.

Non-goals:
- Changing `contract.RequestContext` shape.
- Changing route matching behavior.
- Changing response bodies.

Files:
- router/dispatch.go
- router/router_contract_test.go

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required.

Done Definition:
- Unnamed routes cannot inherit an upstream or previous `RouteName`.
- Router tests, race tests, and vet pass.

Outcome:

