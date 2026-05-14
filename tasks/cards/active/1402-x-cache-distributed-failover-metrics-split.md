# Card 1402

Milestone: v1-cleanup-phase-4
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/cache/distributed
Owned Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/failover.go
- x/cache/distributed/metrics.go
- x/cache/distributed/distributed_test.go
Depends On:
- 1401

Goal:
- Split distributed cache failover and metrics helpers into focused files without changing cache behavior.

Scope:
- Move failover attempt selection, retry behavior, and fallback helpers into `failover.go`.
- Move private metrics update/snapshot helpers into `metrics.go`.
- Preserve metrics counters, health-derived counts, failover retry defaults, and error return behavior.

Non-goals:
- Do not change public `DistributedMetrics` fields.
- Do not change health checker or hash ring implementation.
- Do not add exporter integrations.
- Do not change cache operation contracts.

Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/failover.go
- x/cache/distributed/metrics.go
- x/cache/distributed/distributed_test.go

Tests:
- go test -timeout 30s ./x/cache/distributed
- go vet ./x/cache/distributed
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless package comments about failover or metrics move.

Done Definition:
- Failover and metrics logic have separate file ownership.
- Existing distributed cache tests pass unchanged.
- No public API, metrics field, or failover behavior changes.

Outcome:

