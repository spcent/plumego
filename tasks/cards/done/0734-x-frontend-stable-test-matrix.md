# Card 0734: x/frontend Stable Test Matrix

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend_test.go`
- `x/frontend/frontend_bench_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0733

Goal:
Broaden stable-readiness evidence with targeted regression tests and basic
performance coverage.

Scope:
- Add coverage for uppercase MIME override behavior or document the exact
  case-sensitive contract.
- Add coverage for `RegisterFS` security boundary expectations.
- Add benchmarks for normal asset serving and precompressed asset serving.
- Update docs so test coverage claims match actual coverage.

Non-goals:
- Do not change production behavior unless a test exposes a bug.
- Do not add external benchmarking dependencies.
- Do not run repo-wide gates in this card.

Files:
- `x/frontend/frontend_test.go`
- `x/frontend/frontend_bench_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go test -bench=Benchmark -run '^$' ./x/frontend`
- `go vet ./x/frontend/...`

Docs Sync:
Update README and primer with accurate coverage and benchmark guidance.

Done Definition:
- Stable edge-case coverage is explicit.
- Benchmarks compile and run.
- Documentation no longer overclaims coverage.
- The listed validation commands pass.

Outcome:
- MIME type overrides now match extensions case-insensitively.
- Added coverage for uppercase asset extensions and unsafe `RegisterFS` request
  paths being rejected before backend open.
- Added benchmarks for normal and precompressed asset serving.
- Updated docs to describe benchmark coverage and stable edge-case coverage.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go test -bench=Benchmark -run '^$' ./x/frontend`
  - `go vet ./x/frontend/...`
