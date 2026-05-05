# Card 0765

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
- 0764-x-data-sharding-cluster-config-safe-defaults

Goal:
Prevent unresolved single-row sharding queries from returning plausible partial data.

Scope:
- Make unresolved `QueryRowContext` fail closed unless an explicit default shard is configured.
- Clarify `CrossShardFirst` and `CrossShardAll` behavior in tests and docs.
- Preserve `QueryContext` first-success behavior where already documented.

Non-goals:
- Do not implement multi-shard row merging.
- Do not rename exported cross-shard constants in this card.
- Do not introduce a third-party SQL parser.

Files:
- x/data/sharding/router.go
- x/data/sharding/router_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Update x/data docs to state that unresolved `QueryRowContext` requires an explicit default shard.

Done Definition:
- Unresolved `QueryRowContext` no longer falls back to shard 0 by policy alone.
- Cross-shard first-success behavior is covered by focused tests.
- Docs describe the non-merged `CrossShardAll` contract.

