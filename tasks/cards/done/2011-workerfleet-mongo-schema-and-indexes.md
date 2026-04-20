# Card 2011

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/platform/store/mongo
Owned Files:
- `reference/workerfleet/internal/platform/store/mongo/bootstrap.go`
- `reference/workerfleet/internal/platform/store/mongo/indexes.go`
- `reference/workerfleet/internal/platform/store/mongo/worker_snapshot_doc.go`
- `reference/workerfleet/internal/platform/store/mongo/worker_active_task_doc.go`
- `reference/workerfleet/internal/platform/store/mongo/task_history_doc.go`
- `reference/workerfleet/internal/platform/store/mongo/worker_event_doc.go`
- `reference/workerfleet/internal/platform/store/mongo/alert_event_doc.go`
- `reference/workerfleet/docs/storage.md`
Depends On:
- `tasks/cards/active/2009-workerfleet-submodule-bootstrap.md`
- `tasks/cards/active/2010-workerfleet-store-interface-split.md`
Blocked By:
- workerfleet MongoDB design review approval

Goal:
- Define the MongoDB document model, collection names, index set, and TTL strategy for workerfleet current state and seven-day history.
- Keep Mongo-specific document mapping fully isolated from `internal/domain`.

Scope:
- Create MongoDB document types for worker snapshots, active tasks, task history, worker events, and alert events.
- Add idempotent collection and index bootstrap logic.
- Define seven-day TTL behavior using `expire_at` for history collections.

Non-goals:
- Do not wire repositories into the app.
- Do not add heartbeat raw sample retention.
- Do not place `bson` tags on domain types.

Files:
- `reference/workerfleet/internal/platform/store/mongo/bootstrap.go`
- `reference/workerfleet/internal/platform/store/mongo/indexes.go`
- `reference/workerfleet/internal/platform/store/mongo/worker_snapshot_doc.go`
- `reference/workerfleet/internal/platform/store/mongo/worker_active_task_doc.go`
- `reference/workerfleet/internal/platform/store/mongo/task_history_doc.go`
- `reference/workerfleet/internal/platform/store/mongo/worker_event_doc.go`
- `reference/workerfleet/internal/platform/store/mongo/alert_event_doc.go`
- `reference/workerfleet/docs/storage.md`

Tests:
- `cd reference/workerfleet && go test ./internal/platform/store/mongo/...`
- Focused tests for document mapping and index bootstrap idempotence.

Docs Sync:
- Update `reference/workerfleet/docs/storage.md` with collection layout, index list, and TTL rules.

Done Definition:
- MongoDB document types are defined separately from domain types.
- `EnsureIndexes(ctx)` is repeatable and covers all required unique and TTL indexes.
- History retention is modeled with `expire_at` on the history collections only.
- The storage document design is documented well enough for repository implementation cards to reuse directly.

Outcome:
- Added the app-local MongoDB package under `reference/workerfleet/internal/platform/store/mongo` with collection constants, document structs, and index specifications.
- Implemented `EnsureIndexes(ctx, db)` plus reusable `IndexSpec` definitions for current-state, unique lookup, and seven-day TTL indexes.
- Added focused tests for document mapping and index specification coverage.
- Updated `reference/workerfleet/docs/storage.md` with the MongoDB collection layout and TTL strategy.
- Validation run:
  - `cd reference/workerfleet && go mod tidy`
  - `cd reference/workerfleet && go test ./internal/platform/store/mongo/...`
  - `cd reference/workerfleet && go test ./...`
