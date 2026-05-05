# Card 0731

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md
Depends On:
- 0730

Goal:
Freeze nil and empty byte-slice value semantics for stable cache and KV primitives.

Scope:
- Decide and document whether empty values are preserved distinctly from missing values.
- Ensure cache and KV defensive copies preserve zero-length non-nil values where stable behavior requires it.
- Add tests for nil value and empty slice round trips.

Non-goals:
- Do not add serializers or typed value codecs.
- Do not change cache key or KV key policy.
- Do not add distributed cache behavior.

Files:
- store/cache/cache.go
- store/cache/cache_test.go
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/cache ./store/kv
- go vet ./store/cache ./store/kv

Docs Sync:
- Update store README with byte-value ownership and empty-value semantics.

Done Definition:
- Empty value behavior is tested and documented for cache and KV.
- Returned values remain caller-owned.
- Missing keys remain distinguishable through errors/exists APIs.

Outcome:
- Documented byte-value ownership and existence semantics for stable cache and KV primitives.
- Updated cache and KV byte cloning so non-nil zero-length values remain non-nil after round trips while nil values remain nil.
- Added cache and KV tests covering nil values, empty values, and missing-key distinction.

Validation:
- `gofmt -w store/cache/cache.go store/cache/cache_test.go store/kv/kv.go store/kv/kv_test.go`
- `go test -timeout 20s ./store/cache ./store/kv`
- `go vet ./store/cache ./store/kv`
- `go run ./internal/checks/dependency-rules`
