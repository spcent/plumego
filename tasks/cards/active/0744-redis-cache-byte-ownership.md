# Card 0744

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P2
State: active
Primary Module: x/cache/redis
Owned Files:
- x/cache/redis/redis.go
- x/cache/redis/redis_test.go
Depends On:

Goal:
Make Redis cache adapter byte ownership match stable cache behavior.

Scope:
- Defensively clone bytes returned from Client.Get.
- Avoid exposing caller-owned bytes to mutable fake clients where practical.
- Add tests that mutate returned slices and assert cached state is unchanged.

Non-goals:
- Do not change Redis client interfaces unless necessary.
- Do not alter atomic Incr/Append capability behavior.

Files:
- x/cache/redis/redis.go
- x/cache/redis/redis_test.go

Tests:
- go test -timeout 20s ./x/cache/redis
- go test -timeout 20s ./store/cache

Docs Sync:
- None.

Done Definition:
- Mutating a Get result cannot mutate adapter/client-owned cached bytes.
- Targeted tests pass.

Outcome:

