# Card 2003

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: active
Primary Module: reference/workerfleet/internal/platform/store
Owned Files:
- `reference/workerfleet/internal/platform/store/worker_snapshot_store.go`
- `reference/workerfleet/internal/platform/store/active_task_store.go`
- `reference/workerfleet/internal/platform/store/task_history_store.go`
- `reference/workerfleet/internal/platform/store/alert_store.go`
- `reference/workerfleet/internal/platform/store/retention.go`
Depends On:
- `tasks/cards/active/2001-workerfleet-domain-status-model.md`
- `tasks/cards/active/2002-workerfleet-api-contract-and-routes.md`

Goal:
- Define and implement the app-local storage layer for current worker state, active tasks, task history, worker event history, alert events, and seven-day retention.

Scope:
- Store interfaces and concrete implementations for current-state upsert and history append.
- Seven-day retention for task history, worker event history, and alert events.
- Query methods needed by fleet summary, worker detail, task lookup, and alert listing.

Non-goals:
- Do not implement Kubernetes watch logic.
- Do not add alert notification delivery.
- Do not change Plumego stable store packages.

Files:
- `reference/workerfleet/internal/platform/store/worker_snapshot_store.go`
- `reference/workerfleet/internal/platform/store/active_task_store.go`
- `reference/workerfleet/internal/platform/store/task_history_store.go`
- `reference/workerfleet/internal/platform/store/alert_store.go`
- `reference/workerfleet/internal/platform/store/retention.go`

Tests:
- `go test ./reference/workerfleet/internal/platform/store/...`
- Upsert tests for worker snapshots and full active-task set replacement.
- Retention tests proving seven-day cleanup for history tables and current-state preservation.

Docs Sync:
- Update `reference/workerfleet/README.md` or `reference/workerfleet/docs/storage.md` if storage schema or retention behavior is documented for operators.

Done Definition:
- The service can persist current worker state and recover active-task lists without relying on raw heartbeat replay.
- History and alert/event records expire after seven days by policy.
- Store interfaces are app-local and do not widen stable Plumego packages.

Outcome:
- Pending execution.
