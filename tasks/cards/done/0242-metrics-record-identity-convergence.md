# Card 0242

Priority: P1
State: done
Primary Module: metrics
Owned Files:
- `metrics/collector.go`
- `metrics/*_test.go`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`
- `x/observability`
- `x/messaging`
- `x/ai`

Goal:
- Converge stable metric identity onto one canonical field instead of carrying both `Type` and `Name`.
- Keep stable `metrics` generic and avoid maintaining two parallel ways to describe the same record identity.

Problem:
- `metrics.MetricRecord` still exposes both `Type` and `Name`, while stable `metrics` only retains one shared stable type constant (`MetricHTTPRequest`).
- Extensions currently use a mix of `Type`, `Name`, and `Labels`, which leaves collector stats and observability adapters interpreting metric identity through two overlapping fields.
- This is the remaining stable metrics surface that still feels like taxonomy ownership rather than a single generic record contract.

Scope:
- Pick one canonical stable metric identity field and remove the other in the same change.
- Update stable collectors, stats payloads, and extension call sites accordingly.
- Keep shared HTTP observation intact, but express it through the converged record identity model.
- Sync metrics docs and manifest language to the final single-identity contract.

Non-goals:
- Do not redesign collector composition or the HTTP observer interface itself.
- Do not reintroduce a feature-specific metric taxonomy into stable `metrics`.
- Do not leave deprecated compatibility fields behind.

Files:
- `metrics/collector.go`
- `metrics/*_test.go`
- `metrics/module.yaml`
- `docs/modules/metrics/README.md`
- `x/observability`
- `x/messaging`
- `x/ai`

Tests:
- `go test -timeout 20s ./metrics/... ./x/observability/... ./x/messaging ./x/ai/...`
- `go test -race -timeout 60s ./metrics/... ./x/observability/... ./x/messaging ./x/ai/...`
- `go vet ./metrics/... ./x/observability/... ./x/messaging ./x/ai/...`

Docs Sync:
- Keep metrics docs aligned on the rule that stable `metrics` owns a single generic record identity model.

Done Definition:
- Stable `MetricRecord` uses one canonical identity field only.
- Stable collector stats and extension record producers no longer depend on parallel `Type` and `Name` identity semantics.
- Metrics docs and manifest describe the same converged record contract implemented in code.
