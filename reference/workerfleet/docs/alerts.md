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
