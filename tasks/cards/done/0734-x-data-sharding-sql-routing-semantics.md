# Card 0734

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/router.go
- x/data/sharding/rewriter.go
- x/data/sharding/router_test.go
- x/data/sharding/rewriter_test.go
- docs/modules/x-data/README.md
Depends On:
- 0733-x-data-idempotency-store-correctness

Goal:
Make sharding SQL routing and rewriting honest about supported semantics.

Scope:
- Route multi-shard resolutions through the cross-shard policy instead of forcing single-shard `Resolve`.
- Avoid claiming `CrossShardAll` merges rows when it returns the first successful result.
- Make table rewriting fail closed on SQL shapes that the local parser cannot safely rewrite.
- Add tests for IN/range multi-shard routing and unsafe rewrite refusal.

Non-goals:
- Do not add a third-party SQL parser.
- Do not implement result merging.
- Do not change strategy hashing/range/list algorithms.

Files:
- x/data/sharding/router.go
- x/data/sharding/rewriter.go
- x/data/sharding/router_test.go
- x/data/sharding/rewriter_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Update `docs/modules/x-data/README.md` for CrossShardAll and rewrite support limits.

Done Definition:
- IN/range queries no longer bypass cross-shard policy through a single-shard resolver error.
- Rewriter refuses unsafe SQL forms instead of broad string replacement.
- Documentation describes implemented behavior only.

Outcome:
- Routed `IN` and range queries through `ResolveMultiple` so multi-shard reads honor the configured cross-shard policy.
- Made `CrossShardFirst` use the first resolved shard instead of blindly using shard 0 for resolvable multi-shard queries.
- Rewrote per shard during fan-out instead of sending logical table names unchanged.
- Made SQL rewriting fail closed for nested `SELECT` and `UNION` statements.
- Updated x/data docs for multi-shard routing, rewrite limits, and hash strategy wording.

Validation:
- GOCACHE=/private/tmp/plumego-go-build go test -timeout 20s ./x/data/sharding
- GOCACHE=/private/tmp/plumego-go-build go test -race -timeout 60s ./x/data/sharding
- GOCACHE=/private/tmp/plumego-go-build go vet ./x/data/sharding
