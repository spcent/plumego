# Agent Tasks — with-observability

Operating guide for AI coding agents working in this feature demo.

Read `AGENTS.md` (repository root) for the full operating contract.
Read `ARCHITECTURE.md` (this directory) for layout rationale.
Read `docs/modules/x/observability/README.md` for the `x/observability` API.

---

## Zone Classification

### Safe to modify

| Path | What you can do |
|---|---|
| `internal/handler/api.go` | Add inspection endpoints to existing handler structs |
| `internal/app/routes.go` | Add routes; do not reorder existing registrations |
| `internal/config/config.go` | Add config fields; keep `Defaults()` safe for local dev |

### Restricted — requires preflight + reviewer note

| Path | Constraint |
|---|---|
| `internal/app/app.go` (middleware chain) | Order is load-bearing; changing it affects all routes and span timing |
| `internal/app/app.go` (collector/tracer wiring) | Changing the concrete types affects the `/metrics` and `/api/v1/spans` contract |

### Frozen

- Do not remove `/metrics` or change its content type from `text/plain`.
- Do not gate `/api/v1/spans` with auth in this demo — gate it in production by removing the route.
- Do not add `x/*` imports beyond `x/observability`.
- Do not add `init()` functions, global variables, or controller scanning.

---

## Common Task Recipes

### Add a new inspection endpoint

Add a method to `ObservabilityHandler` or `MetricsHandler` in
`internal/handler/api.go`:

```go
// NewInspect returns some new diagnostic data.
//
//	GET /api/v1/inspect → 200 inspectResponse
func (h ObservabilityHandler) NewInspect(w http.ResponseWriter, r *http.Request) {
    logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, inspectData, nil))
}
```

Register the route in `internal/app/routes.go` under the `v1` group:

```go
v1.get("/inspect", http.HandlerFunc(obs.NewInspect))
```

Add a test in `internal/app/app_test.go` verifying the route is registered and
the response is 200 with `Content-Type: application/json`.

### Replace the in-process tracer with an OTLP exporter

In `app.New` (`internal/app/app.go`), replace:

```go
tracer := observability.NewOpenTelemetryTracer(cfg.App.ServiceName)
```

with an OTLP gRPC exporter (consult `x/observability` docs for the constructor).
Wire it into `mwtracing.Middleware(tracer)` — the interface is identical.

Remove `GET /api/v1/spans` from `routes.go` and the `Tracer` field from the
`ObservabilityHandler` struct in `routes.go`. Delete `Tracer` from `App`.

### Change the Prometheus namespace

Update `APP_METRICS_NAMESPACE` in the environment or `.env` file. The namespace
is the prefix on every metric name (e.g. `plumego_http_requests_total`). Use your
service name to avoid collisions when multiple services share a Prometheus
instance.

### Increase max series cardinality

Update `APP_METRICS_MAX_SERIES`. The default is 10 000. Raise it if you have many
distinct label combinations (many routes × many status codes). Monitor memory usage
after the change.

### Add per-route middleware (e.g. auth on /metrics)

Wrap the handler at registration time in `routes.go`:

```go
root.get("/metrics", authMiddleware(exporter.Handler()))
```

Do not add auth inside the handler — keep the middleware boundary explicit.

---

## Validation Commands

```bash
# Module tests
cd reference/with-observability && go test -race -timeout 30s ./...

# Boundary checks
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests

# Full gates (cross-module or release-relevant changes)
make gates
```

---

## Non-goals

- Do not add domain models or business logic.
- Do not add authentication to this demo — production gating belongs in your
  service, not in the reference.
- Do not use this demo as the starting point for production services without
  working through the production checklist.
- Do not mount `x/observability/devtools` — it is explicitly excluded.
