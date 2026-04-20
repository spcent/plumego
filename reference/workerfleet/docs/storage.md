# Workerfleet Storage

The initial workerfleet storage layer is app-local and in-memory.

Submodule dependency policy:

- `reference/workerfleet` is expected to build as the `workerfleet` Go module
- app-local code imports workerfleet packages as `workerfleet/...`
- Plumego stable roots are consumed from the repository root through a local `replace`
- persistence dependencies for workerfleet are submodule-local and must not change the repository root `go.mod`

Current state:

- `internal/platform/store` owns app-local persistence interfaces and shared query/filter types.
- `internal/platform/store/memory` owns the current in-memory backend implementation.
- `worker_snapshot` equivalent state is stored as a per-worker snapshot.
- `worker_active_task` equivalent state is stored as a per-worker active-task set plus a task index.

History and retention:

- task history records are retained for seven days
- worker event records are retained for seven days
- alert records are retained for seven days
- current worker snapshots and active-task sets are not pruned by retention

MongoDB layout:

- `worker_snapshots` stores one current snapshot document per worker and keeps embedded active-task details for worker detail reads.
- `worker_active_tasks` stores one current projection document per active task and supports `task_id` reverse lookup with a unique index.
- `task_history`, `worker_events`, and `alert_events` use `expire_at` as the TTL field for seven-day retention.
- `task_history`, `worker_events`, and `alert_events` are append-only from the app perspective; duplicate generated document IDs are ignored to keep retries idempotent.
- MongoDB document structs and index bootstrap live under `internal/platform/store/mongo` and remain separate from `internal/domain`.

Startup behavior:

- `WORKERFLEET_STORE_BACKEND=memory` wires the in-memory backend and requires no external dependencies.
- `WORKERFLEET_STORE_BACKEND=mongo` opens the MongoDB client, pings the primary, ensures indexes, and then wires the Mongo repositories.
- missing Mongo URI or database settings fail startup before any handler is exposed.
- retention defaults to seven days and can be overridden with `WORKERFLEET_RETENTION_DAYS`.

This layout is intentionally app-local to `reference/workerfleet` and does not widen Plumego stable store packages.
