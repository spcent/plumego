# Workerfleet Storage

The workerfleet storage layer is app-local and supports memory and MongoDB
backends.

Submodule dependency policy:

- `reference/workerfleet` is expected to build as the `workerfleet` Go module
- app-local code imports workerfleet packages as `workerfleet/...`
- Plumego stable roots are consumed from the repository root through a local `replace`
- persistence dependencies for workerfleet are submodule-local and must not change the repository root `go.mod`

Current state:

- `internal/platform/store` owns app-local persistence interfaces and shared query/filter types.
- `internal/platform/store/memory` owns the local in-memory backend implementation.
- `internal/platform/store/mongo` owns the production-oriented MongoDB backend implementation.
- `worker_snapshot` equivalent state is stored as a per-worker snapshot.
- `worker_active_task` equivalent state is stored as a per-worker active-task set plus a task index.
- `case_step_history` equivalent state is exposed through the `CaseStepHistoryStore` interface and is implemented by both memory and MongoDB backends.

Case step drilldown:

- Worker heartbeats report per-case `exec_plan_id` and `current_step`.
- Domain reconciliation emits `task_step_changed` and `task_step_finished` events when a case changes step or finishes a step.
- Memory and MongoDB backends materialize those step events into case step history records with worker, pod, node, exec plan, step, result, attempt, and low-cardinality `error_class` fields.
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
- `task_history`, `case_step_history`, `worker_events`, and `alert_events` use `expire_at` as the TTL field for seven-day retention.
- `task_history`, `case_step_history`, `worker_events`, and `alert_events` are append-only from the app perspective; task history records include the latest `exec_plan_id` and `current_step` snapshot. Duplicate generated document IDs are ignored to keep retries idempotent.
- MongoDB document structs and index bootstrap live under `internal/platform/store/mongo` and remain separate from `internal/domain`.
- `case_step_history` is append-only with `expire_at` TTL and indexes for `task_id + observed_at`, `exec_plan_id + observed_at`, `node_name + observed_at`, `pod_name + observed_at`, and `step + observed_at`.
- `loop_leases` stores Mongo-backed runtime loop ownership by loop name, owner ID, and `expires_at` so kube sync, status sweep, and alert evaluation can coordinate across replicas.

Startup behavior:

- `WORKERFLEET_STORE_BACKEND=memory` wires the in-memory backend and requires no external dependencies.
- `WORKERFLEET_STORE_BACKEND=mongo` opens the MongoDB client, pings the primary, ensures indexes, and then wires the Mongo repositories.
- missing Mongo URI or database settings fail startup before any handler is exposed.
- retention defaults to seven days and can be overridden with `WORKERFLEET_RETENTION_DAYS`; values must be greater than zero and no more than 106751 days.
- when Mongo storage is enabled, the app wires a Mongo `LoopLeaseCoordinator` using `WORKERFLEET_LOOP_LEASE_OWNER` and `WORKERFLEET_LOOP_LEASE_TTL`.

Mongo integration test gate:

- Default repo gates do not require a local MongoDB instance.
- `make workerfleet-mongo-test` is the repo-native optional gate for real MongoDB store behavior.
- When `WORKERFLEET_MONGO_TEST_URI` is unset, the target prints a skip message with the exact environment variable to set.
- When `WORKERFLEET_MONGO_TEST_URI` is set, the target runs `cd reference/workerfleet && go test -timeout 60s ./internal/platform/store/mongo`.
- Run this gate before changing Mongo indexes, document mapping, retention fields, or idempotency behavior.

This layout is intentionally app-local to `reference/workerfleet` and does not widen Plumego stable store packages.
