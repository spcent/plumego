# Production Checklist — with-observability

Run through this list before promoting this service to a production environment.
Items marked **CRITICAL** are security requirements; missing them leaves the
service in an unsafe state.

---

## Observability surfaces

- [ ] **CRITICAL: Gate `GET /metrics`** behind an internal network or
  bearer-token middleware. The Prometheus text exposition reveals endpoint
  structure, route cardinality, and error rates. Do not expose it publicly.

  Add per-route middleware in `routes.go`:
  ```go
  root.get("/metrics", metricsAuthMiddleware(exporter.Handler()))
  ```

- [ ] **CRITICAL: Remove `GET /api/v1/spans`** before going to production.
  This endpoint exposes trace internals (trace IDs, durations, HTTP attributes)
  without authentication. Remove the route from `routes.go` and the `Tracer`
  field from `ObservabilityHandler`.

- [ ] **Replace `OpenTelemetryTracer` with an OTLP exporter** pointing at your
  tracing backend (Jaeger, Zipkin, or a vendor collector). The
  `mwtracing.Middleware` interface is unchanged; only the constructor in
  `app.New` changes.

- [ ] **Set `APP_METRICS_NAMESPACE`** to your service name. The default is
  `plumego`. Use a unique namespace to avoid metric collisions when multiple
  services share a Prometheus instance.

- [ ] **Tune `APP_METRICS_MAX_SERIES`** to match your expected cardinality
  (distinct route × method × status combinations). The default is 10 000.

---

## Transport hardening

- [ ] **Read and write timeouts** are set.
  `cfg.Core.ReadTimeout`, `cfg.Core.WriteTimeout`, `cfg.Core.IdleTimeout`.
  Defaults are permissive for local development.

- [ ] **Body size limit** is configured for your traffic profile.
  `bodylimit.Middleware` reads `cfg.App.MaxBodyBytes`. The default is 1 MiB.

- [ ] **Request timeout** is configured for your p99 latency budget.
  `timeout.Middleware` defaults to 30s. Timeouts return **504 Gateway Timeout**,
  not 408. Set `cfg.Core.WriteTimeout` longer than the request timeout.

- [ ] **TLS is enabled** or the service runs behind a TLS-terminating proxy.
  If self-terminating, set `APP_TLS_ENABLED=true` and provide cert/key paths.

- [ ] **CORS policy** is tightened if the API is consumed by browser clients.
  Replace `cors.CORSOptions{}` with `cors.StrictDefaultOptions(origins...)`.

---

## Security

- [ ] **Security headers** are wired and cover your deployment requirements.
  `middleware/security` sets `X-Frame-Options`, `X-Content-Type-Options`, and
  `Referrer-Policy`. Review against your CSP and HSTS requirements.

- [ ] **Rate limiting** is configured for public endpoints.

- [ ] **Authentication** is enforced on all non-public endpoints.

---

## Observability

- [ ] **Structured access logging** is in the middleware stack.
  Confirm it is logging to the intended destination in your environment.

- [ ] **Request ID propagation** is in the middleware stack.
  Verify that downstream services and log aggregators can correlate requests.

- [ ] **Health endpoints** are reachable.
  `/healthz` (liveness) and `/readyz` (readiness) are registered.
  Confirm your orchestration layer is configured to probe them.

---

## Lifecycle

- [ ] **Graceful shutdown with signal handling** is wired.
  `main.run` creates a `signal.NotifyContext(SIGTERM, SIGINT)` and passes it to
  `app.Start(ctx)`. Verify the shutdown timeout covers your slowest expected
  request.

- [ ] **Configuration is loaded from the environment**, not from checked-in
  `.env` files. `.env` is for local development only.

---

## Testing

- [ ] **All app tests pass** with `go test -race ./...` from this directory.
- [ ] **Metrics accumulate**: after startup, `GET /metrics` returns
  `_uptime_seconds` and `_http_requests_total`.
- [ ] **Spans are recorded**: one span per request appears in
  `OpenTelemetryTracer.Spans()` (or your OTLP backend).
- [ ] **Boundary checks pass**: `go run ./internal/checks/dependency-rules`.
  This service must not import `x/*` beyond `x/observability`.
