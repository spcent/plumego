# Card 0742

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/internal/scaffold
Owned Files: cmd/plumego/internal/scaffold/scaffold.go, cmd/plumego/internal/scaffold/scaffold_test.go, cmd/plumego/internal/codegen/codegen.go, cmd/plumego/internal/codegen/codegen_test.go, cmd/plumego/commands/new.go, cmd/plumego/commands/cli_e2e_test.go
Depends On: 0741

Goal:
Validate scaffold/codegen inputs and make force overwrite behavior less surprising.

Scope:
- Reject invalid project names and invalid module paths before writing scaffold files.
- Avoid leaving stale scaffold files when `--force` overwrites an existing project directory.
- Reject codegen output paths that point at directories or invalid package/name inputs.
- Add focused tests for validation and force overwrite behavior.

Non-goals:
- Do not redesign template storage.
- Do not change supported template names.
- Do not add external module path validation packages.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/internal/codegen/codegen.go`
- `cmd/plumego/internal/codegen/codegen_test.go`
- `cmd/plumego/commands/new.go`
- `cmd/plumego/commands/cli_e2e_test.go`

Tests:
- `go test ./internal/scaffold ./internal/codegen ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- None unless CLI user-facing validation changes need documentation.

Done Definition:
- Scaffold rejects invalid names/modules before writes.
- Forced scaffold regeneration does not leave stale files from another template.
- Codegen reports invalid output paths clearly.

Outcome:
- Added scaffold project name and module path validation before output directories are created.
- Added `ProjectOptions.CleanExisting` and wired `plumego new --force` to remove stale known template files before regeneration.
- Added codegen output path validation that rejects directories and preserves existing-file force behavior.
- Added focused scaffold, codegen, and CLI tests for invalid inputs and stale template cleanup.

Validation:
- `go test ./internal/scaffold ./internal/codegen ./commands`
- `go test ./...`
- `go vet ./...`
