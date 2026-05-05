# Card 0752

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: done
Primary Module: router
Owned Files: router/metadata.go, router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0751-router-print-nil-writer-guard

Goal:
Make public introspection snapshot APIs return empty collections consistently.

Scope:
- Return an empty `[]RouteInfo` from `Routes` for nil or zero-value routers.
- Add regression coverage that `Routes` and `NamedRoutes` return non-nil empty
  collections for uninitialized routers.
- Sync docs for empty snapshot behavior.

Non-goals:
- Changing route sorting.
- Changing route metadata shape.
- Adding new introspection APIs.

Files:
- router/metadata.go
- router/router_contract_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Public introspection snapshots do not require nil-slice special casing for
  uninitialized routers.
- Router targeted tests, race tests, and vet pass.

Outcome:
- Changed `Routes` to return a non-nil empty snapshot for nil and zero-value
  routers.
- Added regression coverage for non-nil empty `Routes` and `NamedRoutes`
  snapshots.
- Documented the uninitialized-router snapshot behavior.

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
