# Card 1551

Milestone: M-015
Recipe: specs/change-recipes/add-package.yaml
Priority: P2
State: done
Primary Module: x/data
Owned Files:
- `x/data/sqlx/adapter.go`
- `x/data/sqlx/adapter_test.go`
- `x/data/sqlx/module.yaml`
- `x/data/sqlx/go.mod`

Goal:
- Create x/data/sqlx/ as a separately versioned sub-package adapting
  github.com/jmoiron/sqlx to the store/db Querier and Transactor interfaces,
  tested offline using go-sqlmock without a live database.

Scope:
- Create x/data/sqlx/go.mod with module github.com/spcent/plumego/x/data/sqlx
  and dependencies github.com/jmoiron/sqlx and github.com/DATA-DOG/go-sqlmock.
- Create x/data/sqlx/adapter.go defining:
  - DB struct wrapping *sqlx.DB.
  - New(driverName, dataSourceName string) (*DB, error).
  - QueryRow, Query, Exec methods satisfying store/db Querier.
  - BeginTx(ctx, opts) (store/db.Tx, error) satisfying store/db Transactor.
  - Tx struct wrapping *sqlx.Tx with Commit, Rollback, and Querier methods.
  - NamedExec and NamedQuery helpers for struct-tag binding.
  - Close() error.
- Create x/data/sqlx/module.yaml with status = experimental.
- Write x/data/sqlx/adapter_test.go using go-sqlmock covering:
  - QueryRow with expected row returns scanned result.
  - Query with expected rows returns slice.
  - Exec with expected result returns rows affected.
  - BeginTx + Commit round-trip.
  - BeginTx + Rollback round-trip.
  - Scan error on mismatched column type (negative path).

Non-goals:
- Do not add sqlx to the main module go.mod.
- Do not implement a query builder or ORM.
- Do not require a live database for any test.
- Do not modify store/db interfaces.

Files:
- `x/data/sqlx/adapter.go`
- `x/data/sqlx/adapter_test.go`
- `x/data/sqlx/module.yaml`
- `x/data/sqlx/go.mod`

Tests:
- `go test -race -timeout 60s ./x/data/sqlx/...`
- `go vet ./x/data/sqlx/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- none at this card.

Done Definition:
- x/data/sqlx/go.mod is a separate module.
- DB struct satisfies store/db Querier and Transactor.
- All six adapter test cases pass with `go test -race`.
- No live database required in CI.

Outcome:
- Added `x/data/sqlx` as a separate Go module with sqlx and go-sqlmock,
  keeping sqlx out of the main module.
- Implemented `DB`, `New`, `NewWithDB`, `QueryRow`, `Query`, `Exec`,
  `BeginTx`, `NamedExec`, `NamedQuery`, `Close`, and transaction wrappers with
  explicit context propagation.
- Current `store/db` exposes the stable `DB` database/sql-shaped contract, not
  the `Querier`, `Transactor`, or `Tx` interfaces named by this card. To avoid
  a stable-root API change, this adapter exposes context-aware `Querier`,
  `Transactor`, and `Tx` contracts inside `x/data/sqlx` and uses `store/db`
  error sentinels for query, transaction, and connection failures.
- Added go-sqlmock coverage for QueryRow, Query, Exec row count, commit,
  rollback, and a scan error negative path.
- Synced `x/data/sqlx/module.yaml`, `x/data/module.yaml`, and dependency rules.
- Validation passed with `x/data/sqlx` race tests, `x/data/sqlx` vet,
  dependency-rules, module-manifests, agent-workflow, `gofmt -l .`, and
  `git diff --check`.
