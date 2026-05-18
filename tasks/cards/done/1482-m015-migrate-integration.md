# Card 1552

Milestone: M-015
Recipe: specs/change-recipes/add-package.yaml
Priority: P2
State: done
Primary Module: x/data
Owned Files:
- `x/data/migrate/migrate.go`
- `x/data/migrate/migrate_test.go`
- `x/data/migrate/module.yaml`
- `x/data/migrate/go.mod`
- `cmd/plumego/commands/migrate.go`

Goal:
- Create x/data/migrate/ wrapping pressly/goose v3 for database migration
  execution, and extend the existing cmd/plumego/commands/migrate.go with
  `plumego migrate up`, `plumego migrate down`, and `plumego migrate status`
  subcommands.

Scope:
- Create x/data/migrate/go.mod with module github.com/spcent/plumego/x/data/migrate
  and dependency github.com/pressly/goose/v3.
- Create x/data/migrate/migrate.go defining:
  - Migrator struct wrapping goose.Provider.
  - New(db *sql.DB, migrationsDir string) (*Migrator, error).
  - Up(ctx context.Context) error — applies all pending migrations.
  - Down(ctx context.Context) error — rolls back one migration.
  - Status(ctx context.Context) ([]MigrationStatus, error) — returns pending
    and applied migration list.
  - MigrationStatus struct: Version int64; Name, State string; AppliedAt *time.Time.
- Create x/data/migrate/module.yaml with status = experimental.
- Write x/data/migrate/migrate_test.go using an in-memory SQLite database via
  database/sql + mattn/go-sqlite3 covering:
  - No migrations directory: New returns error.
  - Empty migrations: Status returns empty slice.
  - One pending migration: Up applies it; Status shows applied.
  - Applied migration: Down rolls it back; Status shows pending again.
- Extend cmd/plumego/commands/migrate.go to support up, down, and status
  subcommands; use plumego.migrate.yaml config file for DB DSN and migrations dir.

Non-goals:
- Do not add goose or sqlite3 to the main module go.mod.
- Do not implement migration file generation (goose has its own create command).
- Do not require a live database for migrate package tests.
- Do not modify store/db interfaces.

Files:
- `x/data/migrate/migrate.go`
- `x/data/migrate/migrate_test.go`
- `x/data/migrate/module.yaml`
- `x/data/migrate/go.mod`
- `cmd/plumego/commands/migrate.go`

Tests:
- `go test -race -timeout 60s ./x/data/migrate/...`
- `go vet ./x/data/migrate/...`
- `go test -race -timeout 60s ./cmd/plumego/...`

Docs Sync:
- Update cmd/plumego/README.md to document migrate up/down/status subcommands.

Done Definition:
- x/data/migrate/go.mod is a separate module.
- Up/Down/Status work correctly against an in-memory SQLite DB.
- `plumego migrate status` exits 0 on a project with no pending migrations.
- All migrate_test.go test cases pass with `go test -race`.

Outcome:
- Added `x/data/migrate` as a separate Go module wrapping goose v3.24.3 with
  an in-memory SQLite test driver, keeping goose and sqlite out of the main and
  CLI modules.
- Implemented `Migrator`, `New`, `NewWithDialect`, `NewWithFS`, `Up`, `Down`,
  `Status`, and `Close`. `Status` projects goose status into the card's
  `MigrationStatus` shape.
- Added offline tests for missing directory, empty migrations, applying one
  pending migration, and rolling it back.
- The CLI already had `migrate status/up/down` runtime commands without a goose
  dependency. This card extends it with `plumego.migrate.yaml` loading for
  `driver`, `db_url`, and `dir`, while preserving flag override behavior and
  keeping database drivers out of the bundled CLI module.
- Synced `x/data/migrate/module.yaml`, `x/data/module.yaml`, dependency rules,
  CLI help, CLI tests, and CLI README migration docs.
- Validation passed with `x/data/migrate` race tests, `x/data/migrate` vet,
  `cmd/plumego` migrate tests, full `cmd/plumego` race tests, `cmd/plumego`
  vet, dependency-rules, module-manifests, agent-workflow, `gofmt -l .`, and
  `git diff --check`.
