# Card 0756

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P3
State: done
Primary Module: router
Owned Files: router/pool.go, router/cache.go, router/router_bench_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0755-router-static-preflight-dedup

Goal:
Record benchmark evidence for keeping current pool and cache complexity in the
stable router.

Scope:
- Add focused benchmark coverage for hot match paths that exercise path parts,
  param values, route matcher pooling, and cache hits.
- Run router benchmarks enough to record evidence in the card outcome.
- Add a short docs note only if benchmark evidence is worth preserving for
  future maintainers.
- Mark the active queue empty after this final card completes.

Non-goals:
- Rewriting pool/cache internals unless benchmark evidence clearly exposes a
  correctness or performance issue.
- Adding non-stdlib benchmark dependencies.
- Changing public router behavior.

Files:
- router/pool.go
- router/cache.go
- router/router_bench_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
- go test -run '^$' -bench 'BenchmarkOpt(StaticRoute|ParamRoute|ParallelStatic)' -benchmem ./router

Docs Sync:
- Optional, only if evidence guidance is added.

Done Definition:
- Benchmark evidence is recorded in the done card.
- Any retained complexity has an explicit measured basis.
- Router targeted tests, race tests, vet, and selected benchmarks pass.
- Active queue is empty.

Outcome:
- Added `BenchmarkOptParamRoute` with explicit cache-hit and cache-miss
  sub-benchmarks for parameterized dispatch.
- Documented the targeted pool/cache benchmark command for future router
  simplification or expansion work.
- Recorded benchmark evidence on darwin/arm64 Apple M1 Pro:
  - `BenchmarkOptStaticRoute-10`: 230.9 ns/op, 488 B/op, 6 allocs/op
  - `BenchmarkOptParamRoute/cache_hit-10`: 814.3 ns/op, 1792 B/op, 13 allocs/op
  - `BenchmarkOptParamRoute/cache_miss-10`: 1104 ns/op, 1995 B/op, 17 allocs/op
  - `BenchmarkOptParallelStatic-10`: 276.6 ns/op, 488 B/op, 6 allocs/op

Validation:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...
- go test -run '^$' -bench 'BenchmarkOpt(StaticRoute|ParamRoute|ParallelStatic)' -benchmem ./router
