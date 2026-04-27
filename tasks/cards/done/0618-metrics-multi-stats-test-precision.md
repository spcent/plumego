# Card 0618

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: metrics
Owned Files:
- metrics/multi_test.go
- metrics/stats_contract_test.go
- metrics/TESTING.md
Depends On: 5203

Goal:
Make multi collector stats tests precise, accurately named, and aligned with the
implemented stats contract.

Scope:
- Replace loose aggregate assertions with exact expectations where behavior is
  deterministic.
- Rename stale duration-related test wording that now verifies counters,
  breakdowns, and start time.
- Add coverage that multi stats breakdown maps are caller-owned.

Non-goals:
- Do not change multi collector runtime semantics.
- Do not add duration aggregation to stable `metrics`.
- Do not introduce repo-wide metrics test helpers.

Files:
- metrics/multi_test.go
- metrics/stats_contract_test.go
- metrics/TESTING.md

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- Update testing guidance only if needed to reflect precise stats assertions.

Done Definition:
- Multi stats tests describe and assert implemented behavior exactly.
- Caller-owned map behavior is covered for multi stats.
- Targeted metrics tests and vet pass.

Outcome:
- Replaced loose multi stats assertions with exact totals, error counts, active
  series, and name breakdown expectations.
- Renamed the stale weighted-average test to describe its actual counter and
  breakdown coverage.
- Added caller-owned `NameBreakdown` coverage for multi stats and simplified the
  stats contract nil-map branch.
