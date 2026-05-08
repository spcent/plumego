# Card 0764

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/cluster.go
- x/data/sharding/cluster_test.go
- x/data/sharding/module.yaml
- docs/modules/x-data/README.md
Depends On:
- 0763-x-data-kvengine-compressed-snapshot-fail-closed

Goal:
Align sharding cluster convenience configuration with fail-closed stable defaults.

Scope:
- Reject invalid negative `DefaultShardIndex` values below `-1`.
- Make read fallback to primary explicit instead of silently enabled by default.
- Document `ClusterDB/New` ownership and stable-surface decision.

Non-goals:
- Do not remove `ClusterDB` in this card.
- Do not change router SQL parsing.
- Do not change rw cluster internals.

Files:
- x/data/sharding/cluster.go
- x/data/sharding/cluster_test.go
- x/data/sharding/module.yaml
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Update x/data docs with cluster ownership and fallback defaults.

Done Definition:
- `DefaultShardIndex < -1` is rejected.
- Default shard config does not silently enable fallback-to-primary.
- Tests and docs cover the safer defaults.

Outcome:
- Rejected `DefaultShardIndex` values below the `-1` disabled sentinel.
- Changed `DefaultShardConfig` so fallback-to-primary is disabled by default.
- Documented `ClusterDB/New` as the retained convenience wrapper that owns and
  closes configured shard DB handles.

Validation:
- `go test -timeout 20s ./x/data/sharding`
- `go test -race -timeout 60s ./x/data/sharding`
- `go vet ./x/data/sharding`
