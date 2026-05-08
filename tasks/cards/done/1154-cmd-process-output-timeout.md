# Card 1154

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: active
Primary Module: cmd/plumego external process
Owned Files: cmd/plumego/internal/executil/*, cmd/plumego/commands/build.go, cmd/plumego/commands/test.go, cmd/plumego/internal/checker/checker.go
Depends On: 0753

Goal:
Make CLI external process execution bounded, cancellable, and consistent.

Scope:
- Add an internal helper for bounded stdout/stderr capture with context timeout.
- Use the helper in `build`, `test`, and checker dependency commands.
- Avoid unbounded output in JSON/YAML payloads.
- Add tests for output truncation and timeout behavior.

Non-goals:
- Do not redesign build or test payload schemas except adding truncation markers.
- Do not change default build/test flags.
- Do not add dependencies.

Files:
- `cmd/plumego/internal/executil/executil.go`
- `cmd/plumego/internal/executil/executil_test.go`
- `cmd/plumego/commands/build.go`
- `cmd/plumego/commands/test.go`
- `cmd/plumego/internal/checker/checker.go`

Tests:
- `go test ./internal/executil ./commands ./internal/checker`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required unless command output fields change materially.

Done Definition:
- Build/test/checker commands use bounded output capture.
- Long-running external processes are cancellable by timeout.
- Truncated output is explicit in payloads.

Outcome:
- Added `internal/executil` with `CommandContext`, optional timeout, bounded
  stdout/stderr capture, truncation markers, and timeout reporting.
- Switched `build` to bounded output capture and a 10 minute process timeout.
- Switched `test` to bounded output capture and a process timeout derived from
  the requested Go test timeout.
- Switched checker dependency commands to bounded output capture with explicit
  timeouts.
- Added executil tests for truncation and timeout behavior.

Validation:
- `go test ./internal/executil ./commands ./internal/checker` from `cmd/plumego`
- `go test ./...` from `cmd/plumego`
- `go vet ./...` from `cmd/plumego`
