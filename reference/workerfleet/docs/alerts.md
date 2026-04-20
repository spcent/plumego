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

Current implementation emits alert records only. Notification delivery is handled by later cards.

Mongo-backed retention:

- `alert_events` stores firing and resolved alert records as append-only documents.
- each alert event carries `expire_at` so MongoDB TTL indexes can remove records after the seven-day retention window.
- alert dedupe and resolution still live in the domain alert engine; MongoDB only persists and filters alert records.
