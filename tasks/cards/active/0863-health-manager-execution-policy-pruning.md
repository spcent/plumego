# Card 0863: Health Manager Execution Policy Pruning

Priority: P1
State: active
Primary Module: health

## Goal

Keep stable `health` focused on reusable health/readiness models and small helpers. Move operational check execution policy, retry behavior, timeout behavior, mutable manager configuration, and HTTP-facing orchestration into `x/ops/healthhttp`.

## Problem

The stable `health` package still owns an operational manager:

- `HealthManager`
- `NewHealthManager`
- `HealthCheckConfig`
- `CheckAllComponents`
- `CheckComponent`
- `GetAllHealth`
- `GetOverallHealth`
- `SetConfig`
- `GetConfig`

That manager performs concurrent checks, timeout wrapping, retry loops, component registration, status aggregation, and mutable config management. This is more than a stable model/readiness primitive and makes `x/ops/healthhttp` depend on stable operational policy instead of owning the HTTP/ops behavior directly.

The target shape is:

- stable `health`: status/readiness/component result primitives and simple readiness helpers
- `x/ops/healthhttp`: manager implementation, check execution policy, HTTP handlers, reporter-owned formatting

## Scope

- Keep stable health status models and readiness primitives that are reusable outside HTTP.
- Move manager execution policy to `x/ops/healthhttp`.
- Remove stable exported manager/config symbols once all callers are migrated.
- Update `x/ops/healthhttp` to own its manager interface or concrete type directly.
- Keep `health` free of HTTP handler ownership.
- Keep `health` free of ops/reporting formatting.
- Update module manifests and docs to reflect the new ownership split.

## Non-Goals

- Do not remove the ability to register and run health checks through `x/ops/healthhttp`.
- Do not change stable health status names unless required by the migration.
- Do not introduce a new stable observability or reporting dependency.
- Do not move this manager to stable `core`.

## Expected Files

- `health/core.go`
- `health/manager.go`
- `health/readiness.go`
- `health/*_test.go`
- `x/ops/healthhttp/*`
- `docs/modules/health/README.md`
- `health/module.yaml`
- `x/ops/module.yaml`

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./health/... ./x/ops/...
go test -race -timeout 60s ./health/... ./x/ops/...
go vet ./health/... ./x/ops/...
```

Then run the required repo-wide gates before committing.

## Done Definition

- Stable `health` no longer exports manager/config/check-execution policy symbols.
- `x/ops/healthhttp` owns health check orchestration and HTTP-facing manager behavior.
- No stable package imports `x/ops`.
- Existing health HTTP capability is preserved under `x/ops/healthhttp`.
- Removed exported health symbols have zero residual references.
- Focused gates and repo-wide gates pass.
