# Card 0728

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/migrate.go, cmd/plumego/internal/migrate/migrate.go, cmd/plumego/internal/migrate/migrate_test.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md
Depends On: 0727

Goal:
Make the `migrate` command's stable boundary explicit and fail predictably for unsupported runtime database use.

Scope:
- Keep offline migration file creation supported.
- Validate driver/runtime database operations before attempting partial work.
- Improve migration file version collision behavior.
- Document supported and unsupported migration command behavior.

Non-goals:
- Do not add database driver dependencies to the CLI module.
- Do not implement a full migration locking framework.
- Do not change SQL dialect semantics beyond fail-closed validation.

Files:
- `cmd/plumego/commands/migrate.go`
- `cmd/plumego/internal/migrate/migrate.go`
- `cmd/plumego/internal/migrate/migrate_test.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./internal/migrate ./commands`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- `migrate create` remains deterministic and collision-resistant.
- Runtime database operations fail early with clear unsupported-driver guidance when no driver is available.
- Docs do not imply bundled DB driver support.
