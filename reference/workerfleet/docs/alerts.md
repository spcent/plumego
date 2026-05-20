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

- `WORKERFLEET_PROFILE=dev|prod` selects the app-local default policy bundle used for worker status and alert thresholds.
- `WORKERFLEET_STATUS_STALE_AFTER`, `WORKERFLEET_STATUS_OFFLINE_AFTER`, `WORKERFLEET_STATUS_STAGE_STUCK_AFTER`, and `WORKERFLEET_STATUS_RESTART_BURST_THRESHOLD` override worker status policy inputs.
- alert defaults derive from the effective status policy unless `WORKERFLEET_ALERT_STAGE_STUCK_AFTER` or `WORKERFLEET_ALERT_RESTART_BURST_THRESHOLD` explicitly override them.
- alert evaluation can be enabled with `WORKERFLEET_ALERT_EVALUATION_ENABLED=true`.
- `WORKERFLEET_ALERT_EVALUATION_INTERVAL` controls the evaluation loop interval.
- notification delivery can be enabled with `WORKERFLEET_NOTIFICATION_ENABLED=true`.
- `WORKERFLEET_NOTIFIER_DELIVERY_TIMEOUT` controls per-dispatch delivery timeout.
- emitted firing and resolved alert records are persisted before notification delivery is attempted.
- notification delivery errors do not crash the service.
- alert evaluation and notification delivery errors are observed through `workerfleet_runtime_errors_total` with low-cardinality `operation` and `error_class` labels.
- startup validation fails closed on contradictory or unsafe thresholds such as `offline_after <= stale_after` or stage-stuck values below the minimum supported floor.

Mongo-backed retention:

- `alert_events` stores firing and resolved alert records as append-only documents.
- each alert event carries `expire_at` so MongoDB TTL indexes can remove records after the seven-day retention window.
- alert dedupe and resolution still live in the domain alert engine; MongoDB only persists and filters alert records.

Deployment reference:

- `reference/workerfleet/deploy/prometheusrule.yaml` contains baseline alert rules built only on the stabilized workerfleet metric catalog.
