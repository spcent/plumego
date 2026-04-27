# Card 5303

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: metrics
Owned Files:
- metrics/collector.go
- metrics/multi.go
- metrics/multi_test.go
- metrics/stats_contract_test.go
- metrics/TESTING.md
Depends On: 5302

Goal:
Normalize `ActiveSeries` calculation consistently for base and multi collector
stats, including child collectors that return name breakdowns without active
series counts.

Scope:
- Add one internal active-series normalization helper.
- Use it in base stats snapshots and multi stats aggregation.
- Add regression coverage for mixed child stats where some children omit
  `ActiveSeries`.

Non-goals:
- Do not add rolling-window aggregation.
- Do not make multi stats deduplicate series across child collectors.
- Do not change `CollectorStats` fields.

Files:
- metrics/collector.go
- metrics/multi.go
- metrics/multi_test.go
- metrics/stats_contract_test.go
- metrics/TESTING.md

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- Clarify that multi collectors sum child-maintained active series and normalize
  omitted active counts from breakdown size.

Done Definition:
- Base and multi stats use one active-series normalization rule.
- Mixed child stats no longer undercount active series.
- Targeted metrics tests and vet pass.

Outcome:
- Added one internal active-series normalization helper shared by base and multi
  stats paths.
- Fixed multi stats aggregation for mixed child collectors where some children
  omit `ActiveSeries` but provide `NameBreakdown`.
- Added regression coverage and synced testing guidance.
