# Card 1507

Milestone: M-022
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/data
Owned Files:
- `x/data/module.yaml`
- `x/data/migrate/module.yaml`
- `x/data/pgx/module.yaml`
- `x/data/sqlx/module.yaml`
- `docs/modules/x/data/README.md`
Depends On: 1503

Goal:
- Replace conceptual `x/data` family entrypoints with real package or symbol
  inventory.
- Add missing `public_entrypoints` to the `migrate`, `pgx`, and `sqlx`
  manifests.

Scope:
- Update only the `x/data` family manifests and any directly corresponding
  README wording needed to explain the chosen entrypoint convention.

Non-goals:
- Do not rename `x/data/pgx` or `x/data/sqlx` in this card.
- Do not change adapter runtime behavior.
- Do not widen this card into distributed cache or sharding subpackages that
  already declare entrypoints.

Files:
- `x/data/module.yaml`
- `x/data/migrate/module.yaml`
- `x/data/pgx/module.yaml`
- `x/data/sqlx/module.yaml`
- `docs/modules/x/data/README.md`

Acceptance Tests:
<!-- none; manifest/docs card -->

Tests:
- `go test -timeout 20s ./x/data/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- `docs/modules/x/data/README.md`

Validation:
- `go test -timeout 20s ./x/data/...`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/public-entrypoints-sync`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Replaced conceptual root-family entrypoints with real `x/data` subordinate
  package names.
- Added concrete exported symbol inventories for the `migrate`, `pgx`, and
  `sqlx` manifests.
- Documented the root-vs-leaf manifest convention in the `x/data` module
  primer.
- Validation:
  - `go test -timeout 20s ./x/data/...`
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/public-entrypoints-sync`
