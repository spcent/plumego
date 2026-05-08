# Card 1235

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P3
State: done
Primary Module: router
Owned Files: router/path.go, router/router_bench_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0760-router-nil-router-options

Goal:
Align cache key helper complexity, comments, and benchmark evidence.

Scope:
- Replace or document the cache key helper so comments match actual allocation
  behavior.
- Prefer simpler standard-library-free code if benchmarks do not justify the
  pooled builder.
- Run and record `BenchmarkOptCacheKey` evidence.
- Keep route cache behavior unchanged.

Non-goals:
- Changing cache key format.
- Reworking match cache eviction or locking.
- Adding non-stdlib benchmark dependencies.

Files:
- router/path.go
- router/router_bench_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go vet ./router/...
- go test -run '^$' -bench BenchmarkOptCacheKey -benchmem ./router

Docs Sync:
- Required if benchmark guidance changes.

Done Definition:
- Cache key helper comments and implementation are consistent with benchmark
  evidence.
- `BenchmarkOptCacheKey` evidence is recorded in the done card.
- Router targeted tests and vet pass.

Outcome:
- Replaced the pooled `strings.Builder` cache key helper with direct string
  concatenation.
- Updated router docs to record that cache key construction is intentionally
  simple unless benchmark evidence changes.
- Recorded benchmark evidence on darwin/arm64 Apple M1 Pro:
  - before: `BenchmarkOptCacheKey-10` 51.14 ns/op, 26 B/op, 2 allocs/op
  - after: `BenchmarkOptCacheKey-10` 18.14 ns/op, 0 B/op, 0 allocs/op

Validation:
- go test -timeout 20s ./router/...
- go vet ./router/...
- go test -run '^$' -bench BenchmarkOptCacheKey -benchmem ./router
