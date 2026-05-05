# Card 0757

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: done
Primary Module: cmd/plumego migrate
Owned Files: cmd/plumego/commands/migrate.go, cmd/plumego/commands/cli_e2e_test.go
Depends On: 0756

Goal:
Make migration step semantics explicit and fail closed.

Scope:
- Reject negative `--steps`.
- Reject `--steps` for `status`.
- Keep `--steps 0` as all for `up` and `down`.
- Add CLI tests for invalid combinations.

Non-goals:
- Do not change migration storage schema.
- Do not add database drivers.
- Do not change migration file naming.

Files:
- `cmd/plumego/commands/migrate.go`
- `cmd/plumego/commands/cli_e2e_test.go`

Tests:
- `go test ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required unless README command examples change.

Done Definition:
- Negative steps and status steps fail with structured errors.
- Valid up/down semantics are unchanged.

Outcome:
- `migrate` now rejects negative `--steps` before runtime database checks.
- `migrate status` now rejects any explicit `--steps`, including `--steps 0`,
  so status cannot silently accept meaningless step limits.
- Added CLI JSON-envelope regression tests for both invalid combinations.

Validation:
- `go test ./commands`
- `go test ./...`
- `go vet ./...`
