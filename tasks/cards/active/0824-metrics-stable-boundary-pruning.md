# Card 0824

Priority: P1
State: active
Primary Module: metrics
Owned Files:
- `metrics/otel.go`
- `metrics/prometheus.go`
- `metrics/exporter.go`
- `metrics/devcollector.go`
- `metrics/smsgateway/metrics.go`
- `docs/modules/metrics/README.md`
- `metrics/module.yaml`
Depends On:
- `0822-observability-id-contract-convergence.md`

Goal:
- Shrink `metrics` back to stable contracts and base collectors by removing exporter-, tracer-, dev-dashboard-, and feature-specific ownership from the stable root.
- Treat this as a boundary reset: keep only the canonical stable surface and remove the rest instead of preserving overlapping stable APIs.

Problem:
- The `metrics` primer and manifest say the stable root should hold interfaces and base collectors, while exporter brand wiring belongs in owning extensions.
- The current package still owns `OpenTelemetryTracer`, `PrometheusCollector` + exporter surface, `DevCollector`, and `metrics/smsgateway`, which are not generic stable contracts.
- This makes the stable observability surface larger than the repo blueprint allows and forces unrelated concerns to share one long-lived API boundary.

Scope:
- Extract or demote non-generic observability entrypoints so `metrics` keeps only the minimal stable collector contracts and base implementations.
- Remove exporter-brand ownership and tracer lifecycle ownership from the stable root.
- Remove dev-only and feature-specific collector families from the stable root or clearly hand them off to owning extensions.
- Sync module docs and manifest metadata to the post-pruning stable surface.
- Delete old stable entrypoints after relocation; do not leave compatibility shims in `metrics`.

Non-goals:
- Do not add any third-party metrics or tracing dependency.
- Do not redesign the base `metrics.Recorder` / observer interfaces unless required to complete the boundary cleanup.
- Do not move observability concerns into `core`.
- Do not keep exporter or tracer brands in the stable package for API continuity.

Files:
- `metrics/otel.go`
- `metrics/prometheus.go`
- `metrics/exporter.go`
- `metrics/devcollector.go`
- `metrics/smsgateway/metrics.go`
- `docs/modules/metrics/README.md`
- `metrics/module.yaml`

Tests:
- `go test -timeout 20s ./metrics/... ./x/observability/... ./x/devtools/...`
- `go test -race -timeout 60s ./metrics/... ./x/observability/... ./x/devtools/...`
- `go vet ./metrics/... ./x/observability/... ./x/devtools/...`

Docs Sync:
- Keep the metrics module primer and manifest aligned on the rule that stable `metrics` owns small contracts and base collectors only.

Done Definition:
- Stable `metrics` no longer owns exporter-brand entrypoints, tracer infrastructure, dev-dashboard collectors, or feature-specific collector families.
- Any remaining stable metrics surface is transport-agnostic, small, and documented as such.
- Module docs and manifest describe the same post-cleanup boundary the code implements.
- No deprecated forwarding surface remains in `metrics` for relocated capabilities.

Outcome:
- Pending.
