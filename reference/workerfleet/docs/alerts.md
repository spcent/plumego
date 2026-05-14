# Workerfleet Alerts

The MVP alert engine evaluates persisted worker snapshots and emits alert records with dedupe keys.

Initial alert set:

- `worker_offline`
- `worker_degraded`
- `worker_not_accepting_tasks`
- `worker_no_heartbeat`
- `worker_stage_stuck`
- `pod_restart_burst`
- `pod_missing`
- `task_conflict`

Alert state model:

- `firing`
- `resolved`

Dedupe model:

- worker-scoped alerts use `alert_type:worker_id`
- task conflict alerts use `alert_type:task_id`

Runtime behavior:

- alert evaluation can be enabled with `WORKERFLEET_ALERT_EVALUATION_ENABLED=true`.
- `WORKERFLEET_ALERT_EVALUATION_INTERVAL` controls the evaluation loop interval.
- notification delivery can be enabled with `WORKERFLEET_NOTIFICATION_ENABLED=true`.
- `WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT` controls per-dispatch delivery timeout.
- emitted firing and resolved alert records are persisted before notification delivery is attempted.
- notification delivery errors do not crash the service.
- alert evaluation and notification delivery errors are observed through `workerfleet_runtime_errors_total` with low-cardinality `operation` and `error_class` labels.

Mongo-backed retention:

- `alert_events` stores firing and resolved alert records as append-only documents.
- each alert event carries `expire_at` so MongoDB TTL indexes can remove records after the seven-day retention window.
- alert dedupe and resolution still live in the domain alert engine; MongoDB only persists and filters alert records.
