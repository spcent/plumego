# Card 0739

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: specs/dependency-rules.yaml, router/module.yaml, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0738-router-internal-structure-cleanup

Goal:
Synchronize router boundary specs so machine-readable rules match the module
manifest and current implementation.

Scope:
- Remove stale router allowance for `middleware` from dependency rules if still
  unused.
- Keep module manifest and router docs aligned with the final stable surface.
- Mark the active queue empty when this batch is done.
- Run boundary/manifest checks.

Non-goals:
- Changing router imports.
- Widening router into middleware ownership.
- Repo-wide release tagging.

Files:
- specs/dependency-rules.yaml
- router/module.yaml
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go run ./internal/checks/dependency-rules
- go run ./internal/checks/module-manifests
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Router dependency rules and manifest agree on allowed stable imports.
- Active queue is empty.
- Boundary checks and router vet pass.

Outcome:
- Removed the stale router allowance for `middleware` from dependency rules.
- Documented that router stable imports are limited to stdlib plus `contract`
  and middleware integration stays outside router.
- Marked the active queue empty.

Validation:
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`
- `go vet ./router/...`
