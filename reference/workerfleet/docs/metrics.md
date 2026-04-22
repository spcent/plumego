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
- `pod`
- `exec_plan_id`
- `step`
- `error_class`

`pod` is allowed only on selected pod-level business metrics. `exec_plan_id` is safe only when active exec plans are bounded; otherwise keep it in MongoDB/API drilldown instead of default Prometheus labels.

Forbidden default labels:

- `task_id`
- `case_id`
- `worker_id`
- `pod_name`
- `pod_uid`
- `raw_error`
- `error_message`

Implemented metric catalog:

- `workerfleet_workers`
- `workerfleet_pods`
- `workerfleet_active_cases`
- `workerfleet_worker_active_cases`
- `workerfleet_worker_accepting_tasks`
- `workerfleet_worker_heartbeat_age_seconds`
- `workerfleet_node_active_cases`
- `workerfleet_case_started_total`
- `workerfleet_case_finished_total`
- `workerfleet_case_completed_total`
- `workerfleet_case_failed_total`
- `workerfleet_case_phase_transitions_total`
- `workerfleet_worker_status_transitions_total`
- `workerfleet_alerts_total`
- `workerfleet_alerts_firing`
- `workerfleet_case_phase_duration_seconds`
- `workerfleet_case_total_duration_seconds`
- `workerfleet_case_duration_seconds`
- `workerfleet_case_step_completed_total`
- `workerfleet_case_step_duration_seconds`
- `workerfleet_case_step_stuck_cases`
- `workerfleet_case_step_oldest_active_age_seconds`
- `workerfleet_worker_report_apply_duration_seconds`
- `workerfleet_kube_inventory_sync_duration_seconds`

State and inventory coverage:

- pod status is represented by `workerfleet_pods{phase,namespace,node}`.
- worker status is represented by `workerfleet_workers{status,namespace,node}`.
- worker heartbeat freshness is represented by `workerfleet_worker_heartbeat_age_seconds{namespace,node,pod,status}`.

Scrape endpoint:

- `GET /metrics` exposes Prometheus text format when the workerfleet metrics handler is wired into routes.
- the workerfleet exporter is app-local and independent from Plumego stable `metrics`.
- Grafana panel guidance and PromQL examples live in [Grafana Dashboard Plan](./grafana.md).
- Full pod/worker/exec-plan/case/step metric design lives in [Case And Step Metrics Design](./case-step-metrics.md).

Instrumentation points:

- worker register and heartbeat paths accept an optional observer and record worker status gauges, heartbeat-age gauges, accepting-task gauges, active-case gauges, task lifecycle counters, phase transition counters, pod-level case completion/failure counters, phase duration histograms, total task duration histograms, case duration histograms, case step completion counters, case step duration histograms, stuck-case gauges, oldest active step age gauges, and worker report apply duration.
- Kubernetes inventory sync accepts an optional observer and records pod phase gauges plus sync duration histograms with `operation` and `result`.
- alert evaluation accepts an optional observer and records emitted alert counters plus firing alert gauges.
- nil observers are safe and leave business behavior unchanged.
- aggregate gauges are labeled only by approved low-cardinality labels. Worker IDs, task IDs, case IDs, pod names, pod UIDs, and raw error messages stay out of Prometheus labels.

Case and step metrics:

- `pod` is required for pod-level throughput and duration distribution panels.
- `exec_plan_id` is a controlled optional label and should be disabled if active plan cardinality is high.
- `case_id` and `task_id` stay out of Prometheus and belong in MongoDB/API drilldown.
- step duration distribution should use histogram metrics rather than per-case gauges.
