# Card 0734: Redis Cache Atomic Capabilities

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P2
State: active
Primary Module: x/cache
Owned Files:
- x/cache/redis/redis.go
- x/cache/redis/redis_test.go
- x/cache/distributed/distributed_test.go
Depends On:
- 0733

Goal:
Make Redis optional cache capability claims match the actual atomic behavior.

Scope:
- Avoid advertising `CounterCache` or `AppenderCache` unless the client supports atomic commands.
- Add optional Redis client interfaces for atomic increment and append behavior.
- Keep base `cache.Cache` behavior unchanged.
- Add tests for unsupported and supported optional capabilities.

Non-goals:
- Do not add a concrete Redis dependency.
- Do not change stable `store/cache.Cache`.
- Do not implement HTTP response caching.

Files:
- x/cache/redis/redis.go
- x/cache/redis/redis_test.go
- x/cache/distributed/distributed_test.go

Tests:
- go test -timeout 20s ./x/cache/redis ./x/cache/distributed ./store/cache
- go test -race -timeout 60s ./x/cache/redis ./x/cache/distributed ./store/cache
- go vet ./x/cache/redis ./x/cache/distributed ./store/cache

Docs Sync:
- Not required unless public docs mention Redis atomic capabilities.

Done Definition:
- Redis adapter no longer implements optional atomic capabilities through read-modify-write.
- Distributed wrappers return `ErrCapabilityUnsupported` for Redis clients lacking atomic support.
- Targeted tests, race tests, and vet pass.

Outcome:
