# Card 0729

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/root.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md, cmd/plumego/MODULE.md
Depends On: 0728

Goal:
Freeze the visible CLI command surface and make command help predictable enough for stable automation.

Scope:
- Update README/MODULE command lists to match the actual registered commands.
- Add command-owned usage/help metadata without changing command execution semantics.
- Make `plumego <command> --help` show command-specific flags and subcommands where available.
- Add tests for top-level command list and at least one command-specific help output.

Non-goals:
- Do not add or remove commands.
- Do not introduce a third-party CLI framework.
- Do not change global flag parsing behavior.

Files:
- `cmd/plumego/commands/root.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`
- `cmd/plumego/MODULE.md`

Tests:
- `go test ./commands`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md`
- `cmd/plumego/MODULE.md`

Done Definition:
- Docs and top-level help agree on the stable command list.
- Command help includes command-specific usage for stable commands.
- Focused tests cover the command surface and detailed help output.
