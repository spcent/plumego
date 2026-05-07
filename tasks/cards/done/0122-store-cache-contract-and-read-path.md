# Card 0122

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
Depends On:
- 0121

Goal:
Make the memory cache counter/value contract clear while recovering scalable read behavior.

Scope:
- Make `Incr` and `Decr` accept canonical textual integer values while preserving existing internal encoded integer values.
- Clarify memory limit comments as payload accounting.
- Avoid serializing non-expired `Get` and `Exists` calls behind the write mutex.
- Add tests for text integer counters, encoded compatibility, and read-after-expiry cleanup behavior.

Non-goals:
- Do not change the `Cache` interface.
- Do not add external cache backends.
- Do not add process memory accounting.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Public comments only.

Done Definition:
- `Set(ctx, key, []byte("1"), ttl)` followed by `Incr(ctx, key, 1)` returns `2`.
- Existing gob-encoded counter values remain readable.
- Non-expired reads avoid the global write mutex.

Outcome:
- Changed memory-cache counters to store canonical textual int64 values while keeping legacy gob-encoded int64 values readable.
- Added focused tests for text integer increment, legacy gob counter compatibility, and textual counter storage.
- Clarified `MaxMemoryUsage` comments as tracked payload-byte accounting rather than process memory accounting.
- Made non-expired `Get` and `Exists` avoid the global write mutex; expired values still use the locked compare-and-delete path.

Validation:
- go test -timeout 20s ./store/cache
- go vet ./store/cache
- go test -race -timeout 60s ./store/cache
