# Card 0836

Priority: P1
State: active
Primary Module: metrics
Owned Files:
- `metrics/collector.go`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`
- `x/observability/prometheus.go`
- `x/messaging`

Goal:
- Shrink stable `metrics` to a generic record/observer contract by removing feature-specific metric taxonomy ownership from the stable root.
- Converge metric naming and type taxonomy onto owning packages instead of keeping a repo-wide feature registry in `metrics`.

Problem:
- `metrics/module.yaml` and `docs/modules/metrics/README.md` describe a small stable contracts package, but `metrics/collector.go` still exports a large feature-specific `MetricType` catalog for HTTP, PubSub, MQ, KV, IPC, and DB operations.
- Stable `BaseMetricsCollector` also owns operation-to-type fallback logic (`ObservePubSub`, `ObserveMQ`, `ObserveKV`, `ObserveIPC`, `ObserveDB`) that effectively turns `metrics` into a central feature taxonomy registry.
- This blurs the boundary between a generic collector contract and feature-specific metrics semantics that should live with the owning stable module or extension.
- Downstream packages such as `x/observability/prometheus` still key behavior on stable feature constants like `MetricHTTPRequest`, which reinforces the repo-wide taxonomy coupling.

Scope:
- Decide the minimum stable `metrics` taxonomy that truly belongs in the shared collector contract.
- Remove or relocate feature-specific metric type constants and naming/fallback logic that exceed that boundary.
- Move feature-owned metric taxonomy to the owning package or make stable collectors work off generic record fields without a central registry.
- Update downstream collectors, exporters, and tests in the same change.
- Sync metrics docs and manifest to the reduced generic-contract boundary.

Non-goals:
- Do not remove `MetricRecord`, `Recorder`, `AggregateCollector`, or `MultiCollector` unless required by the convergence.
- Do not reintroduce repo-wide metrics test helpers or rolling-window aggregation into stable `metrics`.
- Do not add feature-specific exporters to stable `metrics`.
- Do not preserve compatibility aliases for relocated feature taxonomy constants.

Files:
- `metrics/collector.go`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`
- `x/observability/prometheus.go`
- `x/messaging`

Tests:
- `go test -timeout 20s ./metrics/... ./x/observability/... ./x/messaging/... ./middleware/httpmetrics`
- `go test -race -timeout 60s ./metrics/... ./x/observability/... ./x/messaging/... ./middleware/httpmetrics`
- `go vet ./metrics/... ./x/observability/... ./x/messaging/... ./middleware/httpmetrics`

Docs Sync:
- Keep the metrics manifest and primer aligned on the rule that stable `metrics` owns generic collector contracts, not a repo-wide feature metric taxonomy catalog.

Done Definition:
- Stable `metrics` no longer owns a broad feature-specific metric type registry beyond the minimum shared contract.
- Feature metric taxonomy lives with the owning stable module or extension, or stable collectors operate generically without a central feature registry.
- Downstream exporters and tests compile against the converged taxonomy surface with no residual references to deleted stable metrics constants.
- Metrics docs and manifest describe the same reduced stable boundary the code implements.

Outcome:
- Completed.
- Removed broad feature-specific metric taxonomy ownership from stable `metrics`, leaving only the shared HTTP metric type in the stable root.
- Feature taxonomy now lives with the owning package or is expressed through generic metric names and labels instead of a central stable registry.
