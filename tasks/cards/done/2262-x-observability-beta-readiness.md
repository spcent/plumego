# Card 2262

Milestone:
Recipe: specs/change-recipes/review-only.yaml
Priority: P1
State: done
Primary Module: x/observability
Owned Files:
- x/observability/module.yaml
- docs/modules/x-observability/README.md
- docs/ROADMAP.md
- docs/EXTENSION_STABILITY_POLICY.md
Depends On: 2261

Goal:
Evaluate `x/observability` for `beta` readiness and make its production-adapter status explicit.

Scope:
- Review exporter, collector, tracer, testlog, and testmetrics entrypoints against beta criteria.
- Confirm the boundary between stable transport observability middleware and `x/observability` adapter/export wiring.
- If criteria are met, update status and docs.
- If criteria are not met, record the missing coverage or API-freeze evidence.

Non-goals:
- Do not move exporter wiring into stable `middleware`.
- Do not add new metrics backends.
- Do not change existing metric or trace record semantics.

Files:
- `x/observability/module.yaml`
- `docs/modules/x-observability/README.md`
- `docs/ROADMAP.md`
- `docs/EXTENSION_STABILITY_POLICY.md`

Tests:
- `go test -timeout 20s ./x/observability/...`
- `go vet ./x/observability/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required when status or blocker language changes.

Done Definition:
- `x/observability` has a documented promotion decision.
- Stable middleware and extension observability responsibilities remain distinct.

Outcome:
Completed. `x/observability` remains `experimental` because the repository has
no verifiable two-minor-release API freeze evidence for exported
`x/observability/*` symbols. The module primer, roadmap, and extension
stability policy now record that blocker while preserving the stable middleware
versus extension exporter boundary.
