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
- `case_step_history` equivalent state is exposed through the optional `CaseStepHistoryStore` interface and is currently implemented by the memory backend.

Case step drilldown:

- Worker heartbeats report per-case `exec_plan_id` and `current_step`.
- Domain reconciliation emits `task_step_changed` and `task_step_finished` events when a case changes step or finishes a step.
- The memory backend materializes those step events into case step history records with worker, pod, node, exec plan, step, result, attempt, and low-cardinality `error_class` fields.
- The app service exposes case timeline and exec-plan drilldown query methods over `CaseStepHistoryStore`.
- Drilldown filters are intentionally business-level and topology-level: `task_id`, `worker_id`, `exec_plan_id`, `node_name`, `pod_name`, and `step`.
- Prometheus remains aggregate-only and must not use `case_id` or `task_id` labels; Grafana panels should link from abnormal aggregate dimensions to the app drilldown API.

History and retention:

- task history records are retained for seven days
- case step history records are retained for seven days
- worker event records are retained for seven days
- alert records are retained for seven days
- current worker snapshots and active-task sets are not pruned by retention

MongoDB layout:

- `worker_snapshots` stores one current snapshot document per worker and keeps embedded active-task details, including `exec_plan_id` and `current_step`, for worker detail reads.
- `worker_active_tasks` stores one current projection document per active task, including `exec_plan_id` and `current_step`, and supports `task_id` reverse lookup with a unique index.
- `task_history`, `worker_events`, and `alert_events` use `expire_at` as the TTL field for seven-day retention.
- `task_history`, `worker_events`, and `alert_events` are append-only from the app perspective; task history records include the latest `exec_plan_id` and `current_step` snapshot. Duplicate generated document IDs are ignored to keep retries idempotent.
- MongoDB document structs and index bootstrap live under `internal/platform/store/mongo` and remain separate from `internal/domain`.
- Mongo case step history persistence is intentionally deferred until the memory-backed `CaseStepHistoryStore` interface proves stable. The target collection should be append-only with `expire_at` TTL and indexes for `task_id`, `exec_plan_id + observed_at`, `node_name + observed_at`, `pod_name + observed_at`, and `step + observed_at`.

Startup behavior:

- `WORKERFLEET_STORE_BACKEND=memory` wires the in-memory backend and requires no external dependencies.
- `WORKERFLEET_STORE_BACKEND=mongo` opens the MongoDB client, pings the primary, ensures indexes, and then wires the Mongo repositories.
- missing Mongo URI or database settings fail startup before any handler is exposed.
- retention defaults to seven days and can be overridden with `WORKERFLEET_RETENTION_DAYS`; values must be greater than zero and no more than 106751 days.

This layout is intentionally app-local to `reference/workerfleet` and does not widen Plumego stable store packages.
