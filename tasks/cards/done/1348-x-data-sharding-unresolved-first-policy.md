# Card 1348

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/router.go
- x/data/sharding/router_test.go
- docs/modules/x-data/README.md
Depends On:
- 0776-x-data-sharding-sqlite-config-validation

Goal:
Prevent CrossShardFirst from returning shard-0 partial results for unresolved SQL.

Scope:
- Reject unresolved QueryContext routing when the policy cannot resolve a shard key.
- Keep resolved multi-shard CrossShardFirst behavior as first resolved shard.
- Add regression coverage for unresolved queries.

Non-goals:
- Do not add cross-shard result merging.
- Do not change QueryRowContext behavior beyond matching the existing fail-closed contract.
- Do not change sharding strategies.

Files:
- x/data/sharding/router.go
- x/data/sharding/router_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Update x/data docs with fail-closed unresolved CrossShardFirst behavior.

Done Definition:
- Unresolved QueryContext under CrossShardFirst returns a routing error instead of querying shard 0.
- Resolved CrossShardFirst still queries the first resolved shard.
- Tests and docs cover the behavior.

Outcome:
- Made unresolved QueryContext use an explicit DefaultShardIndex when present.
- Made unresolved CrossShardFirst return ErrCrossShardQuery instead of querying
  shard 0.
- Kept resolved multi-shard CrossShardFirst behavior on the first resolved shard.
- Updated router tests and x/data routing docs.

Validation:
- `go test -timeout 20s ./x/data/sharding`
- `go test -race -timeout 60s ./x/data/sharding`
- `go vet ./x/data/sharding`
