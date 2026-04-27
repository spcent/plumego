# Card 5406: Health Model Copy Ownership

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: health
Owned Files:
- `health/core.go`
- `health/readiness.go`
- `health/health_test.go`
- `docs/modules/health/README.md`
- `x/ops/healthhttp/manager.go`
Depends On: none

Goal:
Make ownership of mutable health model fields explicit and reusable so callers
do not need local shallow-copy helpers for `Details`, `Dependencies`, or
`ReadinessStatus.Components`.

Scope:
- Add canonical copy helpers on `HealthStatus`, `ComponentHealth`, and
  `ReadinessStatus`.
- Preserve nil versus empty map and slice behavior.
- Document that maps and slices are copied, while `Details` values remain
  caller-owned.
- Replace the local x/ops component-health copy helper with the canonical
  health model helper.

Non-goals:
- Do not add health managers, HTTP handlers, retry policy, or reporting to
  stable `health`.
- Do not change JSON field names, JSON encoding, or health state semantics.
- Do not deep-copy arbitrary values stored in `HealthStatus.Details`.

Files:
- `health/core.go`
- `health/readiness.go`
- `health/health_test.go`
- `docs/modules/health/README.md`
- `x/ops/healthhttp/manager.go`

Tests:
- `go test -race -timeout 60s ./health/... ./x/ops/...`
- `go test -timeout 20s ./health/... ./x/ops/...`
- `go vet ./health/... ./x/ops/...`

Docs Sync:
Update `docs/modules/health/README.md` for copy helper semantics.

Done Definition:
- Health model copy helpers preserve values and protect copied maps/slices from
  caller mutation.
- x/ops health manager uses the canonical copy path.
- The listed validation commands pass.

Outcome:
