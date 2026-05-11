# Production Checklist — with-gateway

Run through this list in addition to `reference/standard-service/PRODUCTION_CHECKLIST.md`
before deploying a gateway service.

---

## Gateway-specific

- [ ] **Backend TLS verification is enabled**.
  This demo does not configure TLS options on the upstream connection. In
  production, ensure `GatewayConfig` sets appropriate TLS settings and that
  the backend's certificate chain is trusted.

- [ ] **Per-upstream timeouts are configured**.
  Without explicit dial, response-header, and idle timeouts on the upstream
  transport, a slow or unresponsive backend will hold connections open
  indefinitely, exhausting the gateway's connection pool.

- [ ] **Circuit breaking is configured**.
  The demo uses a single target with no retry or circuit-break policy. In
  production, configure retry limits and failure thresholds to avoid cascading
  failures when a backend becomes unhealthy.

- [ ] **Request ID is forwarded to the backend**.
  `middleware/requestid` generates an ID for each inbound request. Configure
  the proxy to forward this ID (e.g., as `X-Request-ID`) so that backend logs
  can be correlated with gateway logs.

- [ ] **Backend health is checked before routing**.
  The demo routes all `/proxy/*` traffic to the configured backend without
  checking liveness. In production, implement an active health check and
  remove unhealthy backends from the target list.

- [ ] **Path prefix stripping is intentional**.
  The wildcard route `/proxy/*` forwards the full path including the `/proxy`
  prefix unless the proxy config strips it. Confirm the backend expects the
  path as forwarded.

- [ ] **Hop-by-hop headers are handled**.
  Reverse proxies must strip or rewrite `Connection`, `Upgrade`, and related
  hop-by-hop headers. Confirm `x/gateway` handles these for your protocol
  (HTTP/1.1, HTTP/2, WebSocket).

---

## Inherited from standard-service

All items from `reference/standard-service/PRODUCTION_CHECKLIST.md` apply:
transport timeouts, body size limits, security headers, TLS, observability,
signal handling.
