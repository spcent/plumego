# Card 0770

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/cache
Owned Files: x/cache/redis/redis.go, x/cache/redis/redis_test.go
Depends On:

Goal:

Make Redis adapter invalid-key validation wrap cache.ErrInvalidKey so callers can classify errors consistently with MemoryCache.

Scope:

- Wrap empty-key and control-character validation errors with cache.ErrInvalidKey.
- Preserve human-readable validation detail.
- Add regression coverage using errors.Is.

Non-goals:

- Changing Redis client interfaces.
- Changing valid key policy.
- Changing distributed or tenant cache behavior.

Files:

- x/cache/redis/redis.go
- x/cache/redis/redis_test.go

Tests:

- go test -race -timeout 60s ./x/cache/redis
- go test -timeout 20s ./x/cache/redis
- go vet ./x/cache/redis

Docs Sync:

- Not required; this aligns adapter errors with the stable cache contract.

Done Definition:

- Redis invalid key errors satisfy errors.Is(err, cache.ErrInvalidKey).
- Existing valid key behavior remains unchanged.
- Package tests and vet pass.

Outcome:

