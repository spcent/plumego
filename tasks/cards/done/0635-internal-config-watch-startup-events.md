# Card 0635

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: internal/config
Owned Files: internal/config/manager.go, internal/config/watch_test.go
Depends On:

Goal:
Prevent `Manager.StartWatchers` from dropping a watch result that is already
buffered when the source is registered.

Scope:
- Remove the non-blocking closed-channel probe that consumes one watch result.
- Let watcher goroutines exit naturally when a source returns a closed channel.
- Add focused coverage for a source whose watch channel has an immediate result.

Non-goals:
- Do not change the `Source` interface.
- Do not change file polling semantics.
- Do not change global config helpers.

Files:
- internal/config/manager.go
- internal/config/watch_test.go

Tests:
- go test ./internal/config
- go test ./internal/...

Docs Sync:
- None; this fixes existing runtime behavior without changing public docs.

Done Definition:
- Immediate watch results are applied and watcher callbacks fire.
- Closed watch channels still return cleanly.
- Focused config tests pass.

Outcome:
- Removed the lossy closed-channel probe in `StartWatchers`.
- Added coverage for a source that emits an immediate buffered watch result.
- Validation: `go test ./internal/config`; `go test ./internal/...`.
