# Card 1550

Milestone: M-015
Recipe: specs/change-recipes/add-package.yaml
Priority: P2
State: active
Primary Module: x/data
Owned Files:
- `x/data/pgx/adapter.go`
- `x/data/pgx/adapter_test.go`
- `x/data/pgx/module.yaml`
- `x/data/pgx/go.mod`

Goal:
- Create x/data/pgx/ as a separately versioned sub-package adapting pgx v5 to
  the store/db Querier and Transactor interfaces, tested offline without a live
  PostgreSQL instance.

Scope:
- Create x/data/pgx/go.mod with module github.com/spcent/plumego/x/data/pgx
  and dependency github.com/jackc/pgx/v5.
- Create x/data/pgx/adapter.go defining:
  - DB struct wrapping *pgxpool.Pool.
  - New(ctx context.Context, connStr string) (*DB, error) — creates pool.
  - QueryRow, Query, Exec methods satisfying store/db Querier.
  - BeginTx(ctx, opts) (store/db.Tx, error) satisfying store/db Transactor.
  - Tx struct wrapping pgx.Tx with Commit, Rollback, and Querier methods.
  - Close() error for pool shutdown.
- Create x/data/pgx/module.yaml with status = experimental.
- Write x/data/pgx/adapter_test.go using pgxmock v2 or a mock pgxpool covering:
  - QueryRow returns scanned value.
  - Query returns multiple rows.
  - Exec affects row count.
  - BeginTx followed by Commit succeeds.
  - BeginTx followed by Rollback succeeds.
  - Connection failure returns error (negative path).

Non-goals:
- Do not add pgx to the main module go.mod.
- Do not implement an ORM or query builder.
- Do not require a live PostgreSQL instance for any test.
- Do not modify store/db interfaces.

Files:
- `x/data/pgx/adapter.go`
- `x/data/pgx/adapter_test.go`
- `x/data/pgx/module.yaml`
- `x/data/pgx/go.mod`

Tests:
- `go test -race -timeout 60s ./x/data/pgx/...`
- `go vet ./x/data/pgx/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- none at this card; operational guidance added when x/data reaches beta.

Done Definition:
- x/data/pgx/go.mod is a separate module.
- DB struct satisfies store/db Querier and Transactor.
- All six adapter test cases pass with `go test -race`.
- No live database required in CI.

Outcome:
-
