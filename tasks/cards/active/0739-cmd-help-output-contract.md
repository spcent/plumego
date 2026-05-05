# Card 0739

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: active
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/commands/root.go, cmd/plumego/internal/output/formatter.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md, cmd/plumego/MODULE.md
Depends On: 0738

Goal:
Keep command help synchronized with real flags and make output behavior explicit across JSON/YAML/text modes.

Scope:
- Remove help entries for unsupported flags and add missing supported flags.
- Avoid duplicate nested usage blocks in command help.
- Define and test stable help output behavior for machine formats.
- Improve text mode command-result rendering if needed to avoid Go struct dumps.

Non-goals:
- Do not replace the entire command framework.
- Do not change command execution data schemas beyond help/text rendering.
- Do not add third-party CLI packages.

Files:
- `cmd/plumego/commands/root.go`
- `cmd/plumego/internal/output/formatter.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`
- `cmd/plumego/MODULE.md`

Tests:
- `go test ./commands ./internal/output`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- `cmd/plumego/README.md`
- `cmd/plumego/MODULE.md`

Done Definition:
- Help text matches implemented flags.
- Help smoke tests cover JSON/YAML/text expectations.
- Text output does not expose raw Go struct formatting for command-result envelopes.

