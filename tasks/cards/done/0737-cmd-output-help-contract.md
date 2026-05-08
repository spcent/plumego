# Card 0737

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: cmd/plumego
Owned Files: cmd/plumego/commands/root.go, cmd/plumego/internal/output/formatter.go, cmd/plumego/commands/*_test.go, cmd/plumego/internal/output/*_test.go
Depends On: 0714

Goal:
Make CLI output and help behavior predictable for automation.

Scope:
- Choose and enforce one default output format.
- Reject unsupported `--format` values instead of silently falling back.
- Make `plumego <command> --help` return usage with exit code 0.
- Keep success and error output envelope consistent for machine formats.

Non-goals:
- Do not introduce a third-party CLI framework.
- Do not change command business behavior.
- Do not change HTTP `contract` response shapes.

Files:
- `cmd/plumego/commands/root.go`
- `cmd/plumego/internal/output/formatter.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/internal/output/formatter_test.go`

Tests:
- `go test ./commands ./internal/output`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md` only if the default format or command list changes.

Done Definition:
- Help commands are successful and useful.
- Invalid formats fail closed.
- Tests cover the output contract.

Outcome:
- Switched the CLI default output format to JSON to match the documented
  machine-first contract.
- Added fail-closed validation for unsupported output formats.
- Added root-level handling for `plumego <command> --help` and `plumego help
  <command>` so command help exits successfully.
- Added regression tests for default JSON output, invalid formats, command help,
  and formatter supported formats.
- Validation Run:
  - `go test ./commands ./internal/output`
  - `go build .`
