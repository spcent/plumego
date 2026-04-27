# Card 0612

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: metrics
Owned Files:
- metrics/collector.go
- metrics/noop.go
- metrics/collector_test.go
- metrics/noop_test.go
- metrics/stats_contract_test.go
Depends On: 5102

Goal:
Normalize `CollectorStats` snapshots so callers get a consistent, defensive
payload from base and no-op collectors.

Scope:
- Ensure returned `NameBreakdown` maps are initialized and caller-mutable without
  affecting collector state.
- Keep zero-value and reset behavior predictable.
- Avoid duplicating empty stats construction between collectors.

Non-goals:
- Do not add new statistics fields.
- Do not retain raw metric records in stable collectors.
- Do not change `NoopCollector` from discard-only behavior.

Files:
- metrics/collector.go
- metrics/noop.go
- metrics/collector_test.go
- metrics/noop_test.go
- metrics/stats_contract_test.go

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- None unless behavior wording in module docs becomes inaccurate.

Done Definition:
- Base and no-op stats snapshots use the same initialized empty breakdown shape.
- Mutating returned stats maps cannot mutate collector internals.
- Targeted metrics tests and vet pass.

Outcome:
- Normalized empty stats construction through a shared helper.
- Made base zero-value stats and no-op stats return initialized,
  caller-owned `NameBreakdown` maps.
- Added regression coverage for zero-value base stats and no-op stats map
  ownership.
