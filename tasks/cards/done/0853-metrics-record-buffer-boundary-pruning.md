# Card 0853

Priority: P1
State: active
Primary Module: metrics
Owned Files:
- `metrics/collector.go`
- `metrics/collector_test.go`
- `metrics/example_test.go`
- `metrics/doc.go`
- `x/observability`
- `x/messaging`
- `x/devtools`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`

Goal:
- Remove record-buffer inspection and retention tuning from stable `metrics`.
- Keep stable collectors focused on recording, HTTP observation, stats, and reset semantics.

Problem:
- `metrics.BaseMetricsCollector` still exposes `WithMaxRecords(...)` and `GetRecords(...)`, which makes the stable layer own record buffering and direct record inspection rather than just collector contracts.
- `x/messaging`, `x/devtools`, and observability tests depend on these helpers, so extensions are still leaning on a stable in-memory record store instead of an owning observability package.
- This keeps stable `metrics` wider than its manifest and README claim.

Scope:
- Remove stable record-buffer inspection/tuning APIs from `metrics.BaseMetricsCollector`.
- Move buffered record inspection ownership to an owning `x/observability` package.
- Update `x/messaging`, `x/devtools`, and related tests in the same change.
- Sync metrics docs and manifest to the reduced collector boundary.

Non-goals:
- Do not redesign generic metric recording or HTTP observation.
- Do not reintroduce feature-specific helpers into stable `metrics`.
- Do not change Prometheus/text exposition formats except as required by the new owner-side record source.

Files:
- `metrics/collector.go`
- `metrics/collector_test.go`
- `metrics/example_test.go`
- `metrics/doc.go`
- `x/observability`
- `x/messaging`
- `x/devtools`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`

Tests:
- `go test -timeout 20s ./metrics/... ./x/observability/... ./x/messaging ./x/devtools`
- `go test -race -timeout 60s ./metrics/... ./x/observability/... ./x/messaging ./x/devtools`
- `go vet ./metrics/... ./x/observability/... ./x/messaging ./x/devtools`

Docs Sync:
- Keep the metrics docs aligned on the rule that stable `metrics` owns contracts and base collectors, while record-buffer inspection/export support lives in `x/observability`.

Done Definition:
- Stable `metrics` no longer exports record-buffer inspection or retention-tuning APIs.
- Extensions consume an owning `x/observability` record source instead of `metrics.BaseMetricsCollector.GetRecords()`.
- Metrics docs and manifest describe the same reduced stable surface implemented in code.

Outcome:
- Completed.
- Removed `WithMaxRecords(...)` and `GetRecords(...)` from the exported stable `metrics.BaseMetricsCollector` surface, keeping record buffering package-local inside stable tests only.
- Added `x/observability/recordbuffer` as the owning buffered-record collector for observability, messaging exporters, and devtools collectors that need record inspection.
- Updated metrics docs/manifests and extension callers so stable `metrics` now owns only recording, shared HTTP observation, stats, and reset semantics.
