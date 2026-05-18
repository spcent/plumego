# Plan for M-011: Benchmark Suite Publication

Milestone: `M-011`
Objective: Publish a benchmark suite in a separate benchmark/ module covering
router matching and middleware chain throughput, with CI-captured results committed
to docs/benchmarks/ and references added to docs/why-plumego.md for community
trust and framework positioning.
Constraints: no external dependencies added to the main module go.mod, benchmark
module is self-contained with its own go.mod, no inline benchmark numbers in
docs (reference results file by path only), chi comparison only (no Fiber/Gin).
Affected Modules: benchmark, docs.

## Phase Map

- Phase 1: Orient — confirm benchmark/ directory does not exist; read router and
  middleware module.yaml files.
- Phase 2: Implement (parallel) — write router benchmarks, middleware chain benchmarks,
  and CI infrastructure files concurrently.
- Phase 3: Validate and Ship — run benchmarks, update docs/why-plumego.md, commit.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1510 | Write router benchmark tests comparing plumego and chi | benchmark | `benchmark/router_bench_test.go` | M-008 | `go test -bench=. -benchmem ./benchmark/...` |
| 1511 | Write middleware chain benchmark tests for 1, 3, and 5 layer chains | benchmark | `benchmark/chain_bench_test.go` | M-008 | `go test -bench=. -benchmem ./benchmark/...` |
| 1512 | Write benchmark module files and commit CI-captured results | benchmark | `benchmark/go.mod`, `benchmark/README.md`, `docs/benchmarks/results-v1.1.0.md` | 1510, 1511 | file exists, verbatim output present |

## Dependency Edges

- `1510 -> 1512`
- `1511 -> 1512`

## Parallel Groups

- Group A (parallel): cards 1510 and 1511 — independent test files.
- Group B (sequential after A): card 1512 — needs both test files to exist so results
  can be captured and committed.

## Risk Register

- Risk: chi router API is incompatible with the current pinned version.
  Mitigation: pin chi to a specific minor version in benchmark/go.mod; update the pin
  if needed but never import chi in the main module.
- Risk: benchmark results vary significantly between runs, making comparisons misleading.
  Mitigation: run with -count=3 and report the median; include the machine spec in
  docs/benchmarks/results-v1.1.0.md header.

## Verification Strategy

- Card-level checks: each benchmark card runs `go test -bench=. -benchmem -count=3`
  immediately after writing the test file to confirm it compiles and produces output.
- Milestone-level checks: run full Acceptance Criteria suite including dependency-rules
  to confirm no main module go.mod pollution.
- Docs check: verify docs/why-plumego.md references docs/benchmarks/results-v1.1.0.md
  by path without embedding numbers.

## Exit Condition

- all three cards completed
- benchmark/ directory exists with router and chain test files
- benchmark/go.mod is separate from the main module
- docs/benchmarks/results-v1.1.0.md has verbatim benchmark output with machine spec header
- docs/why-plumego.md references the results file
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
