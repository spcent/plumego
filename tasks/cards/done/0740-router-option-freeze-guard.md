# Card 0740

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/router.go, router/router_contract_test.go, docs/modules/router/README.md, router/module.yaml, tasks/cards/active/README.md
Depends On: 0739-router-dependency-rule-sync

Goal:
Prevent exported router options from mutating runtime policy after `Freeze()`.

Scope:
- Make `WithMethodNotAllowed` reuse the same freeze-aware mutation path as
  `SetMethodNotAllowed`.
- Add regression coverage that applying the option directly after `Freeze()`
  does not change policy.
- Keep the existing option behavior during `NewRouter(...)`.
- Sync router docs and manifest lifecycle notes if needed.

Non-goals:
- Adding new router options.
- Changing default 405 behavior.
- Changing route registration freeze behavior.

Files:
- router/router.go
- router/router_contract_test.go
- docs/modules/router/README.md
- router/module.yaml
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required if lifecycle wording changes.

Done Definition:
- `WithMethodNotAllowed` cannot bypass frozen router policy.
- Regression tests cover option-after-freeze behavior.
- Router targeted tests, race tests, and vet pass.

Outcome:
- `WithMethodNotAllowed` now calls `SetMethodNotAllowed`, so direct option
  application and constructor usage share the same freeze-aware policy path.
- Added regression coverage that applying `WithMethodNotAllowed(true)` after
  `Freeze()` does not mutate method-not-allowed policy.
- Documented direct option application after `Freeze()` as a no-op and recorded
  exported option bypass as a router change risk.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
