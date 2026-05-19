# Card 0958

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md

Goal:
Normalize stable cache key validation errors so request key errors are not reported as configuration errors.

Scope:
- Return `ErrInvalidKey` for empty keys and control-character keys.
- Preserve `ErrKeyTooLong` for configured length limit violations.
- Add tests covering error identity for invalid keys.
- Sync store docs with key error taxonomy if needed.

Non-goals:
- Do not change KV key policy in this card.
- Do not change delete missing-key semantics.
- Do not add metrics or cache stats.

Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/cache
- go vet ./store/cache
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update store module docs if the key error matrix is documented.

Done Definition:
- Invalid cache keys match `ErrInvalidKey` and no longer match `ErrInvalidConfig`.
- Existing cache behavior remains compatible otherwise.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Changed empty and control-character cache keys to return `ErrInvalidKey` without wrapping `ErrInvalidConfig`.
- Preserved `ErrKeyTooLong` for configured key length violations.
- Updated cache tests for sentinel identity and store docs for the key error taxonomy.

Validation:
- `gofmt -w store/cache/cache.go store/cache/cache_test.go`
- `go test -timeout 20s ./store/cache`
- `go vet ./store/cache`
- `go run ./internal/checks/dependency-rules`
