# Card 0623

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: metrics
Owned Files:
- metrics/multi_test.go
- metrics/TESTING.md
Depends On: 5303

Goal:
Complete fan-out constructor coverage for nil filtering, single-target
passthrough, multi-target fan-out, and no-op internal empty cases.

Scope:
- Add explicit `NewMultiCollector(...)` coverage for nil filtering and
  single-target passthrough.
- Add empty internal fan-out method coverage for collector and HTTP observer
  values.
- Keep constructor behavior unchanged.

Non-goals:
- Do not change fan-out runtime behavior unless tests expose a defect.
- Do not add async fan-out or recovery behavior.
- Do not add new public APIs.

Files:
- metrics/multi_test.go
- metrics/TESTING.md

Tests:
- go test -timeout 20s ./metrics/...
- go test -race -timeout 60s ./metrics/...
- go vet ./metrics/...

Docs Sync:
- Update testing guide only if constructor behavior wording needs precision.

Done Definition:
- Constructor behavior is covered for zero, one, and multiple non-nil targets.
- Empty internal fan-out values are covered as no-ops.
- Targeted metrics tests and vet pass.

Outcome:
- Added explicit `NewMultiCollector(...)` nil-filtering and single-target
  passthrough coverage.
- Added empty internal `MultiCollector` and `multiHTTPObserver` no-op coverage.
- Synced testing guide wording for empty internal fan-out values.
