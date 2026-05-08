# Card 0720: Store Cache Integer Boundary Cleanup

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
Depends On:

Goal:
Fix the remaining `store/cache` integer boundary bug and remove stale compatibility naming before the stable API snapshot is refreshed.

Scope:
- Enumerate `ErrCacheMiss` callers before removing the exported alias.
- Remove `ErrCacheMiss` and the compatibility-only test.
- Make `Decr` reject `math.MinInt64` without overflowing before validation.
- Add focused regression coverage for the boundary case.

Non-goals:
- Do not split the `Cache` interface in this card.
- Do not change normal `Incr`, `Decr`, or `Append` success semantics.
- Do not add provider-specific cache behavior.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Deferred to the stable API snapshot card.

Done Definition:
- `ErrCacheMiss` no longer appears in Go source.
- `Decr` handles `math.MinInt64` without integer overflow.
- Focused cache tests and vet pass.

Outcome:
- Removed the stale exported `ErrCacheMiss` compatibility alias after confirming it had no callers outside `store/cache`.
- Made `MemoryCache.Decr` reject `math.MinInt64` before negating the delta, preserving the `ErrNotInteger` overflow contract.
- Added regression coverage for the minimum-delta boundary.

Validation:
- rg -n --glob '*.go' 'ErrCacheMiss' .
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache
