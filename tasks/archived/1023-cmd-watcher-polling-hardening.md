# Card 1023

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P2
State: done
Primary Module: cmd/plumego/internal/watcher
Owned Files: cmd/plumego/internal/watcher/watcher.go, cmd/plumego/internal/watcher/watcher_test.go, cmd/plumego/commands/dev.go, cmd/plumego/commands/dev_test.go, cmd/plumego/README.md
Depends On: 0742

Goal:
Reduce watcher polling surprises while staying dependency-free.

Scope:
- Make polling interval configurable through internal watcher options or dev command wiring.
- Validate debounce/poll durations and reject non-positive values.
- Improve pattern matching tests around recursive includes/excludes.
- Document remaining polling-based watcher limits for large projects.

Non-goals:
- Do not add `fsnotify` or any other dependency.
- Do not redesign dev server lifecycle.
- Do not change dashboard APIs.

Files:
- `cmd/plumego/internal/watcher/watcher.go`
- `cmd/plumego/internal/watcher/watcher_test.go`
- `cmd/plumego/commands/dev.go`
- `cmd/plumego/commands/dev_test.go`
- `cmd/plumego/README.md`

Tests:
- `go test ./internal/watcher ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- `cmd/plumego/README.md`

Done Definition:
- Watcher rejects invalid timing inputs.
- Polling behavior is configurable and covered by tests.
- README states the dependency-free polling watcher tradeoff.

Outcome:
- Added watcher options with configurable polling interval and a 500ms default.
- Rejected non-positive debounce and poll durations.
- Added recursive include/exclude pattern tests and faster watcher tests using explicit poll options.
- Documented the dependency-free polling watcher tradeoff.

Validation:
- `go test ./internal/watcher ./commands`
- `go test ./...`
- `go vet ./...`
