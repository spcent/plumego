# AGENTS.md - use-cases/guardus

Operational guide for AI agents working on the guardus use-case.

## 1. Purpose

`guardus` is a feature-equivalent v1 reimplementation of [Gatus](https://github.com/TwiN/gatus)
backend, built on plumego stable roots and `x/*` extensions. It exists alongside
`use-cases/gatus` (the original Gatus drop-in) to validate plumego in
"monitoring/observability product" use cases.

The SPA is rewritten in React + Vite (not Vue). Source lives in `web/app/`;
`make frontend-build` produces `web/static/` which is embedded by the Go binary.
Suites and OIDC are out of v1 scope and intentionally omitted.

## 2. Scope (v1)

In:

- Endpoint monitoring: HTTP, TCP, ICMP, DNS, TLS/STARTTLS, gRPC, WebSocket
- ExternalEndpoint (push API)
- SQLite + memory storage
- Basic Auth
- Alert providers: webhook, email, slack, telegram, discord
- Maintenance windows + Connectivity checker
- Prometheus metrics, /health, embedded SPA

Out (v2 / later):

- Suites, OIDC, SSH/SCTP/UDP probes
- SSH tunneling, Remote federation, Announcements, Custom CSS
- PostgreSQL (v1.1)

## 3. Hard rules

- Module-local: this use-case has its own `go.mod` with `replace` to `../..`.
  All non-stdlib deps stay in this `go.mod`; the root plumego module never
  gains dependencies because of this work.
- HTTP handlers MUST use `func(http.ResponseWriter, *http.Request)`.
- **Wire-format compatibility with upstream gatus is required** so the
  SPA works unmodified. Upstream gatus emits naked JSON
  (no `{data:..., meta:...}` envelope) and plain-text error bodies; guardus
  matches that on the wire. Use the helpers in `internal/handler/util.go`:
  - `WriteRawJSON(w, status, body)` for JSON successes
  - `WriteText` / `WriteSVG` for non-JSON content (badges, charts, raw)
  - `WriteErrorString(w, r, status, msg)` for plain-text error bodies
- Do **not** use `contract.WriteResponse` / `contract.WriteError` here —
  that envelope would break the SPA. This is a documented deviation from
  the rest of the plumego repo.
- No reflection-based provider lookup and no `init()` registration. Alert
  providers are dispatched through a typed-field switch on `alerting.Config`
  (see `internal/alerting/config.go`); each supported provider is a named
  field on the struct, and `GetAlertingProviderByAlertType` resolves by
  `alert.Type`. This is a deliberate hardening over upstream gatus's
  reflection-based lookup — adding a new provider means adding a field and a
  switch case, nothing implicit.
- No package-level singletons (no `store.Get()`-style globals). Storage,
  watchdog and config are constructed in `app.New` and injected.
- Path params: `router.Param(r, "key")`, never `c.Params(...)`.
- **No YAML.** Deployment config is sourced from `GUARDUS_*` env vars
  (`config.LoadFromEnv`); endpoint definitions live in the storage backend
  (SQLite or memory) and are read at boot via `store.ListEndpointConfigs`.
  An optional JSON file at `GUARDUS_BOOTSTRAP_FILE` is upserted into the
  store on startup. There is no hot reload — admin-API-driven watchdog
  rebuilds are the planned next milestone. Serialization is `encoding/json`
  only; do not introduce `gopkg.in/yaml.v3` or any other YAML dependency.

## 4. Fiber → net/http migration cheat-sheet

| Fiber                                        | net/http                                                               |
|----------------------------------------------|------------------------------------------------------------------------|
| `c.Params("key")`                            | `router.Param(r, "key")`                                               |
| `c.Query("page")`                            | `r.URL.Query().Get("page")`                                            |
| `c.Get("X-Foo")`                             | `r.Header.Get("X-Foo")`                                                |
| `c.Status(404).SendString(msg)`              | `handler.WriteErrorString(w, r, 404, msg)`                             |
| `c.JSON(v)`                                  | `body, _ := json.Marshal(v); handler.WriteRawJSON(w, 200, body)`       |
| `c.Send(svgBytes)` (badge / chart)           | `w.Header().Set("Content-Type", "image/svg+xml"); _, _ = w.Write(svg)` |
| `app.Get("/x/:k", h)`                        | `routes := newRouteReg(a.Core); routes.get("/x/:k", http.HandlerFunc(h))` |

## 5. Layout

```
internal/
  app/         App, New(cfg), RegisterRoutes(), Start(ctx)
  config/      Env-var loader (LoadFromEnv), validation, JSON bootstrap import
  domain/      Pure domain types (no plumego imports)
    alert/   endpoint/   pattern/   jsonpath/
  client/      HTTP/TCP/ICMP/DNS/TLS/gRPC/WS probe helpers
  storage/     Store interface + memory + sql/sqlite
  watchdog/    Per-endpoint goroutines + alerting state machine
  alerting/    Config + provider registry + 5 providers
  security/    Basic auth + bcrypt/PBKDF2 verifier + sessions placeholder
  handler/     HTTP adapters
  metrics/     Prometheus collector wiring
web/
  app/         React + Vite SPA source (Node.js)
  static/      Built SPA output (embedded by Go via //go:embed)
  static.go    Go embed wrapper
```

## 6. Validation

Per phase: `make test`. End-to-end: `make e2e`. Frontend: `make frontend-build`.

From repo root, before merging significant changes:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/cross-extension-deps
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
make gates
```

The use-case's own deps must NOT leak into the root `go.mod` — verify with
`grep modernc /go.mod` from the repo root (must return nothing).

## 7. Commits

State that the change was made by an agent and include the model name when
opening a commit or PR.
