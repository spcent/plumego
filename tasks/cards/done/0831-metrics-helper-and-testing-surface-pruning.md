# Card 0831

Priority: P1
State: active
Primary Module: metrics
Owned Files:
- `metrics/adapter.go`
- `metrics/helpers.go`
- `metrics/testing.go`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`
Depends On:
- `0824-metrics-stable-boundary-pruning.md`

Goal:
- Align stable `metrics` with its documented “contracts + base collectors” boundary by pruning window-aggregation and repo-wide test-helper ownership from the stable root.
- Converge to one small stable runtime metrics surface instead of mixing runtime contracts with testkit and analytics helper APIs.

Problem:
- `metrics/module.yaml` and `docs/modules/metrics/README.md` now describe a small stable contracts package, but the code still exports `Aggregator`, `AggregatorStats`, `Timer`, the `Measure*Func` helper family, and `MockCollector` call-capture test utilities.
- `MockCollector` is already consumed by tests in other stable and extension modules, which means the stable metrics root is acting as a repo-wide testkit package in addition to its runtime contract role.
- `Aggregator` and the helper families extend the stable public surface beyond the “base collectors and minimal support contracts” boundary that the current module metadata promises.

Scope:
- Decide the minimum stable helper surface that truly belongs with collector contracts and base collectors.
- Remove or relocate rolling-window aggregation APIs that are not part of the canonical stable collector surface.
- Remove or relocate repo-wide testing/mocking utilities out of the stable runtime root.
- Update downstream tests and examples to the relocated helper/test surfaces in the same change.
- Sync the metrics module manifest and primer to the reduced stable public boundary.

Non-goals:
- Do not redesign `Recorder`, observer interfaces, `BaseMetricsCollector`, `NoopCollector`, or `MultiCollector` unless required to complete the pruning.
- Do not reintroduce Prometheus, tracing, devtools, or feature-specific ownership into stable `metrics`.
- Do not preserve compatibility wrappers in `metrics` for relocated helper or testing APIs.
- Do not move runtime metrics behavior into `core`.

Files:
- `metrics/adapter.go`
- `metrics/helpers.go`
- `metrics/testing.go`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`

Tests:
- `go test -timeout 20s ./metrics/... ./store/db/... ./x/messaging/...`
- `go test -race -timeout 60s ./metrics/... ./store/db/... ./x/messaging/...`
- `go vet ./metrics/... ./store/db/... ./x/messaging/...`

Docs Sync:
- Keep the metrics primer and manifest aligned on the rule that stable `metrics` owns small runtime contracts and base collectors, not repo-wide testkit APIs or extra analytics surfaces.

Done Definition:
- Stable `metrics` no longer owns helper or testing APIs that exceed the documented contracts/base-collectors boundary.
- Any relocated testing or aggregation surfaces live in a clearly owned non-stable package.
- Downstream tests and examples use the converged helper/test surfaces with no residual references to deleted metrics APIs.
- Metrics docs and manifest describe the same reduced stable surface the code implements.

Outcome:
- Completed.
- Moved rolling-window aggregation helpers to `x/observability/windowmetrics` and repo-wide metrics testing helpers to `x/observability/testmetrics`.
- Stable `metrics` now keeps only contracts, base collectors, and aggregate composition.
