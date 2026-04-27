# Card 5101

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: metrics
Owned Files:
- metrics/collector.go
- metrics/collector_test.go
- metrics/stats_contract_test.go
- docs/modules/metrics/README.md
Depends On:

Goal:
Align stable base collector error statistics with HTTP observer behavior by
classifying HTTP status errors consistently.

Scope:
- Treat stable HTTP records with 4xx/5xx status labels as error records.
- Keep explicit `MetricRecord.Error` classification unchanged.
- Cover both `ObserveHTTP(...)` and generic `Record(...)` calls using the HTTP
  metric name and labels.

Non-goals:
- Do not add Prometheus/exporter behavior to stable `metrics`.
- Do not change the `AggregateCollector` or `HTTPObserver` method signatures.
- Do not introduce feature-specific metric names or labels.

Files:
- metrics/collector.go
- metrics/collector_test.go
- metrics/stats_contract_test.go
- docs/modules/metrics/README.md

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- Document the implemented HTTP status error classification in
  `docs/modules/metrics/README.md`.

Done Definition:
- Base collector reports `ErrorRecords` for explicit record errors and HTTP
  status codes >= 400.
- Existing collectors keep their public interfaces.
- Targeted metrics tests and vet pass.

Outcome:
- Implemented base HTTP error classification for `status >= 400` across
  `ObserveHTTP(...)` and generic HTTP `Record(...)` calls.
- Added regression coverage for valid HTTP error statuses, invalid status labels,
  and aggregate stats contract expectations.
- Updated metrics module docs with the implemented error classification rule.
