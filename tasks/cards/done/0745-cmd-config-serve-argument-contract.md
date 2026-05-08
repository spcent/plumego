# Card 0745

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: active
Primary Module: cmd/plumego command parsing
Owned Files: cmd/plumego/commands/config.go, cmd/plumego/commands/context.go, cmd/plumego/commands/serve.go, cmd/plumego/commands/root.go, cmd/plumego/commands/cli_e2e_test.go
Depends On: 0744

Goal:
Align config subcommands and serve help with the command argument contract.

Scope:
- Add strict extra positional argument validation for config subcommands.
- Add `--dir` support to config subcommands where project directory matters.
- Route `serve --help` through the same help envelope used by root command help.
- Tighten `resolveDir` to reject non-directory and stat-error paths.

Non-goals:
- Do not redesign config manager semantics.
- Do not change serve runtime lifecycle.
- Do not add new global flags.

Files:
- `cmd/plumego/commands/config.go`
- `cmd/plumego/commands/context.go`
- `cmd/plumego/commands/serve.go`
- `cmd/plumego/commands/root.go`
- `cmd/plumego/commands/cli_e2e_test.go`

Tests:
- `go test ./commands`
- `go test ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`

Docs Sync:
- Not required unless help text or README-visible behavior changes.

Done Definition:
- Config subcommands reject unexpected positional args.
- Config subcommands support project directory resolution consistently.
- `serve --help` emits the same JSON/YAML/text help shape as other commands.
- `--dir` rejects file paths and returns clear stat errors.

Outcome:
- Added shared config subcommand parsing with `--dir` support and strict
  unexpected positional argument rejection.
- Routed config show, validate, init, and env through `resolveDir` instead of
  implicit `os.Getwd()` calls.
- Tightened `resolveDir` to report stat errors and reject file paths.
- Replaced `serve` command-local raw help output with the canonical command
  help envelope for text, JSON, and YAML formats.
- Added CLI regression tests for config `--dir`, config extra args, serve file
  path rejection, and serve machine-readable help.

Validation:
- `go test ./commands` from `cmd/plumego`
- `go test ./...` from `cmd/plumego`
- `go vet ./...` from `cmd/plumego`
