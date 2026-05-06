# Card 0778

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/router.go
- x/data/sharding/router_test.go
- x/data/sharding/types.go
- docs/modules/x-data/README.md
Depends On:
- 0777-x-data-sharding-unresolved-first-policy

Goal:
Make cross-shard fan-out behavior explicit and avoid result leaks on invalid shard plans.

Scope:
- Prevalidate all resolved shard indexes before starting fan-out goroutines.
- Replace the misleading CrossShardAll public policy name in repo code with a first-success fan-out name while x/data is still experimental.
- Update tests and docs for the renamed policy.

Non-goals:
- Do not implement merged cross-shard rows.
- Do not add query planner dependencies.
- Do not change sharding strategy math.

Files:
- x/data/sharding/router.go
- x/data/sharding/router_test.go
- x/data/sharding/types.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding/...
- go test -race -timeout 60s ./x/data/sharding/...
- go vet ./x/data/sharding/...

Docs Sync:
- Update x/data docs with the first-success fan-out policy name and semantics.

Done Definition:
- Invalid resolved shard indexes fail before fan-out starts.
- Repo code no longer uses the misleading CrossShardAll symbol.
- Tests and docs cover the behavior.

Outcome:
- Replaced the exported CrossShardAll symbol in Go code with
  CrossShardFirstSuccess while x/data remains experimental.
- Changed the string/config policy name from `all` to `first_success`.
- Prevalidated all resolved shard indexes before launching fan-out goroutines.
- Updated router/config tests, docs, and module manifest.

Validation:
- `rg -n --glob '*.go' 'CrossShardAll' .` returned no matches.
- `go test -timeout 20s ./x/data/sharding/...`
- `go test -race -timeout 60s ./x/data/sharding/...`
- `go vet ./x/data/sharding/...`
