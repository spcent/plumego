# Card 2025

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: reference/workerfleet/internal/platform/metrics
Owned Files:
- `reference/workerfleet/internal/platform/metrics/model.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go`
- `reference/workerfleet/docs/metrics.md`
- `reference/workerfleet/docs/grafana.md`
Depends On:
- `tasks/cards/active/2023-workerfleet-case-step-domain-and-heartbeat-contract.md`
Blocked By:
- case step domain and heartbeat contract

Goal:
- Implement the pod/worker/case/step Prometheus metrics needed for system state, pod throughput, and node/pod duration distribution.
- Prioritize histogram-based duration distribution for case and step runtime analysis.

Scope:
- Add metric specs for pod status, worker status, heartbeat age, active cases, case completion, case failure, case duration, step completion, step duration, stuck cases, and oldest active step age.
- Add observer logic that derives counters, gauges, and histograms from snapshot diffs.
- Keep `case_id`, `task_id`, `worker_id`, `pod_uid`, and raw error messages out of labels.
- Add tests for label validation and duration histogram recording.

Non-goals:
- Do not add worker heartbeat API fields; that belongs to Card 2023.
- Do not add Grafana JSON dashboards.
- Do not add Mongo case step history.

Files:
- `reference/workerfleet/internal/platform/metrics/model.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go`
- `reference/workerfleet/docs/metrics.md`
- `reference/workerfleet/docs/grafana.md`

Tests:
- `cd reference/workerfleet && go test ./internal/platform/metrics/...`
- `cd reference/workerfleet && go test ./internal/domain/...`
- `cd reference/workerfleet && go test ./...`

Docs Sync:
- Update metrics and Grafana docs with implemented metric names and PromQL.

Done Definition:
- Prometheus exposes pod-level throughput and case/step duration histograms.
- Grafana PromQL examples reference implemented metrics only.
- Forbidden high-cardinality labels are rejected by tests.

