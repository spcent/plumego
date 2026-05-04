# Card 0723

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/checks
Owned Files: internal/checks/extension-api-snapshot/main.go, internal/checks/extension-api-snapshot/README.md, docs/stable-api/snapshots/core-head.snapshot
Depends On: 0722-router-any-sentinel-export

Goal:
Keep stable API snapshots focused on exported compatibility surface instead of unexported struct internals.

Scope:
Update the exported API snapshot tool so exported struct snapshots omit unexported fields.
Regenerate the core stable API snapshot to remove internal `App` implementation details.
Document the snapshot tool behavior.

Non-goals:
Do not change runtime code.
Do not regenerate every existing snapshot in this card.
Do not add release evidence claims.

Files:
internal/checks/extension-api-snapshot/main.go
internal/checks/extension-api-snapshot/README.md
docs/stable-api/snapshots/core-head.snapshot

Tests:
go test -timeout 20s ./internal/checks/extension-api-snapshot
go run ./internal/checks/extension-api-snapshot -module ./core -out docs/stable-api/snapshots/core-head.snapshot
go run ./internal/checks/dependency-rules

Docs Sync:
Update `internal/checks/extension-api-snapshot/README.md`.

Done Definition:
Core snapshot no longer records `App` unexported fields.
Snapshot tool tests pass.
Dependency check passes.

Outcome:
