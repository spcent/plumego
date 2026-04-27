# Card 0381

Milestone: —
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/platform/metrics
Owned Files:
- `reference/workerfleet/internal/platform/metrics/collector.go`
- `reference/workerfleet/internal/platform/metrics/prometheus.go`
- `reference/workerfleet/internal/platform/metrics/prometheus_test.go`
- `reference/workerfleet/internal/app/routes.go`
- `reference/workerfleet/docs/metrics.md`
Depends On:
- `tasks/cards/done/0380-workerfleet-metrics-model.md`
Blocked By:
- workerfleet metrics design review approval

Goal:
- Implement a workerfleet-local Prometheus collector and exporter for business metrics and expose it through an app route.
- Keep the exporter app-local so workerfleet-specific labels and series do not pollute Plumego stable metrics contracts.

Scope:
- Implement counters, gauges, and histogram-style bucket output needed for workerfleet business metrics.
- Add an HTTP handler that emits Prometheus text format.
- Wire `GET /metrics` in the workerfleet route registration path.

Non-goals:
- Do not instrument workerfleet business flows yet.
- Do not change `x/observability/prometheus.go`.
- Do not add Grafana dashboards.

Files:
- `reference/workerfleet/internal/platform/metrics/collector.go`
- `reference/workerfleet/internal/platform/metrics/prometheus.go`
- `reference/workerfleet/internal/platform/metrics/prometheus_test.go`
- `reference/workerfleet/internal/app/routes.go`
- `reference/workerfleet/docs/metrics.md`

Tests:
- `cd reference/workerfleet && go test ./internal/platform/metrics/...`
- `cd reference/workerfleet && go test ./internal/app/...`
- Tests for Prometheus text output, label escaping, histogram bucket output, and `/metrics` route registration.

Docs Sync:
- Update `reference/workerfleet/docs/metrics.md` with scrape endpoint, sample output, and exporter behavior.

Done Definition:
- Workerfleet can expose Prometheus-compatible business metrics through `GET /metrics`.
- The exporter is app-local and does not require stable `metrics` or `x/observability` changes.
- Output is deterministic enough for focused tests and ready for instrumentation.

Outcome:
- Added an app-local workerfleet Prometheus collector and text exporter.
- Added gauge, counter, and histogram recording primitives with low-cardinality label validation.
- Added optional `GET /metrics` route wiring and bootstrap metrics collector creation.
- Updated metrics docs with the scrape endpoint behavior.
- Validation run:
  - `cd reference/workerfleet && go test ./internal/platform/metrics/...`
  - `cd reference/workerfleet && go test ./internal/app/...`
  - `cd reference/workerfleet && go test ./...`
