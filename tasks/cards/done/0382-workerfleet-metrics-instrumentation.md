# Card 0382

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/platform/metrics
Owned Files:
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/domain/ingest_service.go`
- `reference/workerfleet/internal/domain/alert_engine.go`
- `reference/workerfleet/internal/platform/kube/discovery.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go`
Depends On:
- `tasks/cards/done/0381-workerfleet-prometheus-exporter.md`
Blocked By: —

Goal:
- Instrument workerfleet runtime paths so Prometheus captures worker status, pod phase, active case count, phase transition count, phase duration, inventory sync duration, and alert activity.
- Preserve existing domain rules while recording metrics as side-effect observers.

Scope:
- Add explicit optional metrics observer injection for ingest, Kubernetes inventory sync, and alert evaluation.
- Record worker and pod gauges from current snapshots and pod sync results.
- Record case counters and phase duration histograms from active-task reconciliation events.
- Record alert counters and firing gauges from alert engine output.

Non-goals:
- Do not make metrics required for core business flow.
- Do not introduce hidden global collectors.
- Do not use task ID, worker ID, or pod name as default Prometheus labels.
- Do not change Mongo persistence semantics.

Files:
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/domain/ingest_service.go`
- `reference/workerfleet/internal/domain/alert_engine.go`
- `reference/workerfleet/internal/platform/kube/discovery.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go`

Tests:
- `cd reference/workerfleet && go test ./internal/platform/metrics/...`
- `cd reference/workerfleet && go test ./internal/domain/...`
- `cd reference/workerfleet && go test ./internal/platform/kube/...`
- `cd reference/workerfleet && go test ./...`
- Tests for nil observer behavior, phase transition recording, active case gauges, alert metrics, and Kubernetes sync duration recording.

Docs Sync:
- Update `reference/workerfleet/docs/metrics.md` with exact instrumentation points and label behavior.

Done Definition:
- Metrics are recorded from worker heartbeat/reconcile, Kubernetes inventory sync, and alert evaluation paths.
- Nil metrics observers are safe and preserve current behavior.
- High-cardinality labels remain excluded by tests.

Outcome:
- Implemented optional observer injection for ingest, alert evaluation, and Kubernetes inventory sync.
- Added app-local workerfleet metrics instrumentation for worker status, accepting workers, active cases, task lifecycle counters, phase durations, pod phase sync gauges, alert counters, firing alert gauges, and operation durations.
- Kept high-cardinality IDs out of default labels.
