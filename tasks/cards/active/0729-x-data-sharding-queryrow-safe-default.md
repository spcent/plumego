# Card 0729

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/router.go
- x/data/sharding/router_test.go
Depends On:

Goal:
Make sharding QueryRowContext follow the same safe-default routing semantics as QueryContext.

Scope:
- Stop silently falling back to shard 0 when resolution or shard validation fails.
- Return a *sql.Row that surfaces the routing error on Scan, matching database/sql QueryRowContext expectations.
- Keep DefaultShardIndex behavior only for the explicit configured fallback path.
- Add tests for resolution failure, invalid shard index, and rewrite failure behavior.

Non-goals:
- Do not change QueryContext cross-shard policy semantics.
- Do not add result merging.
- Do not redesign Router public constructors.

Files:
- x/data/sharding/router.go
- x/data/sharding/router_test.go

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Update docs/modules/x-data/README.md if the documented QueryRow behavior changes.

Done Definition:
- QueryRowContext does not silently route unresolved sharded queries to shard 0.
- Routing errors are observable by callers through Scan.
- sharding normal and race tests pass.

Outcome:
