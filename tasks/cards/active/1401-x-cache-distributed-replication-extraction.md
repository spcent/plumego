# Card 1401

Milestone: v1-cleanup-phase-4
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/cache/distributed
Owned Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/replication.go
- x/cache/distributed/distributed_test.go
- x/cache/distributed/distributed_bench_test.go
Depends On:
- 1400

Goal:
- Extract distributed cache replication scheduling and execution from the main implementation file.

Scope:
- Move async and sync replication helpers, replication job handling, and drop handling into `replication.go`.
- Preserve replication modes, queue bounds, concurrency limits, timeout behavior, and drop callback semantics.
- Keep public config fields and exported types unchanged.

Non-goals:
- Do not change hash ring behavior.
- Do not change failover strategy behavior.
- Do not alter benchmark intent beyond import or helper name adjustments.
- Do not add new replication modes.

Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/replication.go
- x/cache/distributed/distributed_test.go
- x/cache/distributed/distributed_bench_test.go

Tests:
- go test -timeout 30s ./x/cache/distributed
- go vet ./x/cache/distributed
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless comments around replication modes are moved incorrectly.

Done Definition:
- Replication implementation is isolated in `replication.go`.
- Existing replication tests and benchmarks still compile.
- No public API or behavior change is introduced.

Outcome:

