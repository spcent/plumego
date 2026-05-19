# Card 1512

Milestone: M-011
Recipe: specs/change-recipes/update-docs.yaml
Priority: P1
State: done
Primary Module: benchmark
Owned Files:
- `benchmark/go.mod`
- `benchmark/README.md`
- `docs/benchmarks/results-v1.1.0.md`

Goal:
- Write benchmark/go.mod declaring the separate benchmark module with chi pinned at a
  specific version.
- Write benchmark/README.md explaining how to run the benchmarks and interpret the output.
- Capture verbatim `go test -bench=. -benchmem -count=3 ./benchmark/...` output and
  commit it to docs/benchmarks/results-v1.1.0.md with a machine-spec header.

Scope:
- benchmark/go.mod: module name github.com/spcent/plumego/benchmark; go 1.24.0;
  require plumego main module and chi at a pinned minor version.
- benchmark/README.md: explains prerequisites, run command, how to update results,
  and why chi is the only comparison target.
- docs/benchmarks/results-v1.1.0.md: header with date, Go version, OS/arch, CPU model;
  verbatim benchmark output pasted below the header.

Non-goals:
- Do not add chi to the main module go.mod.
- Do not embed raw benchmark numbers in README.md or docs/why-plumego.md directly;
  reference docs/benchmarks/results-v1.1.0.md by path only.
- Do not run benchmarks against live network or external services.

Files:
- `benchmark/go.mod`
- `benchmark/README.md`
- `docs/benchmarks/results-v1.1.0.md`

Tests:
- `go test -bench=. -benchmem -count=3 ./benchmark/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- docs/why-plumego.md: add a sentence referencing docs/benchmarks/results-v1.1.0.md
  under the performance section.

Done Definition:
- benchmark/go.mod exists as a separate module (not in the main module go.mod).
- docs/benchmarks/results-v1.1.0.md has a machine-spec header and verbatim benchmark output.
- `go run ./internal/checks/dependency-rules` exits 0 confirming no chi import in main module.
- docs/why-plumego.md references docs/benchmarks/results-v1.1.0.md by path.

Outcome:
- Done. Added benchmark README and v1.1.0 benchmark results with machine
  metadata and verbatim `count=3` output, kept chi isolated to the benchmark
  module, and referenced the result file from docs/why-plumego.md. Validated
  benchmark run, dependency-rules, and module-manifests.
