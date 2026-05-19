# Card 1131

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: done
Primary Module: cmd/plumego command help
Owned Files: cmd/plumego/commands/*.go, cmd/plumego/commands/*_test.go
Depends On: 0751

Goal:
Remove current help drift by making command help metadata explicit and command-owned.

Scope:
- Introduce command help metadata types and renderer.
- Let each stable command declare its command flags, subcommands, arguments, and examples.
- Update help output to render metadata instead of a handwritten command switch.
- Add regression tests for current drift: config `--dir`, generate `model`, serve `[directory]`.

Non-goals:
- Do not change command runtime behavior beyond help text.
- Do not generate docs from source in this card.
- Do not change root command list semantics.

Files:
- `cmd/plumego/commands/root.go`
- `cmd/plumego/commands/*.go`
- `cmd/plumego/commands/cli_e2e_test.go`

Tests:
- `go test ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required unless visible command descriptions change.

Done Definition:
- Root only renders command-owned help metadata.
- Help output lists the currently supported config, generate, and serve contracts.
- Tests catch the known help drift cases.

Outcome:
- Added command help metadata types, renderer, and stable command ordering.
- Moved command-specific help declarations onto command types via `Help()`.
- Removed the root command's handwritten command help switch.
- Updated root help to render registered command descriptions from the command
  registry.
- Corrected generate command summary to include models.
- Added regression coverage for config `--dir`, generate model examples, and
  serve `[directory]` help text.

Validation:
- `go test ./commands` from `cmd/plumego`
- `go test ./...` from `cmd/plumego`
- `go vet ./...` from `cmd/plumego`
