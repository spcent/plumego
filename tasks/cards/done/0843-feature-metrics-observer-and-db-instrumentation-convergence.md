# Card 0843

Priority: P1
State: done
Primary Module: metrics
Owned Files:
- `metrics/collector.go`
- `metrics/http_observer.go`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`
- `store/db/metrics.go`
- `store/kv`
- `x/pubsub`
- `x/mq`
- `x/ipc`
- `middleware/httpmetrics`
- `x/observability`

Goal:
- Finish shrinking stable observability contracts by removing feature-owned observer interfaces from stable `metrics`.
- Converge database instrumentation ownership so pool-stat polling and DB instrumentation wrappers live with observability owners instead of the stable store root.

Problem:
- Even after pruning stable metric type constants, `metrics/collector.go` still exports `PubSubObserver`, `MQObserver`, `KVObserver`, `IPCObserver`, and `DBObserver`, so stable `metrics` remains the registry for feature-specific observer contracts.
- `x/pubsub`, `x/mq`, `x/ipc`, and `store/kv` still import these stable observer interfaces instead of owning their own narrow metrics contracts or using a generic recorder path.
- `store/db/metrics.go` still exports `MetricsObserver`, `InstrumentedDB`, and `RecordPoolStats`, which keeps DB observability wrappers and pool-stat reporting in the stable persistence root even though store docs explicitly push DB analytics and monitoring behavior out of the stable layer.
- This leaves the stable roots with a half-pruned observability boundary: feature-specific type taxonomies were removed, but feature-specific observer interfaces and DB instrumentation wrappers still remain.

Scope:
- Decide the minimum shared observability contracts that truly belong in stable `metrics` after the taxonomy cleanup.
- Move feature-specific observer interfaces to the owning package, or replace them with owner-specific recorder usage where that is clearer.
- Remove or relocate stable `store/db` instrumentation wrappers and pool-stat reporting to an observability-owned package.
- Update all downstream call sites and tests in `store/kv`, `x/pubsub`, `x/mq`, `x/ipc`, `middleware/httpmetrics`, and `x/observability`.
- Sync metrics and store docs/manifests to the converged ownership boundary.

Non-goals:
- Do not reintroduce a repo-wide feature metric type catalog in stable `metrics`.
- Do not remove the canonical HTTP metrics path without replacing its callers in the same change.
- Do not leave compatibility aliases for removed feature observer interfaces.
- Do not move plain SQL connectivity helpers or non-observability DB utilities out of `store/db`.

Files:
- `metrics/collector.go`
- `metrics/http_observer.go`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`
- `store/db/metrics.go`
- `store/kv`
- `x/pubsub`
- `x/mq`
- `x/ipc`
- `middleware/httpmetrics`
- `x/observability`

Tests:
- `go test -timeout 20s ./metrics/... ./store/db ./store/kv ./x/pubsub ./x/mq ./x/ipc ./middleware/httpmetrics ./x/observability/...`
- `go test -race -timeout 60s ./metrics/... ./store/db ./store/kv ./x/pubsub ./x/mq ./x/ipc ./middleware/httpmetrics ./x/observability/...`
- `go vet ./metrics/... ./store/db ./store/kv ./x/pubsub ./x/mq ./x/ipc ./middleware/httpmetrics ./x/observability/...`

Docs Sync:
- Keep the metrics and store manifests/primers aligned on the rule that stable roots own only minimal shared contracts, while feature observer interfaces and DB instrumentation wrappers belong to their owning feature or observability package.

Done Definition:
- Stable `metrics` no longer exports feature-specific observer interfaces beyond the minimum shared contract that survives this convergence.
- Stable `store/db` no longer owns DB instrumentation wrappers or pool-stat metrics polling.
- Downstream packages compile against owner-specific or converged observability contracts with no residual references to removed stable observer exports.
- Metrics and store docs/manifests describe the same reduced observability ownership boundary the code implements.

Outcome:
- Completed.
- Removed `PubSubObserver`, `MQObserver`, `KVObserver`, `IPCObserver`, and `DBObserver` from stable `metrics`; feature packages now own their narrow observer contracts locally.
- Kept stable `metrics` focused on `Recorder`, `HTTPObserver`, and base collector composition, while downstream packages continue to use `BaseMetricsCollector` through owner-defined interfaces.
- Moved DB instrumentation wrappers and pool-stat polling from stable `store/db` into `x/observability/dbinsights`.
- Updated metrics/store manifests and primers to describe the converged ownership boundary, and migrated DB instrumentation tests to the observability owner package.
