# Card 0815

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/internal/codegen
Owned Files: cmd/plumego/internal/codegen/codegen.go, cmd/plumego/internal/codegen/codegen_test.go, cmd/plumego/commands/generate.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/README.md
Depends On: 0724

Goal:
Align generated code with the canonical scaffold layout and fail closed on invalid generator input.

Scope:
- Change default handler and middleware paths to canonical app layout.
- Validate generated type names, package names, and HTTP methods before writing files.
- Return an error for unsupported generator types or methods instead of silently emitting partial code.
- Update docs and CLI tests for supported generator types.

Non-goals:
- Do not add a full AST code insertion engine.
- Do not auto-wire generated handlers into routes.
- Do not add non-stdlib formatting dependencies.

Files:
- `cmd/plumego/internal/codegen/codegen.go`
- `cmd/plumego/internal/codegen/codegen_test.go`
- `cmd/plumego/commands/generate.go`
- `cmd/plumego/commands/cli_e2e_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./internal/codegen ./commands`
- `go build .`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- Default generated files land in the same canonical layout taught by `plumego new`.
- Invalid generator input fails before file writes.
- Tests cover invalid names and unsupported HTTP methods.

Outcome:
- Moved default generated handler files to `internal/handler` and middleware
  files to `internal/middleware`.
- Added generator validation for Go identifiers, package names, and supported
  handler methods before file creation.
- Removed the stale `generate component` README example and documented supported
  generator defaults.
- Added codegen and CLI tests for canonical paths and invalid methods.
- Validation Run:
  - `go test ./internal/codegen ./commands`
  - `go build .`
