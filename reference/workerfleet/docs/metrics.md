# Workerfleet Metrics

Workerfleet business metrics are app-local. They must not expand Plumego's stable `metrics` package or `x/observability` with workerfleet-specific labels.

Default Prometheus labels are intentionally low cardinality:

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

Forbidden default labels:

- `task_id`
- `case_id`
- `worker_id`
- `pod_name`

Initial metric catalog:

- `workerfleet_workers`
- `workerfleet_pods`
- `workerfleet_active_cases`
- `workerfleet_worker_accepting_tasks`
- `workerfleet_node_active_cases`
- `workerfleet_case_started_total`
- `workerfleet_case_finished_total`
- `workerfleet_case_phase_transitions_total`
- `workerfleet_worker_status_transitions_total`
- `workerfleet_alerts_total`
- `workerfleet_case_phase_duration_seconds`
- `workerfleet_case_total_duration_seconds`
- `workerfleet_worker_report_apply_duration_seconds`
- `workerfleet_kube_inventory_sync_duration_seconds`

Scrape endpoint:

- `GET /metrics` exposes Prometheus text format when the workerfleet metrics handler is wired into routes.
- the workerfleet exporter is app-local and independent from Plumego stable `metrics`.
