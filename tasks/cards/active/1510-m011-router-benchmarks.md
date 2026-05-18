# Card 1510

Milestone: M-011
Recipe: specs/change-recipes/add-package.yaml
Priority: P1
State: active
Primary Module: benchmark
Owned Files:
- `benchmark/router_bench_test.go`

Goal:
- Write benchmark/router_bench_test.go covering static route match, single-param match,
  and multi-param match against both the plumego router and the chi router at the same
  path shapes.

Scope:
- Three benchmark functions: BenchmarkRouterStatic, BenchmarkRouterSingleParam,
  BenchmarkRouterMultiParam.
- Each benchmark function runs both plumego and chi variants as sub-benchmarks
  (b.Run("plumego", ...) and b.Run("chi", ...)).
- Use net/http/httptest.NewRecorder() and a synthetic *http.Request for each iteration.
- Import chi from the benchmark/go.mod (separate module); do not import chi from the
  main module.

Non-goals:
- Do not benchmark middleware chains in this card (that is card 1511).
- Do not add benchmark code to the router package itself.
- Do not import chi in the main module go.mod.

Files:
- `benchmark/router_bench_test.go`

Tests:
- `go test -bench=. -benchmem -count=3 ./benchmark/...`

Docs Sync:
- none; docs/benchmarks/results-v1.1.0.md is written by card 1512 after this card.

Done Definition:
- benchmark/router_bench_test.go exists and compiles.
- `go test -bench=. -benchmem -count=1 ./benchmark/...` exits 0 and produces ns/op and B/op output for all three shapes.
- Plumego and chi variants both appear in the output.

Outcome:
-
