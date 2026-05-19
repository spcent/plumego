# Card 0785

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: cmd/plumego/internal/devserver
Owned Files: cmd/plumego/internal/devserver/runner.go, cmd/plumego/internal/devserver/runner_test.go
Depends On: 0720

Goal:
Make `plumego dev` app process ownership deterministic by ensuring only one path waits on the child process.

Scope:
- Introduce a single wait owner for the app process.
- Make `Stop` signal the process and wait through shared process completion state instead of calling `Wait` again.
- Avoid holding the runner mutex while waiting for process exit.
- Add regression coverage for stop/restart behavior.

Non-goals:
- Do not redesign builder behavior.
- Do not change dashboard API response DTOs.
- Do not add non-stdlib process management dependencies.

Files:
- `cmd/plumego/internal/devserver/runner.go`
- `cmd/plumego/internal/devserver/runner_test.go`

Tests:
- `go test ./internal/devserver`
- `go test -race ./internal/devserver`
- `go build .`

Docs Sync:
None unless lifecycle behavior visible to users changes.

Done Definition:
- Runner cannot call `Wait` twice for the same process.
- Stop and restart remain idempotent for already-stopped processes.
- Focused devserver tests pass with race detection.

Outcome:
- Added a per-process `waitDone` channel owned by the background wait goroutine.
- Updated `Stop` to signal the process and wait through the shared completion
  channel without holding the runner mutex.
- Added runner lifecycle tests covering stop idempotence and second-start
  rejection while running.
- Validation Run:
  - `go test ./internal/devserver`
  - `go test -race ./internal/devserver`
  - `go build .`
