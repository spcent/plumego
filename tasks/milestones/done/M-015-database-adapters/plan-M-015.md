# Plan for M-015: Database Adapters

Milestone: `M-015`
Objective: Ship x/data/pgx, x/data/sqlx, and x/data/migrate as dependency-free
adapter contracts and migration helpers tested offline without a live database.
Constraints: no `go.mod` may exist under `x/**`, no pgx/sqlx/goose in main
module go.mod, store/db stable-root interfaces are read-only in this milestone,
offline tests only (no live PostgreSQL in CI), `plumego migrate` uses
caller-registered database/sql drivers.
Affected Modules: x/data, cmd/plumego.

## Phase Map

- Phase 1: Orient — read store/db interfaces to capture exact Querier and
  Transactor signatures before writing any adapter code.
- Phase 2: Implement (parallel) — write pgx adapter, sqlx adapter, and migrate
  runner concurrently since they are independent sub-packages.
- Phase 3: Test — confirm negative-path tests for all three adapters; confirm
  migration runner covers the four state transitions.
- Phase 4: Validate and Ship — run acceptance criteria, update reference/standard-service
  README, commit.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1550 | Create x/data/pgx/ dependency-free adapter contracts and offline tests | x/data | `x/data/pgx/adapter.go`, `x/data/pgx/adapter_test.go`, `x/data/pgx/module.yaml` | M-009 | `go test ./x/data/pgx/...`, `go vet ./x/data/pgx/...` |
| 1551 | Create x/data/sqlx/ database/sql-shaped helpers and offline tests | x/data | `x/data/sqlx/adapter.go`, `x/data/sqlx/adapter_test.go`, `x/data/sqlx/module.yaml` | M-009 | `go test ./x/data/sqlx/...`, `go vet ./x/data/sqlx/...` |
| 1552 | Create x/data/migrate/ runner contracts and add plumego migrate subcommands | x/data | `x/data/migrate/migrate.go`, `x/data/migrate/migrate_test.go`, `x/data/migrate/module.yaml`, `cmd/plumego/commands/migrate.go` | M-009 | `go test ./x/data/migrate/...`, `plumego migrate status` fails clearly without a registered driver |

## Dependency Edges

- Cards 1550, 1551, 1552 all depend on M-009 (store/db interfaces confirmed stable at beta).
- Cards 1550, 1551, 1552 are independent of each other.

## Parallel Groups

- Group A (parallel): cards 1550, 1551, 1552 — independent sub-packages, no file overlap
  except cmd/plumego/commands/migrate.go which belongs exclusively to 1552.

## Risk Register

- Risk: store/db Querier or Transactor interface signatures differ from what the adapter
  expects, causing type-assertion failures in tests.
  Mitigation: Phase 1 explicitly identifies exact interface signatures before writing
  any adapter code; card 1550 is blocked until signatures are confirmed.
- Risk: concrete driver expectations leak into the adapter contracts.
  Mitigation: use local mocks and standard-library shapes only; no x/* go.mod
  files or third-party driver imports are allowed.

## Verification Strategy

- Card-level checks: each adapter card runs its own `go test` immediately after writing
  the adapter; negative-path tests (connection failure, rollback, scan error) are
  verified in Phase 3.
- Migration runner check: confirm `plumego migrate status` fails clearly when no
  caller-registered SQL driver is available.
- Dependency audit: `go run ./internal/checks/dependency-rules` confirms pgx, sqlx,
  goose are absent from the main module go.mod.

## Exit Condition

- all three adapter cards completed
- no `go.mod` exists under `x/**`, and each adapter exposes explicit contracts
- offline negative-path tests pass for all three adapters
- `plumego migrate` subcommands use caller-registered SQL drivers
- reference/standard-service README mentions optional DB adapter path
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
