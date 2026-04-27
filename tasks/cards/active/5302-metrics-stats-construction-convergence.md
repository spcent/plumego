# Card 5302

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: metrics
Owned Files:
- metrics/collector.go
- metrics/noop.go
- metrics/collector_test.go
- metrics/noop_test.go
- docs/modules/metrics/README.md
Depends On: 5301

Goal:
Converge collector stats initialization so empty and started stats are built
through one internal path.

Scope:
- Add a single internal helper for initialized `CollectorStats` construction.
- Use the helper from base constructor, base clear, and no-op stats.
- Keep zero start time for no-op/empty stats and non-zero start time for base
  collectors.

Non-goals:
- Do not add new `CollectorStats` fields.
- Do not change public APIs.
- Do not widen stable metrics into aggregation windows or exporters.

Files:
- metrics/collector.go
- metrics/noop.go
- metrics/collector_test.go
- metrics/noop_test.go
- docs/modules/metrics/README.md

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- Keep stats snapshot wording accurate if helper semantics affect wording.

Done Definition:
- Base and no-op stats construction share one initialized stats helper.
- Tests keep base start time and no-op zero start time behavior locked.
- Targeted metrics tests and vet pass.

Outcome:
