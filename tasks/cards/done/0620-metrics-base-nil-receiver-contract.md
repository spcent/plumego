# Card 0620

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
Align `BaseMetricsCollector` nil-receiver behavior with the defensive no-op
behavior already used by no-op and fan-out collectors.

Scope:
- Make nil `*BaseMetricsCollector` `Record`, `ObserveHTTP`, and `Clear` calls
  no-op.
- Make nil `*BaseMetricsCollector` `GetStats` return empty initialized stats.
- Add regression coverage for nil receiver behavior.

Non-goals:
- Do not change zero-value non-nil collector behavior.
- Do not add raw record retention or exporter behavior.
- Do not change public interfaces.

Files:
- metrics/collector.go
- metrics/collector_test.go
- docs/modules/metrics/README.md

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- Document nil collector calls as safe no-ops for base/fan-out collectors.

Done Definition:
- Nil base collector calls do not panic.
- Nil base stats return initialized empty `NameBreakdown`.
- Targeted metrics tests and vet pass.

Outcome:
- Made nil `*BaseMetricsCollector` `Record`, `ObserveHTTP`, and `Clear` calls
  no-op.
- Made nil `*BaseMetricsCollector.GetStats()` return initialized empty stats.
- Added regression coverage and synced metrics module README nil-safety wording.
