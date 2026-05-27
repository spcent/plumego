# Production Checklist — production-service

Run through this list in addition to
`reference/standard-service/PRODUCTION_CHECKLIST.md` before deploying.
`production-service` extends `standard-service`; all inherited items apply.

---

## Authentication

- [ ] **`APP_API_TOKEN` is set** via a deployment secret.
  The profile route (`GET /api/profile`) uses `middleware/auth` with
  `RequireBearer`. When `APP_API_TOKEN` is empty the handler fails closed with
  `401 Unauthorized` — there is no fallback. Do not set a default in the
  deployment environment.

- [ ] **`OPS_TOKEN` is set** via a deployment secret.
  `GET /ops/metrics` uses the same bearer check. An unset token makes the
  endpoint unconditionally inaccessible, which is safe but may break monitoring.
  Confirm the token is injected before enabling the ops route.

- [ ] **No token values appear in the `/api/status` response.**
  `/api/status` surfaces deployment labels and policy flags, not secret values.
  Verify before enabling the endpoint in a publicly reachable environment.

---

## Rate limiting

- [ ] **Abuse guard rate and burst are set for your traffic profile.**
  `cfg.App.RateLimit` and `cfg.App.RateBurst` default to permissive values for
  local development. Set them via `APP_RATE_LIMIT` and `APP_RATE_BURST` before
  production. The `AbuseGuardMiddleware` token-bucket applies per process; add
  a shared backend (e.g., `x/resilience/ratelimit`) if you run multiple replicas.

- [ ] **`RateLimit.Stop()` is called on shutdown.**
  `App.Start` calls `defer a.RateLimit.Stop()`. Verify no custom shutdown logic
  bypasses this defer.

---

## Tenant resolution

- [ ] **Tenant ID header is correct for your environment.**
  The profile route reads `X-Tenant-ID` via `x/tenant/resolve`. Confirm the
  header name matches what your API gateway or client injects, and that missing
  or empty tenant IDs are handled by the fail-closed policy.

- [ ] **Profile store path is set for persistent storage.**
  `APP_PROFILE_STORE_PATH` defaults to an in-memory store. For durable storage
  use the JSON file store or replace `App.Profiles` with a repository backed by
  your persistence layer. The in-memory store resets on restart.

---

## Tracing

- [ ] **Tracer is replaced with a real implementation.**
  `app.go` wires a `noopTracer` by default. For distributed tracing, replace
  `noopTracer{}` with a real `tracing.Tracer` backed by OpenTelemetry or your
  preferred provider before production.

---

## Observability

- [ ] **HTTP metrics collection is confirmed.**
  `httpmetrics.Middleware` records per-route request stats into
  `metrics.BaseMetricsCollector`. Verify that `GET /ops/metrics` surfaces the
  expected counters and that your monitoring system scrapes or reads them.

- [ ] **`x/observability/devtools` is not mounted.**
  This service explicitly excludes devtools routes. Confirm no `/_debug/*`
  routes were added to `routes.go` — debug routes must never reach production.

---

## Inherited from standard-service

All items from `reference/standard-service/PRODUCTION_CHECKLIST.md` apply:
transport timeouts, body size limits, CORS policy, security headers, TLS, signal
handling, and test coverage.
