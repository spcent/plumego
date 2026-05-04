# Card 0723

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor.yaml
Priority: P1
State: active
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/root.go, cmd/plumego/commands/check.go, cmd/plumego/commands/config.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/internal/output/formatter.go
Depends On: 0722

Goal:
Converge CLI flag parsing and machine-readable output for non-success command states.

Scope:
- Support `--format=value` and `--env-file=value` global flags.
- Keep command-specific flags from being accidentally consumed as globals after the command token.
- Route warning/degraded command outcomes through a documented output envelope.
- Add CLI tests for global flag forms and warning/failure envelopes.

Non-goals:
- Do not replace the CLI with a third-party command framework.
- Do not change command names.
- Do not redesign streaming dev events.

Files:
- `cmd/plumego/commands/root.go`
- `cmd/plumego/commands/check.go`
- `cmd/plumego/commands/config.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/internal/output/formatter.go`

Tests:
- `go test ./commands ./internal/output`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md` if exit/output behavior changes.

Done Definition:
- Global flag parsing is predictable and tested.
- Non-zero warning/degraded outputs keep machine-readable structure.
- Existing successful command output remains compatible.
