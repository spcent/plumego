# Card 0738

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/dev.go, cmd/plumego/commands/routes.go, cmd/plumego/commands/test.go, cmd/plumego/commands/check.go, cmd/plumego/commands/cli_e2e_test.go
Depends On: 0737

Goal:
Align project command flag ordering and directory semantics for automation-friendly CLI use.

Scope:
- Use interspersed command flag parsing for `dev`, `routes`, and `test`.
- Add `check --dir <path>` while preserving current default cwd behavior.
- Reject unexpected positional arguments where the command has no positional contract.
- Add regression tests for mixed flag order and `check --dir`.

Non-goals:
- Do not change dev server runtime behavior.
- Do not change checker rule definitions.
- Do not introduce new dependencies.

Files:
- `cmd/plumego/commands/dev.go`
- `cmd/plumego/commands/routes.go`
- `cmd/plumego/commands/test.go`
- `cmd/plumego/commands/check.go`
- `cmd/plumego/commands/cli_e2e_test.go`

Tests:
- `go test ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- None unless usage text changes in a user-visible way.

Done Definition:
- Project commands accept command-local flags before or after their positional arguments consistently.
- `check --dir` validates a target project without changing cwd.
- Extra unsupported positionals return structured errors.

Outcome:
- Switched `dev`, `routes`, and `test` to interspersed command flag parsing.
- Added `check --dir` while preserving cwd as the default target.
- Added unexpected positional argument errors for `dev`, `routes`, and `check`.
- Added CLI tests for `check --dir`, package-before-flag `test`, and unexpected argument failures.

Validation:
- `go test ./commands`
- `go test ./...`
- `go vet ./...`
