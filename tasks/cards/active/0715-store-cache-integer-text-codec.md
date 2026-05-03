# Card 0715: Store Cache Integer Text Codec

Priority: P1
State: active
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go

Goal:
Align `store/cache` integer operations with the package's byte-oriented cache contract.

Scope:
- Replace gob-only integer encoding with stable base-10 text encoding.
- Make `Incr` and `Decr` work after `Set(ctx, key, []byte("1"), ttl)`.
- Preserve `ErrNotInteger` for invalid integer bytes.
- Add overflow coverage for increment/decrement.

Non-goals:
- Do not add typed cache APIs.
- Do not add feature-specific counter abstractions.
- Do not change cache TTL preservation behavior.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Not required unless package comments mention integer storage format.

Done Definition:
- Integer cache values use readable decimal bytes.
- Existing invalid integer behavior remains covered.
- Focused tests and vet pass.
