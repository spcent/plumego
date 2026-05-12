# Production Checklist — with-ops

Run through this list before deploying a service with `x/ops` routes.

---

## Ops-specific

- [ ] **Ops token is a strong secret**.
  This demo uses `local-admin-token` as the default. In production, generate a
  cryptographically random token (at least 32 bytes) and supply it via a secret
  manager, not a plain environment variable in a checked-in file.

- [ ] **Ops routes are network-restricted**.
  `/ops/*` routes should not be reachable from the public internet. Restrict
  access at the network layer (firewall rule, ingress policy, VPN requirement)
  in addition to the token check.

- [ ] **`ops.Options.Enabled` is controlled by environment**.
  Do not hard-code `Enabled: true`. Gate it on an environment variable so the
  ops surface can be disabled in environments where it is not needed.

- [ ] **`QueueStats` hook is implemented against a real queue**.
  The demo returns synthetic queue statistics. In production, implement the hook
  against your actual queue backend (Redis, SQS, RabbitMQ, etc.).

- [ ] **Graceful shutdown is wired**.
  This demo calls `http.ListenAndServe` directly. In production, use
  `http.Server.Shutdown(ctx)` with a context that responds to `SIGTERM`
  and `SIGINT` to drain in-flight requests before the process exits.

- [ ] **Metrics endpoint is access-controlled**.
  `/metrics` exposes internal runtime counters. Restrict it to internal networks
  or add the same token check used by `/ops/*` routes.

---

## Inherited from standard-service

All items from `reference/standard-service/PRODUCTION_CHECKLIST.md` apply:
transport timeouts, body size limits, security headers, TLS, observability,
signal handling.

Note: this demo does not use `core.App`. Apply the transport hardening items
directly on the `http.Server` struct before calling `ListenAndServe`.
