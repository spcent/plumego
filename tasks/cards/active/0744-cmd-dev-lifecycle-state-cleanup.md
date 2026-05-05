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
