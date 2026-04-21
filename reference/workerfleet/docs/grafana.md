# Workerfleet Grafana Dashboard Plan

This dashboard assumes Prometheus scrapes `GET /metrics` from the workerfleet service and stores seven days of workerfleet metrics. The workerfleet metrics are app-local and intentionally avoid worker IDs, task IDs, case IDs, and pod names as default labels.

Recommended template variables:

- `$namespace`: `label_values(workerfleet_workers, namespace)`
- `$node`: `label_values(workerfleet_workers{namespace="$namespace"}, node)`
- `$task_type`: `label_values(workerfleet_active_cases, task_type)`
- `$phase`: `label_values(workerfleet_active_cases, phase)`

Scrape assumptions:

```yaml
scrape_configs:
  - job_name: workerfleet
    metrics_path: /metrics
    static_configs:
      - targets:
          - workerfleet.default.svc.cluster.local:8080
```

Use Kubernetes service discovery instead of `static_configs` when workerfleet is deployed behind changing service endpoints. Keep the scrape target at service level, not pod level, because workerfleet exports fleet aggregates.

## Fleet Overview

Panels:

- Worker status distribution: total workers by `status`.
- Non-online workers by namespace and node.
- Workers accepting tasks.
- Active cases by task type and phase.
- Current pod phases.

PromQL:

```promql
sum by (status) (workerfleet_workers)
```

```promql
sum by (namespace, node, status) (workerfleet_workers{status!="online"})
```

```promql
sum by (namespace, node) (workerfleet_worker_accepting_tasks)
```

```promql
sum by (namespace, task_type, phase) (workerfleet_active_cases)
```

```promql
sum by (namespace, node, phase) (workerfleet_pods)
```

Recommended alerting:

- fire when `sum(workerfleet_workers{status="offline"}) > 0` for 5 minutes.
- fire when `sum(workerfleet_worker_accepting_tasks) == 0` for 2 minutes.
- warn when degraded workers exceed an agreed percentage of the fleet.

## Node Capacity And Heatmap

Panels:

- Active case count by node, task type, and phase.
- Non-online worker count by node.
- Pod phase count by node.

PromQL:

```promql
sum by (node, task_type, phase) (workerfleet_node_active_cases)
```

```promql
sum by (node, status) (workerfleet_workers{status!="online"})
```

```promql
sum by (node, phase) (workerfleet_pods)
```

Use a heatmap or table panel with `node` as the row key and phase/status columns for quick skew detection. Avoid adding pod names or worker IDs to the panel query; jump from aggregate panels to the workerfleet API for per-worker detail.

## Case Throughput

Panels:

- Case start rate by namespace and task type.
- Case finish rate by namespace, task type, and final status.
- Phase transition rate.

PromQL:

```promql
sum by (namespace, task_type) (rate(workerfleet_case_started_total[5m]))
```

```promql
sum by (namespace, task_type, status) (rate(workerfleet_case_finished_total[5m]))
```

```promql
sum by (namespace, task_type, from_phase, to_phase) (rate(workerfleet_case_phase_transitions_total[5m]))
```

Recommended alerting:

- warn when case finish rate drops to zero while active cases remain above zero.
- warn when failed or canceled finish rate crosses the task-type baseline.

Example:

```promql
sum(workerfleet_active_cases) > 0
and
sum(rate(workerfleet_case_finished_total[10m])) == 0
```

## Phase Latency

Panels:

- Phase duration p95 by task type and phase.
- Phase duration p99 by task type and phase.
- Total case duration p95 by task type and final status.

PromQL:

```promql
histogram_quantile(
  0.95,
  sum by (le, task_type, phase) (
    rate(workerfleet_case_phase_duration_seconds_bucket[5m])
  )
)
```

```promql
histogram_quantile(
  0.99,
  sum by (le, task_type, phase) (
    rate(workerfleet_case_phase_duration_seconds_bucket[5m])
  )
)
```

```promql
histogram_quantile(
  0.95,
  sum by (le, task_type, status) (
    rate(workerfleet_case_total_duration_seconds_bucket[5m])
  )
)
```

Recommended alerting:

- warn when phase p95 exceeds the stage timeout used by workerfleet status policy.
- fire when a task phase latency panel shows sustained p99 growth together with degraded workers.

## Runtime And Inventory Health

Panels:

- Worker report apply duration p95.
- Kubernetes inventory sync duration p95.
- Kubernetes inventory sync error count.

PromQL:

```promql
histogram_quantile(
  0.95,
  sum by (le, operation) (
    rate(workerfleet_worker_report_apply_duration_seconds_bucket[5m])
  )
)
```

```promql
histogram_quantile(
  0.95,
  sum by (le, operation, result) (
    rate(workerfleet_kube_inventory_sync_duration_seconds_bucket[5m])
  )
)
```

```promql
sum by (operation) (
  rate(workerfleet_kube_inventory_sync_duration_seconds_count{result="error"}[5m])
)
```

Recommended alerting:

- warn when inventory sync errors occur for more than 5 minutes.
- warn when inventory sync p95 grows close to the scrape interval or sync interval.

## Alerts

Panels:

- Emitted alert records by type, severity, and status.
- Current firing alerts by type and severity.

PromQL:

```promql
sum by (alert_type, severity, status) (rate(workerfleet_alerts_total[5m]))
```

```promql
sum by (alert_type, severity) (workerfleet_alerts_firing)
```

Recommended alerting:

- route workerfleet firing alerts to Feishu and the configured webhook receiver.
- alert on persistent firing alert gauges instead of only alert creation rate.

## Cardinality Rules

Keep default Grafana panels at aggregate level:

- allowed labels: `namespace`, `node`, `status`, `phase`, `task_type`, `alert_type`, `severity`, `from_phase`, `to_phase`, `from_status`, `to_status`, `operation`, `result`
- forbidden labels: `task_id`, `case_id`, `worker_id`, `pod_name`
- use workerfleet query APIs for per-worker or per-task drilldown.

The service target is one cluster for this phase. Multi-cluster dashboards should add an external Prometheus label such as `cluster` at scrape or remote-write time instead of changing workerfleet metric label defaults.
