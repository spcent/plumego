# Card 1511

Milestone: M-011
Recipe: specs/change-recipes/add-package.yaml
Priority: P1
State: active
Primary Module: benchmark
Owned Files:
- `benchmark/chain_bench_test.go`

Goal:
- Write benchmark/chain_bench_test.go covering 1-layer, 3-layer, and 5-layer middleware
  chains with a no-op final handler and a realistic handler that writes JSON.

Scope:
- Six benchmark functions: BenchmarkChain1NoOp, BenchmarkChain3NoOp, BenchmarkChain5NoOp,
  BenchmarkChain1JSON, BenchmarkChain3JSON, BenchmarkChain5JSON.
- Middleware layers are func(http.Handler) http.Handler; each layer calls next exactly once
  and adds a response header (to avoid dead-code elimination by the compiler).
- JSON handler uses json.NewEncoder(w).Encode on a fixed struct.
- Use net/http/httptest.NewRecorder() and a synthetic *http.Request for each iteration.
- No chi import in this file; plumego chain only.

Non-goals:
- Do not benchmark route matching in this card (that is card 1510).
- Do not add benchmark code to the middleware package itself.
- Do not use reflect or unsafe to construct middleware chains.

Files:
- `benchmark/chain_bench_test.go`

Tests:
- `go test -bench=. -benchmem -count=3 ./benchmark/...`

Docs Sync:
- none; docs/benchmarks/results-v1.1.0.md is written by card 1512 after this card.

Done Definition:
- benchmark/chain_bench_test.go exists and compiles.
- `go test -bench=. -benchmem -count=1 ./benchmark/...` exits 0 and produces ns/op and B/op
  for all six benchmark functions.
- Chain depth is visible in benchmark names (1, 3, 5 suffix or sub-benchmark).

Outcome:
-
