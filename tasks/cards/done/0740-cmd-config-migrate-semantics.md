# Card 0740

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/config.go, cmd/plumego/internal/configmgr/configmgr.go, cmd/plumego/commands/migrate.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md
Depends On: 0739

Goal:
Make configuration diagnostics and migration command status/no-op semantics stable and predictable.

Scope:
- Return `.env` parse errors from `config env` instead of silently ignoring them.
- Reclassify migration no-op outcomes consistently as warning/degraded instead of generic errors.
- Decide and document whether `migrate status` creates the schema table or requires it to exist.
- Add focused tests for env parse failure and migration no-op/status behavior where feasible without external drivers.

Non-goals:
- Do not add database drivers.
- Do not implement migration locking.
- Do not change SQL migration file format.

Files:
- `cmd/plumego/commands/config.go`
- `cmd/plumego/internal/configmgr/configmgr.go`
- `cmd/plumego/commands/migrate.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./commands ./internal/configmgr ./internal/migrate`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- Invalid `.env` syntax fails `config env` with a structured error.
- Migration no-op responses use a stable warning/exit-code contract.
- `migrate status` side effects are either removed or explicitly documented and tested.

Outcome:
- Changed `config env` to propagate `.env` parse errors.
- Updated `GetEnvVars` to return an error and migrated all callers.
- Changed migration no-op `up`/`down` paths to warning envelopes with exit code 2.
- Avoided schema table creation for `migrate status`; only mutating runtime commands ensure the table.
- Documented migration status/no-op semantics.

Validation:
- `rg -n --glob '*.go' 'GetEnvVars' .`
- `go test ./commands ./internal/configmgr ./internal/migrate`
- `go test ./...`
- `go vet ./...`
