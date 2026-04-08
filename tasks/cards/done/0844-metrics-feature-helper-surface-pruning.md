# Card 0844

Priority: P1
State: active
Primary Module: metrics
Owned Files:
- `metrics/collector.go`
- `metrics/multi.go`
- `metrics/noop.go`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`
- `x/observability`
- `x/devtools`
- `x/pubsub`
- `x/mq`
- `x/ipc`
- `store/kv`

Goal:
- Finish shrinking stable `metrics` so it owns only generic recording, the shared HTTP observer, and small collector composition.
- Remove feature-specific `Observe*` helper methods that still keep messaging, KV, IPC, and DB metric semantics in the stable root.

Problem:
- `metrics/collector.go` still defines `AggregateCollector` as a repo-wide “everything collector” with `ObservePubSub`, `ObserveMQ`, `ObserveKV`, `ObserveIPC`, and `ObserveDB`, even though the manifest and docs now say feature-specific metrics ownership belongs to the owning package.
- `BaseMetricsCollector`, `NoopCollector`, `MultiCollector`, and the internal base forwarder still implement those feature helpers, so stable `metrics` continues to own feature naming, label conventions, and helper semantics for non-HTTP domains.
- `x/observability/prometheus`, `x/devtools/devcollector`, and `x/observability/testmetrics` still build on these stable helper methods, which means the feature-specific surface was removed from interfaces in `0843` but not from the stable collector implementations themselves.
- This leaves the stable boundary half-pruned: feature packages now own their narrow observer interfaces, but stable `metrics` still ships the convenience helpers and naming rules those packages should own.

Scope:
- Converge the stable collector contracts on generic `Record`, the shared `HTTPObserver`, stats, and reset semantics only.
- Remove non-HTTP feature helper methods from stable `AggregateCollector`, `BaseMetricsCollector`, `NoopCollector`, and `MultiCollector`.
- Move feature-specific helper construction and forwarding to the owning package or to `x/observability` helper adapters where shared extension-side reuse is still justified.
- Update downstream call sites and tests in `x/observability`, `x/devtools`, `x/pubsub`, `x/mq`, `x/ipc`, and `store/kv`.
- Sync metrics docs and manifest to the final reduced collector surface.

Non-goals:
- Do not remove the canonical `Record` path.
- Do not remove the shared stable HTTP observer contract.
- Do not reintroduce a repo-wide feature metric type catalog or compatibility wrapper layer.
- Do not redesign feature metrics payloads beyond what is required to remove stable helper ownership.

Files:
- `metrics/collector.go`
- `metrics/multi.go`
- `metrics/noop.go`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`
- `x/observability`
- `x/devtools`
- `x/pubsub`
- `x/mq`
- `x/ipc`
- `store/kv`

Tests:
- `go test -timeout 20s ./metrics/... ./x/observability/... ./x/devtools ./x/pubsub ./x/mq ./x/ipc ./store/kv`
- `go test -race -timeout 60s ./metrics/... ./x/observability/... ./x/devtools ./x/pubsub ./x/mq ./x/ipc ./store/kv`
- `go vet ./metrics/... ./x/observability/... ./x/devtools ./x/pubsub ./x/mq ./x/ipc ./store/kv`

Docs Sync:
- Keep the metrics manifest and primer aligned on the rule that stable `metrics` owns only generic recording plus the shared HTTP observer, while non-HTTP feature helper semantics belong to their owning package or extension-side observability helpers.

Done Definition:
- Stable `metrics` no longer exports non-HTTP feature helper methods or contracts.
- `x/observability`, `x/devtools`, and feature packages compile against generic `Record` or owner-local helper layers with no residual references to removed stable `ObservePubSub` / `ObserveMQ` / `ObserveKV` / `ObserveIPC` / `ObserveDB` methods.
- Stable metrics docs and manifest describe the same reduced collector surface the code implements.

Outcome:
- Completed.
- Removed `ObservePubSub`, `ObserveMQ`, `ObserveKV`, `ObserveIPC`, and `ObserveDB` from stable `AggregateCollector`, `BaseMetricsCollector`, `NoopCollector`, and `MultiCollector`.
- Added `x/observability/featuremetrics` as the extension-owned home for non-HTTP metric record builders and recorder helpers.
- Updated `x/devtools` to use extension-owned helpers for DB/MQ/KV/IPC/PubSub recording, and reduced `x/observability/PrometheusCollector` to the stable generic + HTTP collector contracts.
- Migrated stable helper behavior tests out of `metrics` into the extension-owned helper package and updated downstream tests to stop relying on removed stable helper methods.
