# Card 0714: Store Cache Close Contract

Priority: P1
State: active
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go

Goal:
Make `store/cache.MemoryCache` lifecycle behavior explicit and consistent with stable store primitives after `Close`.

Scope:
- Add a stable closed-cache sentinel error.
- Track closed state in `MemoryCache`.
- Make public cache operations return the closed error after `Close`.
- Keep `Close` idempotent.
- Add focused coverage for post-close reads and writes.

Non-goals:
- Do not add provider-specific cache behavior.
- Do not change the `Cache` interface shape.
- Do not add cache metrics or introspection.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Not required unless public package comments need lifecycle wording.

Done Definition:
- Closed cache operations fail consistently with a namespaced sentinel.
- Repeated `Close` remains nil.
- Focused tests and vet pass.
