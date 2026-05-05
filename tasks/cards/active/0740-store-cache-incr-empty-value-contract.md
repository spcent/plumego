# Card 0740

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md

Goal:
Align cache integer operations with stable empty-byte-value semantics.

Scope:
- Treat an existing empty cache value as a non-integer for `Incr` and `Decr`.
- Preserve missing-key creation behavior.
- Add focused tests for empty value increment/decrement behavior.
- Sync store docs with integer operation semantics.

Non-goals:
- Do not change cache byte-value round-trip behavior.
- Do not change KV behavior.
- Do not add typed codecs.

Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/cache
- go vet ./store/cache
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update store module docs for cache integer operations on existing empty values.

Done Definition:
- Existing empty cache values return `ErrNotInteger` for `Incr`/`Decr`.
- Missing keys still create integer values.
- Targeted tests, vet, and dependency checks pass.
