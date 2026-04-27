# Card 0326

Milestone:
Recipe: specs/change-recipes/refine-api.yaml
Priority: P1
State: active
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
Depends On:
- 0325-store-package-example-polish

Goal:
Give `store/cache` invalid-key failures a stable sentinel distinct from configuration errors while preserving compatibility.

Scope:
- Add `ErrInvalidKey` for empty and control-character cache keys.
- Keep `ErrInvalidConfig` matchability for existing callers that treated invalid keys as config failures.
- Keep `ErrKeyTooLong` behavior unchanged.
- Add focused sentinel and invalid-key tests.

Non-goals:
- Do not change the `Cache` interface.
- Do not add tenant-aware key policy or HTTP cache behavior.
- Do not change max-key-length semantics.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Not required.

Done Definition:
- Empty/control-character cache keys match `ErrInvalidKey`.
- Existing empty/control-character key errors still match `ErrInvalidConfig`.
- Cache targeted tests and vet pass.
