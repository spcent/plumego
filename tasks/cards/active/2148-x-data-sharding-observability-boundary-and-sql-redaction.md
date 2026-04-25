# Card 2148: x/data Sharding Observability Boundary and SQL Redaction

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: x/data
Owned Files:
- `x/data/sharding/tracing.go`
- `x/data/sharding/tracing_test.go`
- `x/data/sharding/metrics.go`
- `x/data/sharding/metrics_test.go`
- `docs/modules/x-data/README.md`
Depends On: none

Goal:
Keep sharding observability local and safe by removing raw SQL from trace
attributes and clarifying the boundary between topology metrics and generic
observability infrastructure.

Problem:
`x/data/sharding/tracing.go` defines local `Tracer`, `Span`, `SpanEvent`, and
`SpanStatus` types even though generic tracing infrastructure lives in
`x/observability`. It also records raw SQL in attributes such as `db.statement`,
`sql.original`, and shard-query traces. Raw SQL can contain literals or
application identifiers, so a topology helper should not expose it by default.
`metrics.go` separately emits Prometheus text directly, which is useful, but its
relationship to stable `metrics` and `x/observability` is not explicit.

Scope:
- Stop recording raw SQL strings in sharding trace attributes by default; keep
  operation, table/shard metadata, argument count, and safe query classification.
- Add focused tests proving trace attributes do not include raw SQL literals.
- Clarify whether local tracing types are intentionally no-op topology helpers
  or should be renamed/narrowed to avoid competing with `x/observability`
  tracing.
- Keep Prometheus text output behavior stable, but add tests for label escaping
  or output shape if current coverage is missing.
- Update the x/data primer with the observability boundary and raw-SQL policy.

Non-goals:
- Do not import `x/observability` into `x/data` unless the dependency rules and
  module manifest explicitly allow it.
- Do not remove existing public sharding metrics APIs in this card.
- Do not change sharding routing behavior.
- Do not add external tracing or Prometheus dependencies.

Files:
- `x/data/sharding/tracing.go`
- `x/data/sharding/tracing_test.go`
- `x/data/sharding/metrics.go`
- `x/data/sharding/metrics_test.go`
- `docs/modules/x-data/README.md`

Tests:
- `go test -race -timeout 60s ./x/data/sharding`
- `go test -timeout 20s ./x/data/sharding`
- `go vet ./x/data/sharding`

Docs Sync:
Update `docs/modules/x-data/README.md` with the sharding observability boundary
and the default no-raw-SQL trace attribute policy.

Done Definition:
- Sharding trace attributes no longer expose raw SQL text by default.
- Tests cover SQL redaction or omission in trace helpers.
- The module primer explains why sharding keeps topology metrics local and where
  generic tracing infrastructure belongs.
- The listed validation commands pass.

Outcome:
