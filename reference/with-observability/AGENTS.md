# AGENTS.md - reference/with-observability

Operational guide for agents working in `reference/with-observability`.

This file is intentionally short. It compresses the local rules agents need for
routine changes in this feature demo. The repository root `AGENTS.md` remains
authoritative; when a task becomes architectural, security-sensitive, or
cross-module, fall back to the root workflow and the matching
`specs/task-routing.yaml` entry.

## 1. Minimal Context

For changes confined to this directory, load only:

- repository root `AGENTS.md`, if not already loaded
- this file
- the touched Go files and their tests

Read these local docs only when they answer a concrete question:

- `ARCHITECTURE.md`: layout rationale and middleware ordering
- `AGENT_TASKS.md`: adding metrics labels, inspection endpoints, or OTLP exporter wiring
- `PRODUCTION_CHECKLIST.md`: gating `/metrics`, removing span inspection, OTLP export

## 2. Purpose And Boundaries

`reference/with-observability` extends `standard-service` with production-grade
HTTP metrics (Prometheus) and distributed tracing (OpenTelemetry). It teaches
the minimal wiring change to go from a noop collector to a real one. Read
`standard-service` first.

Hard rules:

- No `x/*` imports beyond `x/observability`.
- No new third-party dependencies.
- No hidden globals, `init()` registration, reflection routing, or controller scanning.
- Keep `main.go` thin: load config, construct app, register routes, start.
- Keep middleware wiring explicit in `internal/app/app.go`.
- Keep all public routes explicit in `internal/app/routes.go`.
- Preserve `func(http.ResponseWriter, *http.Request)` handler shape.
- Use `contract.WriteResponse` for success and `contract.WriteError` with
  `contract.NewErrorBuilder()` for errors.

## 3. Package Ownership

- `main.go`: process entrypoint only.
- `internal/config`: app-local config loading, defaults, flags, environment.
- `internal/app/app.go`: core app construction, collector and tracer wiring,
  middleware order, graceful shutdown.
- `internal/app/routes.go`: route table — `/metrics` via exporter, inspection
  routes, health and identity routes.
- `internal/handler/api.go`: `APIHandler` (identity/hello), `ObservabilityHandler`
  (stats/spans), `MetricsHandler` (StatsReader demo).
- `internal/handler/health.go`: liveness and readiness probes.
- `internal/handler/write.go`: `logWriteErr` helper.

## 4. Change Patterns

Add an inspection endpoint:

1. Add handler method in `internal/handler/api.go` on an existing handler struct.
2. Register the route in `internal/app/routes.go` under the `v1` group.
3. Add a focused test in `internal/app/app_test.go`.

Change the metrics collector:

1. Replace `observability.NewPrometheusCollector` in `app.New` with the new collector.
2. Ensure it implements `metrics.HTTPObserver` for `httpmetrics.Middleware`.
3. Ensure it implements `metrics.StatsReader` for `MetricsHandler.Observer`.

Replace the in-process tracer with an OTLP exporter (production):

1. Construct an OTLP exporter instead of `observability.NewOpenTelemetryTracer`.
2. Wire it into `mwtracing.Middleware` — the interface is the same.
3. Remove the `GET /api/v1/spans` route from `routes.go`.

## 5. Validation

For docs-only changes:

```bash
git diff --check
```

For Go changes inside this service:

```bash
cd reference/with-observability && go test -race -timeout 30s ./...
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests
```

## 6. Review Focus

When reviewing or optimizing this service, check:

- accidental `x/*` imports beyond `x/observability`
- `/api/v1/spans` is not gated (must be removed or gated in production)
- `/metrics` is not protected (must be gated in production)
- `httpmetrics.Middleware` is wired with the real collector, not noop
- `mwtracing.Middleware` is wired after `httpmetrics` and before `timeout`
- response envelope uses `contract.WriteResponse` / `contract.WriteError` only
