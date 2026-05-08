# Card 0730

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/new.go, cmd/plumego/commands/config.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/internal/output/formatter_test.go
Depends On: 0729

Goal:
Make successful machine-readable command output consistently use the CLI command-result envelope.

Scope:
- Convert success paths that still call `Print` directly for command results to `Success`.
- Keep plain help text as direct print output.
- Add tests for `new --dry-run`, `config show`, and `config env` top-level output shape.

Non-goals:
- Do not change the command-result schema.
- Do not change text help output.
- Do not redesign formatter internals beyond required tests.

Files:
- `cmd/plumego/commands/new.go`
- `cmd/plumego/commands/config.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/internal/output/formatter_test.go`

Tests:
- `go test ./commands ./internal/output`
- `go build .`

Docs Sync:
- None unless behavior text changes.

Done Definition:
- Stable success command outputs have `status`, `message`, and `data`.
- Help text remains human-readable.
- Focused CLI tests lock the envelope for affected commands.

Outcome:
- Converted `new --dry-run`, `config show`, and `config env` to success command-result envelopes.
- Preserved direct text output for help paths.
- Updated CLI tests for the affected JSON output shapes.

Validation:
- `go test ./commands ./internal/output`
- `go build .`
