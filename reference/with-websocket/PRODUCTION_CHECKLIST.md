# Production Checklist — with-websocket

Run through this list in addition to `reference/standard-service/PRODUCTION_CHECKLIST.md`
before deploying a WebSocket service.

---

## WebSocket-specific

- [ ] **Authentication is enabled** on the WebSocket upgrade route.
  This demo sets `AllowUnauthenticated = true` for simplicity. In production,
  set `AllowUnauthenticated = false` and provide a JWT secret or a custom
  token validator.

- [ ] **Allowed origins are restricted**.
  This demo sets `AllowedOrigins: []string{"*"}`. In production, list only the
  origins your client applications are served from.

- [ ] **Hub capacity limits** are configured.
  `x/websocket` supports per-hub and per-room connection limits. Unbounded hubs
  allow resource exhaustion under high load.

- [ ] **WebSocket shutdown timeout** is tuned.
  `a.WS.Shutdown(ctx)` drains active connections. The context deadline should be
  long enough for clients to receive a close frame and disconnect cleanly, but
  short enough to not block the process from exiting.

- [ ] **Ping/pong keepalives** are configured if connections are long-lived.
  Without keepalives, idle connections may be dropped by intermediate proxies
  without the server or client being notified.

- [ ] **Reverse proxy is configured for WebSocket**.
  If running behind a load balancer or ingress, confirm that the proxy forwards
  `Upgrade` and `Connection` headers and does not buffer WebSocket frames.

---

## Inherited from standard-service

All items from `reference/standard-service/PRODUCTION_CHECKLIST.md` apply:
transport timeouts, body size limits, security headers, TLS, observability,
signal handling.

The WebSocket upgrade route is exempt from read/write timeout middleware — the
HTTP server timeout does not apply after the connection is upgraded. Ensure
application-level keepalive and idle timeout are configured directly on the
WebSocket server configuration.
