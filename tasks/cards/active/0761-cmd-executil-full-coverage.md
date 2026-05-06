# Card 0761

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: active
Primary Module: cmd/plumego executil
Owned Files: cmd/plumego/commands/build.go, cmd/plumego/commands/test.go, cmd/plumego/internal/scaffold/scaffold.go, cmd/plumego/internal/executil/executil.go, cmd/plumego/internal/executil/executil_test.go
Depends On: 0760

Goal:
Route remaining CLI helper external commands through bounded, timeout-aware execution.

Scope:
- Use `executil` for build metadata helpers (`go version`, `git rev-parse`).
- Use `executil` for coverage parsing command execution.
- Use `executil` for scaffold `git init`.
- Keep output limits and timeout behavior consistent.

Non-goals:
- Do not change the primary build/test command output schemas.
- Do not change generated project contents.
- Do not add dependencies.

Files:
- `cmd/plumego/commands/build.go`
- `cmd/plumego/commands/test.go`
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/executil/executil.go`
- `cmd/plumego/internal/executil/executil_test.go`

Tests:
- `go test ./internal/executil ./internal/scaffold ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required.

Done Definition:
- Non-test direct `exec.Command` call sites in build/test/scaffold are replaced or justified.
- Helper command failures include bounded diagnostic output where useful.

Outcome:
