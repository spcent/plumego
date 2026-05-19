# Card 1262

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: done
Primary Module: cmd/plumego help
Owned Files: cmd/plumego/commands/help.go, cmd/plumego/commands/root_help_test.go, cmd/plumego/commands/config_cli_test.go
Depends On: 0763

Goal:
Reduce help metadata drift risk with automated command flag contract checks.

Scope:
- Add tests comparing command help flag entries against command `FlagSet` declarations.
- Make config help subcommand-specific enough to avoid implying unsupported flags.
- Keep current text rendering style.

Non-goals:
- Do not replace command parsing with a new framework.
- Do not redesign root help.
- Do not change machine help envelope shape.

Files:
- `cmd/plumego/commands/help.go`
- `cmd/plumego/commands/root_help_test.go`
- `cmd/plumego/commands/config_cli_test.go`

Tests:
- `go test ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required.

Done Definition:
- Missing or stale help flags fail tests.
- Config help no longer presents `show`-only flags as universally accepted.

Outcome:
- Added a command help drift test that parses each command source file for declared `FlagSet` flags and fails when help omits active flags.
- Made `config` help subcommand-specific so `show`-only flags are not presented as universal config flags.
- Updated `routes` help to list unsupported `--middleware` and `--group` flags explicitly, matching current parser behavior.

Validation:
- `go test ./commands`
- `go test ./...`
- `go vet ./...`
