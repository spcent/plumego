# Card 1398

Milestone: M-004
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/config.go
- store/cache/memory.go
- store/cache/stats.go
- store/cache/cache_test.go
Depends On:
- 1394

Goal:
Mechanically split `store/cache` so cache configuration, memory implementation, errors, and stats are easier to review.

Scope:
- Move cache config/defaults into a focused file.
- Move `MemoryCache` implementation into a focused file.
- Move stats helpers into a focused file.
- Preserve `ErrCacheMiss = ErrNotFound` compatibility and all public method names.
- Keep tests behavior-only; add no new feature surface.

Non-goals:
- Do not change cache eviction, TTL, close, context, or write-boundary behavior.
- Do not introduce distributed cache, metrics export, HTTP cache, or provider config into stable `store/cache`.
- Do not create a generic shared bytes helper package.

Files:
- store/cache/cache.go
- store/cache/config.go
- store/cache/memory.go
- store/cache/stats.go
- store/cache/cache_test.go

Tests:
- go test -race -timeout 60s ./store/cache
- go test -timeout 20s ./store/cache
- go vet ./store/cache

Docs Sync:
- None expected unless exported documentation comments move or change.

Done Definition:
- `store/cache` behavior and exported API remain unchanged.
- Large-file review radius is reduced by mechanical file split.
- Target checks pass.

Outcome:
