# Card 0380

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/platform/metrics
Owned Files:
- `reference/workerfleet/internal/platform/metrics/model.go`
- `reference/workerfleet/internal/platform/metrics/model_test.go`
- `reference/workerfleet/docs/metrics.md`
Depends On:
- `tasks/cards/done/0379-workerfleet-mongo-config-and-wiring.md`
Blocked By:
- workerfleet metrics design review approval

Goal:
- Define the workerfleet business metrics model for Prometheus/Grafana without changing the stable `metrics` package or `x/observability`.
- Establish metric names, label cardinality rules, and app-local record types for pod, node, worker, case, phase, and alert observability.

Scope:
- Add app-local workerfleet metrics constants and helper types under `reference/workerfleet/internal/platform/metrics`.
- Define allowed labels and forbid high-cardinality labels such as `task_id`, `worker_id`, and `pod_name` in default exported series.
- Document the initial metric catalog and Prometheus naming conventions.

Non-goals:
- Do not implement `/metrics`.
- Do not instrument heartbeat, Kubernetes sync, or alert evaluation yet.
- Do not modify the root `metrics` stable module.
- Do not add Prometheus client dependencies yet.

Files:
- `reference/workerfleet/internal/platform/metrics/model.go`
- `reference/workerfleet/internal/platform/metrics/model_test.go`
- `reference/workerfleet/docs/metrics.md`

Tests:
- `cd reference/workerfleet && go test ./internal/platform/metrics/...`
- Tests for metric name stability and label allowlist validation.

Docs Sync:
- Add `reference/workerfleet/docs/metrics.md` describing metric names, labels, and cardinality policy.
- Link the metrics document from `reference/workerfleet/README.md` if the final implementation exposes it to operators.

Done Definition:
- The workerfleet metric catalog is defined in one app-local package.
- High-cardinality labels are explicitly rejected by tests or helper validation.
- The model is ready for a Prometheus exporter without changing domain, handler, or stable metrics packages.

Outcome:
- Added app-local workerfleet metrics model under `reference/workerfleet/internal/platform/metrics`.
- Defined the initial Prometheus metric catalog, metric kinds, allowed labels, and high-cardinality label rejection.
- Added `reference/workerfleet/docs/metrics.md` and linked it from the workerfleet README.
- Validation run:
  - `cd reference/workerfleet && go test ./internal/platform/metrics/...`
  - `cd reference/workerfleet && go test ./...`
