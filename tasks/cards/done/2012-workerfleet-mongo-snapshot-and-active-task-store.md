# Card 2012

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/platform/store/mongo
Owned Files:
- `reference/workerfleet/internal/platform/store/mongo/client.go`
- `reference/workerfleet/internal/platform/store/mongo/mappers.go`
- `reference/workerfleet/internal/platform/store/mongo/worker_snapshot_store.go`
- `reference/workerfleet/internal/platform/store/mongo/active_task_store.go`
- `reference/workerfleet/internal/app/service_test.go`
Depends On:
- `tasks/cards/active/2010-workerfleet-store-interface-split.md`
- `tasks/cards/active/2011-workerfleet-mongo-schema-and-indexes.md`
Blocked By:
- workerfleet MongoDB design review approval

Goal:
- Implement the MongoDB-backed current-state repositories so worker list, worker detail, task reverse lookup, and fleet summary can run against persisted Mongo data.
- Preserve the current rule that `worker_snapshots` is the authority for worker state and `worker_active_tasks` is a projection for task lookup.

Scope:
- Add Mongo client wrapper and domain-to-doc mappers.
- Implement worker snapshot and active-task repositories against MongoDB.
- Keep read behavior aligned with the existing memory backend.

Non-goals:
- Do not implement history or alert repositories.
- Do not add runtime config switching yet.
- Do not move worker status evaluation into Mongo queries.

Files:
- `reference/workerfleet/internal/platform/store/mongo/client.go`
- `reference/workerfleet/internal/platform/store/mongo/mappers.go`
- `reference/workerfleet/internal/platform/store/mongo/worker_snapshot_store.go`
- `reference/workerfleet/internal/platform/store/mongo/active_task_store.go`
- `reference/workerfleet/internal/app/service_test.go`

Tests:
- `cd reference/workerfleet && go test ./internal/platform/store/mongo/...`
- `cd reference/workerfleet && go test ./internal/app/...`
- Integration coverage for snapshot upsert, filtered list queries, active-task replacement, task reverse lookup, and fleet summary.

Docs Sync:
- Update `reference/workerfleet/docs/storage.md` if implementation-level behavior differs from the approved schema card.

Done Definition:
- Worker snapshots and active tasks can be written to and read from MongoDB through app-local interfaces.
- Multi-task replacement semantics remain full-set replacement per worker.
- Query results from the Mongo backend match the existing memory backend for the same inputs.
- App-level tests demonstrate that current-state reads still work without transport changes.

Outcome:
- Added the Mongo current-state `Store` and client wrapper under `reference/workerfleet/internal/platform/store/mongo`.
- Implemented worker snapshot persistence, filtered snapshot listing, fleet counts, active-task replacement, active-task listing, and task reverse lookup.
- Added domain/document round-trip mappers and compile-time interface assertions for the Mongo store.
- Added an optional MongoDB integration test gated by `WORKERFLEET_MONGO_TEST_URI`; default local test runs skip the external integration path.
- Validation run:
  - `cd reference/workerfleet && go test ./internal/platform/store/mongo/...`
  - `cd reference/workerfleet && go test ./internal/app/...`
  - `cd reference/workerfleet && go test ./...`
  - `cd reference/workerfleet && go build ./...`
