# Card 0745

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
- 0744-x-data-sharding-sql-lexer-rewrite-boundary

Goal:
Make CrossShardAll execution semantics match its documented first-success behavior and avoid waiting on slower shards once a successful result is available.

Scope:
- Use a cancellable child context for cross-shard fan-out.
- Return the first successful Rows as soon as it is available.
- Drain late goroutine results in the background and close late successful rows.
- Add a test proving a slow shard does not delay first-success return.

Non-goals:
- Do not merge rows across shards.
- Do not add query planner APIs.
- Do not rename CrossShardAll in this card.

Files:
- x/data/sharding/router.go
- x/data/sharding/router_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Clarify CrossShardAll is first-success fan-out with cancellation, not result merging.

Done Definition:
- First successful shard can return before slower shards finish.
- Late successful result sets are closed.
- All-error behavior still returns ErrAllShardsFailed.

Outcome:
