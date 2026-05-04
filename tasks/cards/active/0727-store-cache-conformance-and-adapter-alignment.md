# Card 0727: Store Cache Conformance And Adapter Alignment

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P0
State: active
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
- x/cache/redis/redis.go
- x/cache/redis/redis_test.go
- x/cache/distributed/distributed_test.go
Depends On:
- 0726

Goal:
Add reusable cache conformance coverage and align first-party cache adapters with the stable cache semantics.

Scope:
- Add focused conformance tests for base cache and atomic capability behavior.
- Ensure Redis adapter integer encoding and overflow behavior match stable cache expectations.
- Ensure distributed cache reports unsupported atomic/append capability clearly if a node lacks it.

Non-goals:
- Do not add external Redis dependencies or networked integration tests.
- Do not change tenant scoping semantics.

Files:
- store/cache/cache.go
- store/cache/cache_test.go
- x/cache/redis/redis.go
- x/cache/redis/redis_test.go
- x/cache/distributed/distributed_test.go

Tests:
- go test -timeout 20s ./store/cache ./x/cache/redis ./x/cache/distributed
- go test -race -timeout 60s ./store/cache ./x/cache/redis ./x/cache/distributed
- go vet ./store/cache ./x/cache/redis ./x/cache/distributed

Docs Sync:
- Required if new capability errors or helper comments are added.

Done Definition:
- Cache semantics are protected by conformance-style tests.
- Redis adapter no longer diverges from stable integer counter encoding or overflow semantics.
- Focused tests and vet pass.

Outcome:
