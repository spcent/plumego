# Card 0420: AI Metrics Tags Safe Construction
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: done
Primary Module: x/ai
Owned Files:
- x/ai/metrics/metrics.go
- x/ai/metrics/metrics_test.go
- docs/modules/x-ai/README.md
Depends On: none

Goal:
Add a non-panicking metric-tag construction path. `metrics.Tags` is a convenience
helper that panics on an odd number of string arguments, which is awkward for
runtime-built tags and inconsistent with the repository preference for explicit
error paths in reusable helpers.

Scope:
- Add an error-returning tag builder for runtime inputs.
- Keep `Tags` as the short convenience wrapper for static call sites if that
  preserves compatibility.
- Add tests for even input, odd input error behavior, and the compatibility
  wrapper.
- Keep the collector interfaces unchanged.

Non-goals:
- Do not introduce a metrics backend dependency.
- Do not change metric key rendering or snapshot semantics.
- Do not rewrite existing static `Tags(...)` call sites unless needed by tests.

Tests:
- go test -race -timeout 60s ./x/ai/metrics
- go test -timeout 20s ./x/ai/metrics
- go vet ./x/ai/metrics

Docs Sync:
Update `docs/modules/x-ai/README.md` if it documents metric tag helpers.

Done Definition:
- Runtime metric-tag construction can return an error instead of panicking.
- Existing static `Tags(...)` convenience behavior remains test-covered.
- The x/ai/metrics validation commands pass.

Outcome:
Added `TagsE` with `ErrOddTagArguments` for runtime tag construction and kept
`Tags` as the static convenience wrapper. Tests cover successful `TagsE`
construction, odd-input error behavior, and the compatibility panic wrapper.

Validation:
- `go test -race -timeout 60s ./x/ai/metrics`
- `go test -timeout 20s ./x/ai/metrics`
- `go vet ./x/ai/metrics`
