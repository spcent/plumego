# Production Checklist — standard-service

Run through this list before promoting the service to a production environment.
Each item is a conscious decision, not a default assumption.

---

## Transport hardening

- [ ] **Read and write timeouts** are set.
  `cfg.Core.ReadTimeout`, `cfg.Core.WriteTimeout`, `cfg.Core.IdleTimeout`.
  Defaults are permissive for local development. Set explicit values for production.

- [ ] **Body size limit** middleware is in the middleware stack.
  Add `middleware.BodyLimit(maxBytes)` before handlers that accept request bodies.
  The default unbounded body allows memory exhaustion attacks.

- [ ] **Request timeout** middleware is configured.
  Add `middleware.Timeout(duration)` to enforce a per-request wall-clock limit.

- [ ] **TLS is enabled** or the service runs behind a TLS-terminating proxy.
  If self-terminating, set `cfg.Core.TLS.Enabled = true` and provide cert/key paths.

- [ ] **CORS policy** is defined if the API is consumed by browser clients.
  Add `middleware.CORS(config)` with explicit allowed origins. Do not use wildcard
  origins for authenticated APIs.

---

## Security

- [ ] **Security headers** middleware is in the stack.
  Add `middleware.SecureHeaders()` to set `X-Content-Type-Options`,
  `X-Frame-Options`, `Strict-Transport-Security`, and related headers.

- [ ] **Rate limiting** is configured for public endpoints.
  Add `middleware.RateLimit(config)` to protect high-traffic or unauthenticated routes.

- [ ] **Authentication** is enforced on all non-public endpoints.
  There is no default authentication. Every protected route must have an explicit
  auth middleware or handler guard.

- [ ] **No debug endpoints are mounted** by default.
  `x/observability/devtools` is explicitly excluded from this service. Verify that no debug
  routes have been added to `routes.go`.

---

## Observability

- [ ] **Structured access logging** is in the middleware stack.
  `middleware/accesslog` is wired in `app.go`. Confirm it is logging to the
  intended destination in your environment.

- [ ] **Request ID propagation** is in the middleware stack.
  `middleware/requestid` is wired in `app.go`. Verify that downstream services
  and log aggregators can correlate requests by ID.

- [ ] **Health endpoints** are reachable.
  `/healthz` (liveness) and `/readyz` (readiness) are registered in `routes.go`.
  Confirm your orchestration layer (Kubernetes, load balancer) is configured to
  probe them.

- [ ] **Metrics** are collected if required.
  This service does not mount a metrics endpoint by default. Add `x/observability`
  and an explicit `/metrics` route if your environment requires it. See
  `reference/with-ops` for an example.

---

## Lifecycle

- [ ] **Graceful shutdown with signal handling** is wired.
  `main.run` creates a `signal.NotifyContext(SIGTERM, SIGINT)` and passes it to
  `app.Start(ctx)`, which triggers `app.Core.Shutdown` when the context is
  canceled. Verify that the shutdown timeout is long enough for the slowest
  expected request in your environment.

- [ ] **Configuration is loaded from the environment**, not from checked-in files.
  `.env` files are for local development only. Production environments should
  supply configuration through environment variables injected by the platform.

---

## Testing

- [ ] **All handler tests pass** with `go test -race ./...`.
- [ ] **Integration smoke test**: the service starts, `/healthz` returns 200,
  `/readyz` returns 200.
- [ ] **Boundary checks pass**: `go run ./internal/checks/dependency-rules`.
  This service must not import `x/*`.
