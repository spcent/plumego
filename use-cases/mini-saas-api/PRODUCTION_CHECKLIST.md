# Production Checklist — mini-saas-api

Before promoting this multi-tenant SaaS API to a production environment, work through this
list. Each item is a conscious decision and a security or reliability boundary that affects
how the service behaves under real traffic.

---

## Secrets and Credentials

- [ ] **`APP_JWT_SECRET` is set to at least 32 cryptographically random bytes.**
  Generate with: `openssl rand -hex 16` (produces 32 hex chars = 16 bytes; use -hex 32 for 32 bytes).
  This secret signs all JWT access tokens (HS256) and refresh token families.
  **Do not reuse across environments.** Do not hardcode in code or config files.
  Store in your platform's secret management (AWS Secrets Manager, Vault, GitHub Secrets, etc.).
  The server logs an error and refuses to start if the secret is shorter than 32 chars.

- [ ] **Refresh tokens are hashed in the embedded KV store.**
  The store is file-backed (`.data/` directory). In production, this directory must be:
  - Persisted across server restarts (mount a volume if running in containers)
  - Readable only by the application user (file mode 0600)
  - Backed up as part of your disaster recovery plan (contains active refresh token families)
  - Rotated periodically; old entries are not automatically purged

- [ ] **No secrets appear in logs.**
  The service filters these fields from structured logs:
  - `APP_JWT_SECRET` value (never logged)
  - Password hashes (never logged, only stored)
  - Refresh token values (never logged; token IDs only)
  - `Idempotency-Key` header (never logged)
  Verify by checking access logs and error responses — no secrets should be present.

---

## Authentication and Authorization

- [ ] **Bearer token authentication is enforced on all protected endpoints.**
  Routes under `/api/v1/*` require an `Authorization: Bearer <JWT>` header.
  Public endpoints (signup, login, refresh, health, readiness, metrics) do not require auth.
  Verify by testing: unauthenticated requests to protected routes return 401 Unauthorized.

- [ ] **JWT tokens are validated and expire appropriately.**
  - Access tokens expire after `APP_JWT_ACCESS_TTL` (default 15 minutes)
  - Refresh tokens expire after `APP_JWT_REFRESH_TTL` (default 168 hours = 7 days)
  - Expired access tokens trigger 401; client must refresh.
  - Expired refresh tokens trigger 401; client must log in again.
  Test: request with an expired token should return 401, not 403 or 200.

- [ ] **Refresh token rotation is single-use and family revocation is enforced.**
  - Each refresh token is issued once and becomes invalid after use.
  - If the same refresh token is reused (token theft), the entire token family is revoked.
  - Client must log in again after family revocation.
  Test: refresh once (succeeds), refresh with the old token (401 + "reuse detected").

- [ ] **RBAC is enforced: owner > admin > member.**
  - Admin endpoints check `principal.Role >= admin` in handlers.
  - Last-owner invariant prevents removing the last owner from a workspace.
  - Cross-tenant lookups return 404 (not 403) to avoid existence leaks.
  Test: member attempts to delete a member (fails); owner succeeds.

- [ ] **Password strength is enforced on signup and account creation.**
  Default: ≥8 characters, mixed case, at least one digit.
  Test: signup with weak password (e.g., "password") returns 400.

---

## Transport and Network

- [ ] **TLS is enabled or the service runs behind a TLS-terminating proxy.**
  If self-terminating:
  - Set `APP_TLS_ENABLED=true`
  - Provide `APP_TLS_CERT_FILE` and `APP_TLS_KEY_FILE` paths
  - Verify certificate is valid for your domain
  If behind a proxy (e.g., Caddy, Nginx, load balancer):
  - Ensure `X-Forwarded-Proto`, `X-Forwarded-For`, `X-Forwarded-Host` headers are passed
  - Set `core.Config.ProxyHeaders = true` if the framework supports it
  Test: `curl https://api.example.com/healthz` returns 200.

- [ ] **Read, write, and idle timeouts are configured.**
  Set in `internal/config/config.go`:
  - `cfg.Core.ReadTimeout` (default 10s) — time to read the complete request
  - `cfg.Core.WriteTimeout` (default 10s) — time to write the complete response
  - `cfg.Core.IdleTimeout` (default 120s) — time to close idle connections
  For a typical SaaS API, 10s read/write and 120s idle are reasonable.
  Increase if your bulk operations exceed these windows.

- [ ] **Per-request timeout is set and appropriate for your workload.**
  `timeout.Middleware` is configured in `internal/app/app.go` (default 30s).
  This is a wall-clock limit for the entire request (not per operation).
  Set it to your p99 latency budget. Test: a slow handler that sleeps > timeout returns 408 Request Timeout.

- [ ] **Request body size limit is configured for your expected traffic.**
  Default: 1 MiB (`APP_MAX_BODY_BYTES=1048576`).
  Adjust if you accept file uploads or large bulk operations.
  Test: POST a body larger than the limit; expect 413 Payload Too Large.

---

## CORS and Access Control

- [ ] **CORS origins are explicitly allowed.**
  `APP_CORS_ALLOWED_ORIGINS` defaults to empty (allows all, wildcard `*`).
  **Must be set in production** to a comma-separated list:
  ```
  APP_CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
  ```
  Do not use wildcard origins for authenticated APIs.
  The service logs a WARN at startup when this is empty, making the risk visible.
  Test: a request from `https://untrusted.example.com` should be rejected (no CORS headers in response).

- [ ] **Security headers are in place.**
  `middleware/securityheaders` is wired in `internal/app/app.go` and sets:
  - `X-Frame-Options: DENY` (no framing)
  - `X-Content-Type-Options: nosniff` (no MIME sniffing)
  - `Referrer-Policy: strict-origin-when-cross-origin`
  Verify with: `curl -i https://api.example.com/healthz | grep X-`

---

## Rate Limiting and Abuse Guards

- [ ] **Authentication endpoints (signup, login, refresh) are guarded against brute-force.**
  `middleware/abuseguard` is wired on the `/auth/*` routes.
  Default: 10 requests per minute per IP; burst up to 20.
  A client exceeding the limit receives 429 Too Many Requests.
  Configure in `internal/app/routes.go` if needed; token bucket resets every minute.
  Test: 11 rapid requests to `/auth/login` from the same IP → 429 on the 11th.

- [ ] **Per-tenant rate limiting is configured and appropriate for your scale.**
  All authenticated routes enforce:
  - Token-bucket rate limit: `APP_TENANT_RPS` requests/sec (default 50), burst `APP_TENANT_BURST` (default 100)
  - Fixed-window quota: `APP_TENANT_QUOTA_PER_MINUTE` (default 600 requests/min)
  The limits are uniform across all tenants in this version.
  A tenant exceeding the limit receives 429 Too Many Requests.
  Plan for:
    - Spike handling: set burst higher than sustained RPS if spikes are expected
    - Fair sharing: if you have N tenants, divide the total capacity by N
    - Dynamic limits: consider implementing a tier-based system (plan-based RPS/quota) in a future card
  Test: make >100 simultaneous requests from one tenant; expect 429 on requests after burst.

- [ ] **Observability is configured to detect rate-limit and quota violations.**
  Prometheus metrics are exposed on GET `/metrics`:
  - `http_requests_total` counter includes status codes (429 for rate/quota limits)
  - `http_request_duration_seconds` histogram captures latency
  Configure your monitoring to alert when 429 status code percentage exceeds a threshold.

---

## Observability and Monitoring

- [ ] **Request ID propagation is in place.**
  Every request is stamped with a unique ID via `middleware/requestid`.
  The ID appears in:
  - Access logs (structured `request_id` field)
  - Error responses (`"request_id": "..."` in JSON)
  - Correlation headers (if downstream services read `X-Request-ID`)
  Use the request ID to trace a single user's requests through your logs.

- [ ] **Structured access logging is enabled and shipped to your aggregator.**
  `middleware/accesslog` logs every request/response as structured JSON:
  ```json
  {
    "timestamp": "2026-06-14T12:00:00Z",
    "request_id": "01HX...",
    "method": "POST",
    "path": "/api/v1/projects",
    "status": 201,
    "duration_ms": 45,
    "client_ip": "192.0.2.1"
  }
  ```
  Configure your log forwarder (Datadog, Splunk, ELK) to consume these logs.

- [ ] **Health probes are wired and monitored.**
  Two endpoints:
  - `GET /healthz` (liveness) — always returns 200 if the process is running. No dependency checks.
  - `GET /readyz` (readiness) — returns 200 + component status if all probes pass; 503 if any fail.
  Configure your load balancer / Kubernetes to:
    - Probe `/healthz` every 10-15 seconds (restart if no response)
    - Probe `/readyz` every 5-10 seconds (drain traffic if fails)

- [ ] **Prometheus metrics are collected and exported.**
  Metrics are exposed on `GET /metrics` (Prometheus text format).
  Metrics include:
  - `http_requests_total{method,path,status}` — request count
  - `http_request_duration_seconds{method,path}` — latency histogram
  Configure Prometheus or your metrics backend to scrape `/metrics` every 15-60 seconds.
  Optional: guard `/metrics` with a bearer token by setting `APP_METRICS_TOKEN=<secret>`.

- [ ] **Metrics token is set if `/metrics` is publicly accessible.**
  Default: empty (no guard). If your `/metrics` endpoint is exposed to the internet,
  set `APP_METRICS_TOKEN` to a strong random string:
  ```
  APP_METRICS_TOKEN=$(openssl rand -hex 16)
  ```
  Then require: `curl -H "Authorization: Bearer $APP_METRICS_TOKEN" https://api.example.com/metrics`
  Test: unauthorized requests to `/metrics` (without token) return 401.

- [ ] **Audit trail is durable and queryable.**
  All tenant/project mutations are appended to an in-memory audit log.
  In this version, the log is not persisted; it is lost on server restart.
  For production, implement a durable audit store:
  1. Create an `AuditStore` interface similar to the domain repositories
  2. Implement it with your database backend
  3. Wire it in `internal/app/app.go` and inject into domain services
  Until then, regularly export audit events to an external service for compliance.

---

## Data Persistence and Storage

- [ ] **All domain stores are implemented with a durable backend.**
  This service ships with in-memory repositories:
  - `internal/domain/user/store.go` (MemoryStore)
  - `internal/domain/tenantspace/store.go` (MemoryStore)
  - `internal/domain/project/store.go` (MemoryStore)
  - `internal/domain/audit/audit.go` (in-memory buffer, max 1000 entries)
  For production:
  1. Design your database schema (users, tenants, memberships, projects, audit_events)
  2. Implement `Repository` interfaces for each domain with SQL/NoSQL driver
  3. Wire durable stores in `internal/app/app.go`
  4. Add connection pooling, retry logic, and timeouts
  5. Test: all domain tests should pass against your durable backend
  The handler and business logic layers don't change — only the storage backend.

- [ ] **Embedded KV store (`.data/`) is persisted and backed up.**
  The `.data/` directory holds:
  - JWT signing keys (bootstrap on first run)
  - Refresh token hashes and family trees
  This is the only persisted state in the current service.
  Ensure:
  - `.data/` is on a volume that survives container restarts
  - `.data/` is excluded from version control (add to `.gitignore`)
  - `.data/` is included in your backup/disaster recovery plan
  - File permissions: 0600 (readable only by the application user)
  Consider: if you implement a database backend, migrate this to your SQL/NoSQL store as well.

- [ ] **Database schema migrations are planned.**
  When you implement durable stores, plan for:
  - Initial schema creation on app startup
  - Version control for schema migrations (e.g., separate `migrations/` directory)
  - Backward-compatible schema changes (rolling updates, old+new fields coexist temporarily)
  - Test: schema migration runs without data loss on a replica before applying to production

---

## Testing and Validation

- [ ] **All tests pass with the race detector.**
  ```bash
  cd use-cases/mini-saas-api && go test -race -timeout 60s ./...
  ```
  No goroutine data races, no concurrent access bugs.

- [ ] **Manual smoke test passes.**
  ```bash
  go run . &
  sleep 2
  curl http://localhost:8090/healthz  # 200 OK
  curl http://localhost:8090/readyz   # 200 OK
  bash api/curl.sh                    # end-to-end walkthrough
  ```

- [ ] **Boundary checks pass.**
  ```bash
  go run ./internal/checks/dependency-rules    # stable roots only, no x/* in main module
  go run ./internal/checks/reference-layout    # directory shape matches canonical
  go run ./internal/checks/module-manifests    # go.mod and AGENTS.md present
  ```

---

## Security Checklist

- [ ] **Tenant isolation is verified at every layer.**
  - Handlers extract tenant ID from JWT claim
  - x/tenant middleware resolves and injects tenant context
  - Every repository method accepts explicit `tenantID` parameter
  - Cross-tenant queries return 404 (not 403)
  Test: create projects in two different tenants; user A cannot see user B's projects.

- [ ] **RBAC enforcement is visible and tested.**
  - Member cannot promote themselves to admin
  - Admin cannot demote the last owner
  - Owner can do everything
  Test: member creates project (succeeds); member deletes project (fails with 403).

- [ ] **Idempotency-Key enforcement prevents duplicate mutations.**
  Mutating routes (`POST`, `PUT`, `PATCH`, `DELETE`) accept `Idempotency-Key` header.
  - First request with key K: execute normally, cache response
  - Second request with key K and same body: replay cached response (look for `X-Idempotency-Replay: true`)
  - Second request with key K and different body: return 409 Conflict
  Test: create a project twice with the same key and body → same project ID both times.

---

## Deployment Checklist

- [ ] **Container image is built with multi-stage Dockerfile (if containerized).**
  Build stage: compile Go binary with build flags (`-X main.version=...`).
  Runtime stage: minimal base image (alpine:3.19 or distroless) with only the binary.
  Example:
  ```dockerfile
  FROM golang:1.26 as builder
  WORKDIR /src
  COPY . .
  RUN go build -o app -ldflags "-X main.version=v1.0.0" ./cmd/mini-saas-api

  FROM alpine:3.19
  RUN apk add --no-cache ca-certificates
  COPY --from=builder /src/app /app
  EXPOSE 8090
  ENTRYPOINT ["/app"]
  ```

- [ ] **Environment variables are injected, not baked into the image.**
  All configuration must come from the runtime environment:
  - `APP_JWT_SECRET` from secrets
  - `APP_CORS_ALLOWED_ORIGINS` from config
  - `APP_ADDR`, timeouts, limits from deployment config
  Never hardcode credentials or config in the image.

- [ ] **Health probes are configured in your orchestration platform.**
  Kubernetes example:
  ```yaml
  livenessProbe:
    httpGet:
      path: /healthz
      port: 8090
    initialDelaySeconds: 5
    periodSeconds: 10
  readinessProbe:
    httpGet:
      path: /readyz
      port: 8090
    initialDelaySeconds: 5
    periodSeconds: 5
  ```

- [ ] **Graceful shutdown is configured (container/orchestration level).**
  The service listens for `SIGTERM` and allows in-flight requests 15 seconds to complete.
  Configure:
  - Docker: `STOPSIGNAL SIGTERM`, `stop_grace_period: 30s` in Compose
  - Kubernetes: `terminationGracePeriodSeconds: 30` in Pod spec
  Test: send SIGTERM, confirm that in-flight requests complete without 500 errors.

---

## Going to Production

Read `README.md` for:
- Full route table and authentication requirements
- Configuration table (all `APP_*` variables)
- How to extend the service (add endpoints, swap storage)

Read `ARCHITECTURE.md` for:
- Ownership and dependency direction
- Request flow through middleware and handlers
- Security boundaries and tenant isolation
- Module usage (stable roots + beta extensions)

For further hardening, see `reference/production-service/` which adds:
- Bearer-token authentication (global, fail-closed)
- Per-IP abuse guards (global)
- Distributed tracing (x/observability)
- Protected ops metrics and health routes (x/observability/ops)
