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
- `error_class`

`pod`, `exec_plan_id`, and `step` are experimental labels. They are allowed
only on selected drilldown-oriented business metrics and are disabled by
default in `prod`. Keep case-level and task-level identifiers in MongoDB/API
drilldown instead of Prometheus labels.

Forbidden default labels:

- `task_id`
- `case_id`
- `worker_id`
- `pod_name`
- `pod_uid`
- `raw_error`
- `error_message`

Implemented metric catalog:

Stable metric catalog:

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
- `workerfleet_alerts_firing`
- `workerfleet_case_phase_duration_seconds`
- `workerfleet_case_total_duration_seconds`
- `workerfleet_worker_report_apply_duration_seconds`
- `workerfleet_kube_inventory_sync_duration_seconds`
- `workerfleet_runtime_errors_total`

Experimental metric catalog:

- `workerfleet_worker_active_cases`
- `workerfleet_worker_heartbeat_age_seconds`
- `workerfleet_case_completed_total`
- `workerfleet_case_failed_total`
- `workerfleet_case_duration_seconds`
- `workerfleet_case_step_completed_total`
- `workerfleet_case_step_duration_seconds`
- `workerfleet_case_step_stuck_cases`
- `workerfleet_case_step_oldest_active_age_seconds`

Experimental metric gating:

- `WORKERFLEET_EXPERIMENTAL_METRICS_ENABLED` controls whether experimental pod, exec-plan, case, and step series are emitted at all.
- `dev` profile enables experimental metrics by default for local analysis.
- `prod` profile disables experimental metrics by default so pod, `exec_plan_id`, and step-heavy series do not become part of the default fleet scrape surface.
- stable metric families remain enabled regardless of the experimental flag.

State and inventory coverage:

- pod status is represented by `workerfleet_pods{phase,namespace,node}`.
- worker status is represented by `workerfleet_workers{status,namespace,node}`.
- pod-level worker heartbeat freshness is represented by the experimental `workerfleet_worker_heartbeat_age_seconds{namespace,node,pod,status}` series when experimental metrics are enabled.

Scrape endpoint:

- `GET /metrics` exposes Prometheus text format when the workerfleet metrics handler is wired into routes.
- the workerfleet exporter is app-local and independent from Plumego stable `metrics`.
- Grafana panel guidance and PromQL examples live in [Grafana Dashboard Plan](./grafana.md).
- `reference/workerfleet/deploy/servicemonitor.yaml` provides a reference Prometheus Operator scrape manifest for `/metrics`.
- Full pod/worker/exec-plan/case/step metric design lives in [Case And Step Metrics Design](./case-step-metrics.md).

Instrumentation points:

- worker register and heartbeat paths accept an optional observer and split instrumentation into two sources:
- worker snapshots drive state gauges such as worker status, accepting-task state, active-case gauges, and, when experimental metrics are enabled, pod-level heartbeat age, stuck-case gauges, and oldest active step age.
- worker domain events drive counters and histograms such as task starts, finishes, phase transitions, case completion/failure totals, total case duration, and case step completion/duration.
- Kubernetes inventory sync accepts an optional observer and records pod phase gauges plus sync duration histograms with `operation` and `result`; list/watch relist recovery stays behind the same low-cardinality operation labels.
- alert evaluation accepts an optional observer and records emitted alert counters plus firing alert gauges.
- runtime loops report Kubernetes sync, status sweep, alert evaluation, and notification delivery errors through `workerfleet_runtime_errors_total{operation,error_class}`.
- notification delivery uses bounded notifier classes such as `http_4xx`, `http_429`, `http_5xx`, `network`, and `configuration`; raw status text, URLs, response bodies, and secret values stay out of metric labels.
- nil observers are safe and leave business behavior unchanged.
- stable aggregate gauges are labeled only by approved low-cardinality labels. Worker IDs, task IDs, case IDs, pod names, pod UIDs, and raw error messages stay out of Prometheus labels.

Case and step metrics:

- `pod` is required for optional pod-level throughput and duration distribution panels.
- `exec_plan_id` is a controlled optional label and should be disabled if active plan cardinality is high.
- `case_id` and `task_id` stay out of Prometheus and belong in MongoDB/API drilldown.
- step duration distribution should use histogram metrics rather than per-case gauges.
- metrics carrying `pod`, `exec_plan_id`, or `step` remain experimental until the label cardinality and panel usage are proven stable in production.
