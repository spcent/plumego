# Workerfleet Technical Design

## 1. Background

Workerfleet is a Plumego-based reference application for monitoring simulation workers running in Kubernetes. The target production shape is one Kubernetes pod per worker, around 8000 pods in one cluster, with each worker process able to run multiple tasks concurrently.

The service answers these operational questions:

- Is each worker online, degraded, offline, or unknown?
- Is the worker process alive and able to accept new tasks?
- Which tasks are currently running on each worker?
- What phase is each current task in?
- When was the worker last connected?
- Which pods or nodes show abnormal state distribution?
- Which alerts are firing or recently resolved?

The app is intentionally placed under `reference/workerfleet` as an app-local reference service. Workerfleet-specific storage, MongoDB dependencies, metrics, alerting, and Kubernetes logic must stay inside this submodule and must not expand Plumego stable roots.

## 2. Scope And Assumptions

In scope:

- Single Kubernetes cluster.
- One pod maps to one worker.
- One worker can run multiple tasks concurrently.
- Worker online state means the worker process is alive and can accept tasks, not just that it recently sent a heartbeat.
- Current tasks require task phase and phase name, but not progress percentage or ETA.
- Current state plus seven-day historical task, event, and alert records.
- Feishu and generic webhook notification channels.
- Prometheus metrics for later Grafana visualization.

Out of scope for this phase:

- Multi-cluster routing or aggregation.
- Per-task progress percentage and ETA.
- Distributed task scheduling.
- Strong exactly-once notification delivery.
- Adding workerfleet-specific APIs or dependencies to Plumego stable packages.

## 3. High-Level Architecture

```mermaid
flowchart LR
  Worker["Worker process"] -->|"register / heartbeat"| API["Workerfleet HTTP API"]
  Kube["Kubernetes API"] -->|"pod list/watch"| KubeSync["Inventory Sync"]
  API --> Ingest["Ingest Service"]
  KubeSync --> Reconcile["Pod Reconcile"]
  Ingest --> Domain["Domain Rules"]
  Reconcile --> Domain
  Domain --> Store["Store Interface"]
  Store --> Memory["Memory Backend"]
  Store --> Mongo["MongoDB Backend"]
  Store --> Query["Query API"]
  Store --> Alert["Alert Engine"]
  Alert --> Notify["Feishu / Webhook"]
  Domain --> Metrics["Prometheus Metrics"]
  Alert --> Metrics
  KubeSync --> Metrics
  Metrics --> Grafana["Grafana"]
```

Layer responsibilities:

- `main.go`: process entrypoint, environment config, route registration, HTTP server lifecycle, graceful shutdown.
- `internal/app`: application bootstrap, explicit dependency wiring, service methods, route registration.
- `internal/handler`: HTTP request and response layer.
- `internal/domain`: worker status rules, task reconciliation, pod reconciliation, alerts, domain events.
- `internal/platform/store`: app-local storage interfaces and shared query/filter types.
- `internal/platform/store/memory`: local in-memory backend.
- `internal/platform/store/mongo`: MongoDB persistence backend.
- `internal/platform/kube`: Kubernetes pod mapping, pod list/watch client, inventory sync.
- `internal/platform/metrics`: workerfleet Prometheus collector, exporter, and instrumentation observer.
- `internal/platform/notifier`: Feishu and generic webhook notification sinks.

## 4. Module Boundary Policy

Workerfleet is a standalone Go module:

- Module path: `workerfleet`
- Location: `reference/workerfleet`
- Plumego root dependency: `replace github.com/spcent/plumego => ../..`
- MongoDB dependencies: only in `reference/workerfleet/go.mod`

Boundary rules:

- Stable Plumego roots must not import workerfleet packages.
- Workerfleet domain must not depend on HTTP, MongoDB, Kubernetes, or Prometheus packages.
- Metrics are injected as explicit optional observers, not hidden globals.
- Storage dependencies are behind app-local interfaces.
- No workerfleet-specific labels, stores, or alert rules are added to Plumego stable packages.

## 5. Domain Model

Worker identity:

- `worker_id`
- `namespace`
- `pod_name`
- `pod_uid`
- `node_name`
- `container_name`
- `image`
- `version`

Worker runtime:

- `process_alive`
- `accepting_tasks`
- `last_seen_at`
- `last_ready_at`
- `last_heartbeat_at`
- `last_error`
- `restart_count`

Worker status:

- `online`
- `degraded`
- `offline`
- `unknown`

Online means the process is alive and ready to accept tasks. A worker can still be online while busy when it has active tasks but is not accepting additional work. A worker becomes degraded when signals are stale, it reports errors, it is not accepting tasks while idle, or task phases are stuck. A worker becomes offline when the process is not alive, the pod has failed, the pod has disappeared, or heartbeat expiry crosses policy thresholds.

Task phases:

- `unknown`
- `queued`
- `preparing`
- `running`
- `finalizing`
- `succeeded`
- `failed`
- `canceled`

Each worker heartbeat reports the full active-task set. The service treats `active_tasks` as a replacement snapshot, not a delta stream.

## 6. Worker Ingest Flow

```mermaid
sequenceDiagram
  participant W as Worker
  participant H as HTTP Handler
  participant I as Ingest Service
  participant D as Domain Rules
  participant S as Store
  participant M as Metrics

  W->>H: POST /v1/workers/register
  H->>I: RegisterCommand
  I->>S: GetWorkerSnapshot
  I->>D: apply status policy
  I->>S: UpsertWorkerSnapshot
  I->>S: AppendWorkerEvent
  I->>M: ObserveWorkerSnapshot

  W->>H: POST /v1/workers/heartbeat
  H->>I: WorkerReport
  I->>S: GetWorkerSnapshot
  I->>D: MergeWorkerReport
  D-->>I: snapshot + domain events
  I->>S: UpsertWorkerSnapshot
  I->>S: AppendWorkerEvent
  I->>S: AppendTaskHistory
  I->>M: ObserveWorkerSnapshot
```

Register flow:

- merges worker identity into the current snapshot.
- evaluates initial status.
- writes a `worker_registered` event when the worker is new.

Heartbeat flow:

- updates process liveness and readiness.
- replaces the active-task set.
- emits task start, phase change, finish, heartbeat, readiness, and status transition events.
- persists task history for completed tasks.
- updates Prometheus counters, gauges, and histograms through the metrics observer.

## 7. Kubernetes Inventory Sync

Kubernetes inventory sync maps pod state into worker snapshots. It is designed to reconcile platform truth with worker-reported truth.

Inputs:

- Kubernetes API host.
- Namespace.
- Label selector.
- Worker container name.
- Service account token or explicit bearer token.

Required RBAC:

- `get`, `list`, and `watch` on pods in the target namespace.

Pod mapping:

- pod name and UID become worker identity fields.
- namespace and node name come from pod metadata/spec.
- container image and restart count come from the selected worker container.
- pod phase maps to worker pod phase.

Reconciliation behavior:

- pod restart count increases emit pod restart events.
- pod disappearance marks pod deletion.
- failed or succeeded pods push worker status toward offline.
- pod metrics are exported as aggregate low-cardinality gauges.

Current implementation note:

- The `internal/platform/kube` sync primitives exist.
- The HTTP service entrypoint currently does not start a background Kubernetes sync loop. That loop should be added as a separate runtime card with explicit interval, error handling, and shutdown policy.

## 8. Storage Design

Store interfaces are app-local under `internal/platform/store`.

Current state collections:

- `worker_snapshots`: one current snapshot per worker.
- `worker_active_tasks`: one current task projection per active task.

History collections:

- `task_history`
- `worker_events`
- `alert_events`

Retention:

- task history: seven days.
- worker events: seven days.
- alert events: seven days.
- current worker snapshots and active-task projections are not TTL-pruned.

MongoDB behavior:

- startup validates Mongo URI and database before handlers are exposed.
- startup pings MongoDB and ensures indexes.
- history writes are append-only from the app perspective.
- generated duplicate history IDs are treated as idempotent retry success.
- `expire_at` drives TTL cleanup for history and alert collections.

Memory backend:

- default local backend.
- intended for local development, tests, and demos.
- not a production persistence option for 8000 workers.

## 9. HTTP API Design

Base path: `/v1`

Core endpoints:

- `POST /v1/workers/register`
- `POST /v1/workers/heartbeat`
- `GET /v1/workers`
- `GET /v1/workers/:worker_id`
- `GET /v1/tasks/:task_id`
- `GET /v1/fleet/summary`
- `GET /v1/alerts`
- `GET /metrics`

Response rules:

- success responses use the existing workerfleet handler envelope.
- errors use Plumego `contract.WriteError`.
- route wiring is explicit, one method plus one path plus one handler per registration line.

Query behavior:

- workers can be filtered by status, namespace, node, task type, and accepting-task state.
- alerts can be filtered by worker, alert type, and status.
- task lookup checks current active-task projection first, then latest task history.

## 10. Metrics And Grafana

Metrics are exported through `GET /metrics` in Prometheus text format.

Metric goals:

- fleet size by worker status.
- pod phase distribution.
- pod-level worker state, heartbeat age, and active case count.
- pod-level case throughput, including hourly success and failure counts.
- active case count by namespace, node, task type, and phase.
- active case count by node.
- task start and finish rates.
- task phase transition rates.
- phase and total task duration histograms.
- case total duration distribution by node and pod.
- case step duration distribution by node, pod, step, and controlled exec plan.
- worker status transition counters.
- alert emission counters and firing gauges.
- ingest and Kubernetes sync operation durations.

Default labels:

- `namespace`
- `node`
- `status`
- `phase`
- `task_type`
- `alert_type`
- `severity`
- `from_phase`
- `to_phase`
- `from_status`
- `to_status`
- `operation`
- `result`

Forbidden labels:

- `task_id`
- `case_id`
- `worker_id`
- `pod_name`
- `pod_uid`

The next case/step metrics phase intentionally allows `pod` on selected metrics because pod-level throughput and duration distribution are explicit business requirements. `exec_plan_id` is optional and should only be enabled when active plan cardinality is bounded.

Grafana dashboards should stay aggregate-first. Per-case and per-task drilldown should use workerfleet APIs and MongoDB history instead of high-cardinality Prometheus labels. The complete case/step metric plan is documented in [Case And Step Metrics Design](../case-step-metrics.md).

## 11. Alerting And Notification

Initial alert types:

- `worker_offline`
- `worker_degraded`
- `worker_not_accepting_tasks`
- `worker_no_heartbeat`
- `worker_stage_stuck`
- `pod_restart_burst`
- `pod_missing`
- `task_conflict`

Alert state:

- `firing`
- `resolved`

Dedupe:

- worker-scoped alerts use `alert_type:worker_id`.
- task conflict alerts use `alert_type:task_id`.

Notification channels:

- Feishu webhook.
- Generic JSON webhook.

Current implementation note:

- Alert evaluation and notifier primitives exist.
- The HTTP service entrypoint currently does not start a periodic alert evaluation and dispatch loop. That loop should be added separately with explicit interval, retry, timeout, and error reporting policy.

## 12. Runtime Configuration

HTTP:

- `WORKERFLEET_HTTP_ADDR`, default `:8080`
- `WORKERFLEET_SHUTDOWN_TIMEOUT`, default `10s`

Storage:

- `WORKERFLEET_STORE_BACKEND=memory|mongo`
- `WORKERFLEET_MONGO_URI`
- `WORKERFLEET_MONGO_DATABASE`
- `WORKERFLEET_MONGO_CONNECT_TIMEOUT`
- `WORKERFLEET_MONGO_OPERATION_TIMEOUT`
- `WORKERFLEET_MONGO_MAX_POOL_SIZE`
- `WORKERFLEET_RETENTION_DAYS`

Planned runtime loop configuration:

- Kubernetes sync enabled flag.
- Kubernetes sync interval.
- Kubernetes namespace and label selector.
- worker container name.
- alert evaluation interval.
- notifier delivery timeout.
- Feishu webhook URL.
- generic webhook URL and headers.

## 13. Capacity And Reliability Considerations

Target scale:

- around 8000 workers.
- multiple active tasks per worker.
- single cluster.

Key design choices for scale:

- heartbeats replace full active-task sets to avoid server-side partial merge ambiguity.
- current worker snapshots are stored separately from historical append-only records.
- active tasks have a reverse lookup projection for task detail queries.
- Prometheus labels avoid worker ID, task ID, case ID, and pod name.
- Mongo indexes are created at startup for current-state and lookup paths.

Operational risks:

- simultaneous heartbeat bursts can create write pressure.
- full active-task replacement requires workers to report complete state correctly.
- stale pod inventory can delay pod failure visibility until sync recovers.
- notification dispatch currently needs runtime loop and retry policy wiring.

Mitigations:

- keep HTTP handlers thin and persistence writes explicit.
- use MongoDB connection pooling and operation timeouts.
- keep worker status policy deterministic and test-covered.
- expose ingest and inventory sync duration metrics.
- use Grafana aggregate panels before drilling into APIs.

## 14. Security And Failure Policy

Security requirements:

- do not log webhook secrets, bearer tokens, private keys, or signatures.
- fail closed when required Mongo config is missing.
- use explicit service account or bearer token for Kubernetes API access.
- keep notification headers out of error messages.

Failure behavior:

- invalid startup config prevents handler exposure.
- nil metrics observers are safe and do not block business flow.
- notifier errors are returned to the dispatcher.
- storage errors propagate to HTTP handlers as structured errors.

## 15. Implementation Status

Implemented:

- workerfleet submodule with module path `workerfleet`.
- HTTP worker register, heartbeat, list, detail, task, fleet summary, and alert query routes.
- app bootstrap with memory and Mongo backends.
- Mongo schema, indexes, snapshot, active task, history, event, and alert persistence.
- domain worker status, active-task reconciliation, pod reconciliation, and alert rules.
- Feishu and generic webhook notifier primitives.
- Prometheus collector, exporter, instrumentation, and `/metrics` route.
- service entrypoint with HTTP server startup and graceful shutdown.
- Grafana dashboard planning documentation.

Not yet wired into the running process:

- periodic Kubernetes inventory sync loop.
- periodic alert evaluation loop.
- alert notification dispatch loop.
- runtime env config for Kubernetes and notifiers.

Recommended next cards:

- add Kubernetes inventory sync runtime loop.
- add alert evaluation and notification runtime loop.
- add service health/readiness endpoints.
- add operational config examples for Kubernetes deployment.
