# Card 1116

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/metrics.go
- x/data/sharding/metrics_test.go
- x/data/sharding/config/watcher.go
- x/data/sharding/config/watcher_test.go
- docs/modules/x-data/README.md
Depends On:
- 0749-x-data-rw-lifecycle-and-weight-api

Goal:
Remove misleading sharding metrics fields and make dynamic config watcher lifecycle idempotent.

Scope:
- Stop exposing uncomputed P50/P95/P99 values in snapshots and Prometheus output, or clearly zero them through documented omission.
- Make ConfigWatcher.Stop idempotent.
- Prevent Start from being called more than once concurrently.
- Add tests for Stop idempotency and metrics snapshot behavior.

Non-goals:
- Do not implement a full percentile reservoir.
- Do not change RouterMetrics in this card.
- Do not change config file schema.

Files:
- x/data/sharding/metrics.go
- x/data/sharding/metrics_test.go
- x/data/sharding/config/watcher.go
- x/data/sharding/config/watcher_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding/...
- go test -race -timeout 60s ./x/data/sharding/...
- go vet ./x/data/sharding/...

Docs Sync:
- Update x/data docs for metrics helper limits and watcher lifecycle.

Done Definition:
- Metrics output no longer implies percentile values are computed.
- Watcher Stop can be called repeatedly without panic.
- Repeated Start returns a clear error.

Outcome:
- Removed uncomputed latency percentile fields from `MetricsSnapshot`.
- Added JSON snapshot coverage proving percentile fields are not exposed.
- Made `ConfigWatcher.Stop` idempotent and repeated `Start` return a clear
  error.
- Documented metrics helper limits and watcher lifecycle semantics.

Validation:
- `go test -timeout 20s ./x/data/sharding/...`
- `go test -race -timeout 60s ./x/data/sharding/...`
- `go vet ./x/data/sharding/...`
