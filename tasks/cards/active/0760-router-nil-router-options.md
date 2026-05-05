# Card 0760

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: active
Primary Module: router
Owned Files: router/router.go, router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0759-router-reverse-url-strict-params

Goal:
Make nil `RouterOption` behavior consistent with nil `RouteOption` behavior.

Scope:
- Treat nil `RouterOption` values passed to `NewRouter` as no-ops.
- Add regression coverage that nil options do not panic and non-nil options
  still apply.
- Sync docs only if public option behavior is documented.

Non-goals:
- Changing the `RouterOption` type.
- Adding new options.
- Changing freeze semantics.

Files:
- router/router.go
- router/router_contract_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Optional.

Done Definition:
- `NewRouter(nil)` and mixed nil/non-nil options do not panic.
- Existing option behavior remains unchanged.
- Router targeted tests, race tests, and vet pass.

Outcome:
