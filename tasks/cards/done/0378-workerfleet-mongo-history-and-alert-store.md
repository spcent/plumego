# Card 0378

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/platform/store/mongo
Owned Files:
- `reference/workerfleet/internal/platform/store/mongo/task_history_store.go`
- `reference/workerfleet/internal/platform/store/mongo/worker_event_store.go`
- `reference/workerfleet/internal/platform/store/mongo/alert_store.go`
- `reference/workerfleet/docs/storage.md`
- `reference/workerfleet/docs/alerts.md`
Depends On:
- `tasks/cards/done/0376-workerfleet-mongo-schema-and-indexes.md`
- `tasks/cards/done/0377-workerfleet-mongo-snapshot-and-active-task-store.md`
Blocked By:
- workerfleet MongoDB design review approval

Goal:
- Implement MongoDB-backed history and alert persistence so workerfleet keeps seven days of task history, worker events, and alert records.
- Preserve append-only history semantics and the existing alert dedupe lifecycle.

Scope:
- Add Mongo repositories for task history, worker events, and alert records.
- Persist firing and resolved alerts using the existing domain alert model.
- Reuse the schema and TTL rules from card 2011.

Non-goals:
- Do not add heartbeat raw sample storage.
- Do not alter alert type enums or domain event names.
- Do not change the handler API.

Files:
- `reference/workerfleet/internal/platform/store/mongo/task_history_store.go`
- `reference/workerfleet/internal/platform/store/mongo/worker_event_store.go`
- `reference/workerfleet/internal/platform/store/mongo/alert_store.go`
- `reference/workerfleet/docs/storage.md`
- `reference/workerfleet/docs/alerts.md`

Tests:
- `cd reference/workerfleet && go test ./internal/platform/store/mongo/...`
- `cd reference/workerfleet && go test ./internal/domain/...`
- Coverage for task-history append and lookup, event append, alert append/list, and resolved alert queries.

Docs Sync:
- Update `reference/workerfleet/docs/storage.md` with history collection behavior.
- Update `reference/workerfleet/docs/alerts.md` with Mongo-backed alert retention notes.

Done Definition:
- Task history, worker events, and alerts are persisted in MongoDB with seven-day retention fields.
- Latest-task fallback and alert-list queries work against Mongo data.
- Alert firing and resolved records remain queryable without changing domain logic.
- History persistence remains append-only from the app perspective.

Outcome:
- Implemented Mongo task-history, worker-event, and alert repositories using the existing app-local store interfaces.
- Added document-to-domain mappers for task history, worker events, and alert records.
- Preserved append-only history semantics and made duplicate generated document IDs idempotent for retry safety.
- Updated Mongo integration coverage for task history lookup, worker event listing, and firing/resolved alert filtering; tests remain skipped unless `WORKERFLEET_MONGO_TEST_URI` is set.
- Updated storage and alert docs with Mongo-backed retention behavior.
- Validation run:
  - `cd reference/workerfleet && go test ./internal/platform/store/mongo/...`
  - `cd reference/workerfleet && go test ./internal/domain/...`
  - `cd reference/workerfleet && go test ./...`
  - `cd reference/workerfleet && go build ./...`
