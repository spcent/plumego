# Card 0726: Store Cache Capability Split

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P0
State: active
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
- x/cache/distributed/node.go
- x/cache/distributed/distributed.go
- x/tenant/store/cache/tenant_cache.go
- x/cache/leaderboard/leaderboard.go
Depends On:
- 0725

Goal:
Split the stable cache base contract from optional counter and append capabilities so provider backends are not forced to implement operations they do not support.

Scope:
- Shrink `store/cache.Cache` to base get/set/delete/exists/clear operations.
- Add `CounterCache`, `AppenderCache`, and `AtomicCache` capability interfaces.
- Keep `MemoryCache`, `x/cache/redis.Adapter`, `x/cache/distributed.DistributedCache`, and tenant cache wrappers on the appropriate capability interfaces.
- Update downstream compile-time expectations and tests.

Non-goals:
- Do not remove `MemoryCache.Incr`, `MemoryCache.Decr`, or `MemoryCache.Append`.
- Do not change normal cache operation semantics.
- Do not add metrics or HTTP response-cache behavior.

Files:
- store/cache/cache.go
- store/cache/cache_test.go
- x/cache/distributed/node.go
- x/cache/distributed/distributed.go
- x/tenant/store/cache/tenant_cache.go
- x/cache/leaderboard/leaderboard.go

Tests:
- go test -timeout 20s ./store/cache ./x/cache/distributed ./x/cache/redis ./x/cache/leaderboard ./x/tenant/store/cache
- go test -race -timeout 60s ./store/cache ./x/cache/distributed ./x/cache/redis ./x/cache/leaderboard ./x/tenant/store/cache
- go vet ./store/cache ./x/cache/distributed ./x/cache/redis ./x/cache/leaderboard ./x/tenant/store/cache

Docs Sync:
- Required in comments and store module docs if capability names affect user-facing guidance.

Done Definition:
- Basic cache implementations only need the base `Cache` interface.
- Atomic/append callers depend on explicit capability interfaces.
- Focused tests and vet pass.

Outcome:
