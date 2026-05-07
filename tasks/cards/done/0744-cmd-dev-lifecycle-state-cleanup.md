# Card 0744

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: active
Primary Module: cmd/plumego dev lifecycle
Owned Files: cmd/plumego/commands/dev.go, cmd/plumego/commands/dev_test.go, cmd/plumego/internal/devserver/runner.go, cmd/plumego/internal/devserver/runner_test.go
Depends On: 0743

Goal:
Close dev command lifecycle and runner state restoration gaps.

Scope:
- Ensure dashboard, watcher, and signal notifications are cleaned up on every dev command exit path.
- Restore runner state if process pipe setup fails during start.
- Add focused tests for cleanup/state restoration regressions where possible.

Non-goals:
- Do not redesign the dashboard API.
- Do not change watcher polling behavior.
- Do not add external process management dependencies.

Files:
- `cmd/plumego/commands/dev.go`
- `cmd/plumego/commands/dev_test.go`
- `cmd/plumego/internal/devserver/runner.go`
- `cmd/plumego/internal/devserver/runner_test.go`

Tests:
- `go test ./commands ./internal/devserver`
- `go test ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`

Docs Sync:
- Not required unless CLI-visible behavior changes.

Done Definition:
- `dev` stops dashboard resources on watcher/event errors and signal exits.
- `dev` always unregisters signal notifications after registration.
- `AppRunner.Start` leaves no stale `starting` state when startup setup fails.

Outcome:
- Added dashboard stop cleanup after successful dev dashboard start so watcher,
  event, no-reload, signal, and context-cancel exits all close resources.
- Registered signal notifications with deferred cleanup in both dev wait paths.
- Added reload-mode context cancellation handling instead of leaving the loop
  blocked after caller cancellation.
- Restored `AppRunner` startup state through a single deferred failure cleanup
  before the process is fully started.
- Added regression tests for dev shutdown cleanup and runner restart after a
  failed start attempt.

Validation:
- `go test ./commands ./internal/devserver` from `cmd/plumego`
- `go test ./...` from `cmd/plumego`
- `go vet ./...` from `cmd/plumego`
