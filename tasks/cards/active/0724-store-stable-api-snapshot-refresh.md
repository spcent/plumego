# Card 0724: Store Stable API Snapshot Refresh

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P0
State: active
Primary Module: store
Owned Files:
- docs/stable-api/snapshots/store-head.snapshot
Depends On:
- 0723

Goal:
Refresh the stable store API snapshot so release evidence matches the actual current public surface.

Scope:
- Regenerate `docs/stable-api/snapshots/store-head.snapshot` from the current `store/...` packages.
- Compare the refreshed snapshot against itself to verify snapshot-tool readability.
- Run store checks and required boundary checks.

Non-goals:
- Do not make additional API changes in this card.
- Do not refresh unrelated stable-root snapshots.

Files:
- docs/stable-api/snapshots/store-head.snapshot

Tests:
- go run ./internal/checks/extension-api-snapshot -module ./store/... -out docs/stable-api/snapshots/store-head.snapshot
- go run ./internal/checks/extension-api-snapshot -compare docs/stable-api/snapshots/store-head.snapshot docs/stable-api/snapshots/store-head.snapshot
- go test -timeout 20s ./store/...

Docs Sync:
- This card is the stable API evidence sync.

Done Definition:
- Store snapshot contains no stale removed symbols.
- Snapshot compare succeeds.
- Store tests and boundary checks pass.

Outcome:
