# Card 5201

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: metrics
Owned Files:
- metrics/collector.go
- metrics/collector_test.go
- docs/modules/metrics/README.md
Depends On:

Goal:
Make the base collector record path match its aggregate-only responsibility and
make stable HTTP records complete at construction time.

Scope:
- Move HTTP record timestamp initialization into `NewHTTPRecord(...)`.
- Remove dead timestamp and label normalization from `BaseMetricsCollector.Record(...)`
  because the base collector does not retain raw records.
- Keep error classification and name breakdown behavior unchanged.

Non-goals:
- Do not add raw record buffering back to stable `metrics`.
- Do not change public collector interfaces.
- Do not move exporter or transport policy into stable `metrics`.

Files:
- metrics/collector.go
- metrics/collector_test.go
- docs/modules/metrics/README.md

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- Clarify that the base collector aggregates only and does not normalize or
  retain generic records.

Done Definition:
- `NewHTTPRecord(...)` returns a timestamped record.
- `BaseMetricsCollector.Record(...)` updates aggregate stats without dead record
  mutation.
- Targeted metrics tests and vet pass.

Outcome:
- `NewHTTPRecord(...)` now returns a timestamped stable HTTP record.
- Removed dead generic record timestamp and label normalization from
  `BaseMetricsCollector.Record(...)`; the base collector remains aggregate-only.
- Added regression coverage for timestamped HTTP records and aggregate-only
  generic record handling.
