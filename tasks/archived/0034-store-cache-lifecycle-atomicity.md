# Card 0034

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
Depends On:

Goal:
Make `store/cache.MemoryCache` lifecycle and mutation behavior deterministic without widening the stable cache API.

Scope:
- Make `Close` idempotent so a second close cannot panic.
- Serialize cache mutations that currently combine `sync.Map` operations with separate size and memory counters.
- Keep compound operations (`Incr`, `Decr`, `Append`) atomic against other memory-cache writes.
- Add focused tests for repeated close and concurrent write/accounting safety.

Non-goals:
- Do not add provider-specific cache behavior.
- Do not move atomic operations out of the stable interface in this card.
- Do not add metrics, introspection, tenant scoping, or HTTP response caching.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Not required unless the public cache contract changes.

Done Definition:
- `Close` is safe to call repeatedly.
- Cache size/memory accounting cannot drift from concurrent write/delete/clear/expiry paths.
- Compound cache write operations remain covered by tests.

Outcome:
- Made `MemoryCache.Close` idempotent through `sync.Once`.
- Serialized memory-cache mutation paths and shared atomic-operation writes through one write lock.
- Fixed replacement accounting for existing zero-length values.
- Replaced sleep-based concurrent cache test synchronization with `sync.WaitGroup`.

Validation:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache
