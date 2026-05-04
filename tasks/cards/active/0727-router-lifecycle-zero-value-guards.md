# Card 0727

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: router
Owned Files: router/router.go, router/registration.go, router/dispatch.go, router/metadata.go, router/router_contract_test.go
Depends On: 0726-router-any-method-contract

Goal:
Make nil and zero-value `Router` behavior explicit and non-panicking.

Scope:
- Guard public router methods that currently dereference `state`.
- Return registration errors for nil or zero-value routers.
- Serve a stable unavailable response instead of panicking for nil or
  zero-value `ServeHTTP`.
- Add focused tests for nil/zero-value lifecycle behavior.

Non-goals:
- Making zero-value Router fully functional.
- Changing `NewRouter` initialization.
- Adding new public lifecycle APIs.

Files:
- router/router.go
- router/registration.go
- router/dispatch.go
- router/metadata.go
- router/router_contract_test.go

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required unless behavior language changes.

Done Definition:
- Nil and zero-value router calls do not panic on public methods.
- Registration on nil/zero-value routers returns errors.
- Router tests, race tests, and vet pass.

Outcome:

