# Card 0611

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: metrics
Owned Files:
- metrics/collector.go
- metrics/collector_test.go
- x/observability/recordbuffer/collector.go
- x/observability/recordbuffer/collector_test.go
- docs/modules/metrics/README.md
Depends On: 5101

Goal:
Remove duplicate HTTP metric record construction and make the stable HTTP record
shape explicit.

Scope:
- Add one canonical stable helper for constructing HTTP `MetricRecord` values.
- Use the helper inside `BaseMetricsCollector.ObserveHTTP(...)`.
- Migrate `x/observability/recordbuffer` to the stable helper instead of
  duplicating the HTTP record name, duration unit, and label keys.
- Keep bytes handling collector-owned; do not encode bytes as a stable label.

Non-goals:
- Do not add buffered record inspection back to stable `metrics`.
- Do not add HTTP transport policy or exporter wiring to stable `metrics`.
- Do not change Prometheus exposition behavior.

Files:
- metrics/collector.go
- metrics/collector_test.go
- x/observability/recordbuffer/collector.go
- x/observability/recordbuffer/collector_test.go
- docs/modules/metrics/README.md

Tests:
- go test -timeout 20s ./metrics/... ./x/observability/recordbuffer/...
- go test -race -timeout 60s ./metrics/... ./x/observability/recordbuffer/...
- go vet ./metrics/... ./x/observability/recordbuffer/...

Docs Sync:
- Document the stable helper and the intentionally omitted bytes label in
  `docs/modules/metrics/README.md`.

Done Definition:
- HTTP record construction has one stable code path.
- Recordbuffer tests prove the helper-backed record shape remains compatible.
- Targeted tests and vet pass.

Outcome:
- Added `metrics.NewHTTPRecord(...)` as the canonical stable HTTP metric record
  constructor.
- Reused that helper in `BaseMetricsCollector.ObserveHTTP(...)` and
  `x/observability/recordbuffer`.
- Added tests for the stable HTTP record shape and recordbuffer compatibility.
