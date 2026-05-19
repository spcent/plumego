# Card 0892

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/internal/watcher
Owned Files: cmd/plumego/internal/watcher/watcher.go, cmd/plumego/internal/watcher/watcher_test.go, cmd/plumego/DEV_SERVER.md
Depends On: 0731

Goal:
Make the polling watcher predictable enough for stable local development while keeping the module dependency footprint unchanged.

Scope:
- Make `Close` idempotent and close channels when the watcher stops.
- Surface walk/stat errors through the error channel.
- Preserve multiple changed paths within a debounce window instead of keeping only one pending path.
- Detect file deletions for watched files.

Non-goals:
- Do not add fsnotify or another dependency in this card.
- Do not change default watch/exclude flags.
- Do not redesign dev reload behavior.

Files:
- `cmd/plumego/internal/watcher/watcher.go`
- `cmd/plumego/internal/watcher/watcher_test.go`
- `cmd/plumego/DEV_SERVER.md`

Tests:
- `go test ./internal/watcher`
- `go test ./commands ./internal/watcher`
- `go build .`

Docs Sync:
- `cmd/plumego/DEV_SERVER.md`

Done Definition:
- Watcher close is safe to call more than once.
- Debounced events can report multiple changes.
- Delete and walk error behavior is tested.

Outcome:
- Made watcher close idempotent and close event/error channels when stopped.
- Added debounced multi-path event emission, watched deletion detection, and non-blocking walk error reporting.
- Documented the dependency-free polling watcher behavior.

Validation:
- `go test ./internal/watcher`
- `go test ./commands ./internal/watcher`
- `go build .`
