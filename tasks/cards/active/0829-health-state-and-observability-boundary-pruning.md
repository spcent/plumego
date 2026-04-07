# Card 0829

Priority: P1
State: active
Primary Module: health
Owned Files:
- `health/core.go`
- `health/history.go`
- `health/metrics.go`
- `health/module.yaml`
- `docs/modules/health/README.md`
Depends On:

Goal:
- Restore stable `health` to state models, readiness helpers, and the minimum in-process manager surface by moving observability- and ops-heavy reporting behavior out of the stable root.
- Remove the current mixed “health state + ops analytics” API boundary and converge to a smaller stable ownership model.

Problem:
- `health/module.yaml` and `docs/modules/health/README.md` describe `health` as models/readiness support, but the package still exports history querying, retention stats, metrics tracking, trend analysis, comprehensive reports, and build metadata.
- `HealthManager` currently combines component registration, readiness, history querying, metrics attachment, retention configuration, and cleanup orchestration into one long-lived stable interface.
- `x/ops/healthhttp` already depends on these richer reporting surfaces, which shows that stable `health` has become an ops/observability capability pack rather than a small state package.

Scope:
- Split stable health state/readiness ownership from history, metrics, trend, and reporting ownership.
- Remove or relocate history-query/reporting/metrics tracker APIs from the stable root to an owning extension package.
- Narrow `HealthManager` and related interfaces so the stable surface reflects state management instead of observability analytics.
- Keep readiness and component health aggregation explicit and transport-agnostic.
- Sync the health module manifest and primer to the reduced stable surface.

Non-goals:
- Do not add HTTP handlers to stable `health`.
- Do not redesign the core component-checker contract unless required to complete the boundary cleanup.
- Do not move feature-specific health checks into stable `health`.
- Do not preserve analytics/reporting APIs in stable `health` for compatibility.

Files:
- `health/core.go`
- `health/history.go`
- `health/metrics.go`
- `health/module.yaml`
- `docs/modules/health/README.md`

Tests:
- `go test -timeout 20s ./health/... ./x/ops/healthhttp/...`
- `go test -race -timeout 60s ./health/... ./x/ops/healthhttp/...`
- `go vet ./health/... ./x/ops/healthhttp/...`

Docs Sync:
- Keep the health primer and manifest aligned on the rule that stable `health` owns state and readiness only, while history/reporting exposure and analytics belong in owning extensions.

Done Definition:
- Stable `health` no longer owns observability-heavy history, metrics-tracker, trend, or reporting APIs.
- `HealthManager` and related stable interfaces describe a reduced state/readiness boundary.
- Any relocated health analytics surface lives in an owning extension package with explicit docs.
- `x/ops` or another owning extension consumes the relocated reporting APIs instead of stable `health`.
- Module docs and manifest describe the same reduced stable boundary the code implements.

Outcome:
- Pending.
