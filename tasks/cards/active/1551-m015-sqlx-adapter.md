# Card 1551

Milestone: M-015
Recipe: specs/change-recipes/add-package.yaml
Priority: P2
State: active
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
-
