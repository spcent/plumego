# Card 0834

Priority: P0
State: active
Primary Module: store
Owned Files:
- `store/db/stats.go`
- `store/db/slowquery.go`
- `store/module.yaml`
- `docs/modules/store/README.md`
- `x/observability`

Goal:
- Reduce stable `store/db` to transport-agnostic instrumentation primitives and remove analytics/reporting ownership that belongs in an observability extension.
- Converge database stats and slow-query inspection onto one clear extension-owned surface while fixing the current duplicate query-type counting bug in the same change.

Problem:
- Stable `store/db` still exports `DBStatsAggregator`, `AggregatingObserver`, `SlowQueryDetector`, `SlowQueryMetricsObserver`, summary strings, recent-record accessors, and top-slow-table reporting.
- These APIs are observability and operator-facing analytics surfaces rather than minimal persistence primitives.
- `DBStatsAggregator.RecordQuery` currently updates query-type counters in both `updateStats` and `updateQueryType`, which means `SelectQueries` / `InsertQueries` / `UpdateQueries` / `DeleteQueries` can be double-counted.
- `store/module.yaml` and `docs/modules/store/README.md` describe stable persistence primitives, not rich performance analytics and slow-query reporting ownership.

Scope:
- Decide the minimum stable `store/db` instrumentation surface that should remain in the stable root.
- Remove or relocate aggregation/reporting/inspection APIs that exceed that primitive boundary.
- Move richer database observability and slow-query inspection to a clearly owned extension under `x/observability`.
- Fix the current query-type double-counting bug as part of the convergence instead of preserving broken stats behavior.
- Update downstream callers, tests, and docs in the same change.

Non-goals:
- Do not add any database driver dependency.
- Do not move SQL execution or persistence contracts into `x/observability`.
- Do not preserve duplicate observer families or compatibility wrappers in stable `store/db`.
- Do not route this work through `health`; the owner is observability, not readiness.

Files:
- `store/db/stats.go`
- `store/db/slowquery.go`
- `store/module.yaml`
- `docs/modules/store/README.md`
- `x/observability`

Tests:
- `go test -timeout 20s ./store/... ./x/observability/... ./x/devtools/...`
- `go test -race -timeout 60s ./store/... ./x/observability/... ./x/devtools/...`
- `go vet ./store/... ./x/observability/... ./x/devtools/...`

Docs Sync:
- Keep the store manifest and primer aligned on the rule that stable `store` owns persistence primitives only, while richer DB analytics and slow-query inspection live under `x/observability`.

Done Definition:
- Stable `store/db` exposes only the minimal instrumentation primitives that truly belong with persistence contracts.
- Aggregation/reporting/slow-query inspection APIs no longer live in stable `store/db`.
- Query-type statistics are no longer double-counted.
- Downstream packages use the converged observability-owned surface with no residual references to deleted stable analytics APIs.
- Store docs and manifest describe the same reduced stable boundary the code implements.

Outcome:
- Pending.
