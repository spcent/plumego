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
- SQLite + MySQL + memory storage (dialect-based abstraction)
- Basic Auth
- Alert providers: webhook, email, slack, telegram, discord
- Maintenance windows + Connectivity checker
- Prometheus metrics, /health, embedded SPA with lazy-loaded routes

Out (v2 / later):

- Suites, OIDC, SSH/SCTP/UDP probes
- SSH tunneling, Remote federation, Announcements, Custom CSS
- PostgreSQL (v2.0)

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
  (SQLite, MySQL, or memory) and are read at boot via `store.ListEndpointConfigs`.
  An optional JSON file at `GUARDUS_BOOTSTRAP_FILE` is upserted into the
  store on startup. There is no hot reload — admin-API-driven watchdog
  rebuilds are the planned next milestone. Serialization is `encoding/json`
  only; do not introduce `gopkg.in/yaml.v3` or any other YAML dependency.

## 4. Storage architecture

The storage layer uses a **dialect abstraction** to support multiple SQL backends
without duplicating business logic. Key files:

- `internal/storage/storage.go` — `Store` interface + factory (`New`)
- `internal/storage/sql/dialect.go` — `dialect` interface + SQLite/MySQL implementations
- `internal/storage/sql/sql.go` — Backend-agnostic SQL logic using dialect methods
- `internal/storage/sql/schema_sqlite.sql` — SQLite DDL (embedded)
- `internal/storage/sql/schema_mysql.sql` — MySQL DDL (embedded)

### Dialect interface

```go
type dialect interface {
    placeholder(n int) string                    // "$1" vs "?"
    placeholders(from, count int) string         // "$1,$2,$3" vs "?,?,?"
    returningID(column string) string            // "RETURNING id" vs ""
    upsertClause(conflictCols string, cols []string) string  // "ON CONFLICT" vs "ON DUPLICATE KEY"
    initPragmas(db *sql.DB) error                // PRAGMA vs no-op
}
```

### Adding a new backend

1. Implement `dialect` interface (e.g., `postgresDialect`)
2. Create `schema_postgres.sql`
3. Add driver import to `sql.go` (e.g., `_ "github.com/lib/pq"`)
4. Add `TypePostgres` constant to `internal/storage/type.go`
5. Update `NewStore` switch in `sql.go` to select dialect + schema
6. Update factory in `storage.go` to handle new type
7. Update config validation in `internal/storage/config.go`

### MySQL-specific notes

- MySQL doesn't support `RETURNING` — use `LastInsertId()` for INSERT
- MySQL uses `?` placeholders (not `$N`)
- MySQL `ON DUPLICATE KEY UPDATE` uses `VALUES(col)` to reference proposed values
- MySQL connection pooling: `MaxOpenConns=10` (no single-writer constraint like SQLite)
- DSN format: `user:pass@tcp(host:port)/dbname?parseTime=true&charset=utf8mb4`

### Testing storage backends

SQLite and memory tests run by default. MySQL tests are opt-in via environment variable:

```bash
# Requires running MySQL with guardus_test database
GUARDUS_TEST_MYSQL_DSN="guardus:test@tcp(127.0.0.1:3306)/guardus_test?parseTime=true" \
  go test ./internal/storage -v
```

Quick MySQL setup with Docker:

```bash
docker run -d --name guardus-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=guardus_test \
  -e MYSQL_USER=guardus \
  -e MYSQL_PASSWORD=test \
  -p 3306:3306 mysql:8
```

## 5. Frontend architecture

The React SPA uses **route-level code splitting** to minimize initial bundle size:

- `web/app/src/App.jsx` — `React.lazy()` + `Suspense` for Home/EndpointDetails
- `web/app/vite.config.js` — Manual chunks: vendor (React/Router), charts (Chart.js), icons (lucide-react)
- Initial load: ~217 KB (down from 663 KB single bundle)
- Lazy-loaded: Home (164 KB), EndpointDetails (13 KB), Charts (261 KB)

### Build & dev

```bash
make frontend-build    # npm ci + vite build → web/static/
cd web/app && npm run dev  # Dev server with /api proxy to :8080
```

## 6. Fiber → net/http migration cheat-sheet

| Fiber                                        | net/http                                                               |
|----------------------------------------------|------------------------------------------------------------------------|
| `c.Params("key")`                            | `router.Param(r, "key")`                                               |
| `c.Query("page")`                            | `r.URL.Query().Get("page")`                                            |
| `c.Get("X-Foo")`                             | `r.Header.Get("X-Foo")`                                                |
| `c.Status(404).SendString(msg)`              | `handler.WriteErrorString(w, r, 404, msg)`                             |
| `c.JSON(v)`                                  | `body, _ := json.Marshal(v); handler.WriteRawJSON(w, 200, body)`       |
| `c.Send(svgBytes)` (badge / chart)           | `w.Header().Set("Content-Type", "image/svg+xml"); _, _ = w.Write(svg)` |
| `app.Get("/x/:k", h)`                        | `routes := newRouteReg(a.Core); routes.get("/x/:k", http.HandlerFunc(h))` |

## 7. Layout

```
internal/
  app/         App, New(cfg), RegisterRoutes(), Start(ctx)
  config/      Env-var loader (LoadFromEnv), validation, JSON bootstrap import
  domain/      Pure domain types (no plumego imports)
    alert/   endpoint/   pattern/   jsonpath/   connectivity/   maintenance/
  client/      HTTP/TCP/ICMP/DNS/TLS/gRPC/WS probe helpers
  storage/     Store interface + memory + sql (dialect-based)
    memory/    In-memory backend
    sql/       SQL backends (SQLite, MySQL) with dialect abstraction
      dialect.go          Dialect interface + implementations
      schema_sqlite.sql   SQLite DDL
      schema_mysql.sql    MySQL DDL
  watchdog/    Per-endpoint goroutines + alerting state machine
  alerting/    Config + provider registry + 5 providers
  security/    Basic auth + bcrypt/PBKDF2 verifier + sessions placeholder
  handler/     HTTP adapters
  metrics/     Prometheus collector wiring
web/
  app/         React + Vite SPA source (Node.js)
    src/
      App.jsx          Route definitions with lazy loading
      views/           Home, EndpointDetails (lazy-loaded)
      components/      Reusable UI components
    vite.config.js     Build config with manual chunk splitting
  static/      Built SPA output (embedded by Go via //go:embed)
  static.go    Go embed wrapper
```

## 8. Validation

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

### Storage-specific validation

```bash
# SQLite + memory (default)
go test ./internal/storage -v

# Add MySQL (opt-in)
GUARDUS_TEST_MYSQL_DSN="..." go test ./internal/storage -v

# Full suite
go test ./... -race
```

## 9. Commits

State that the change was made by an agent and include the model name when
opening a commit or PR.

## 10. Configuration reference

All runtime config via `GUARDUS_*` environment variables. See `env.example` for
the complete list. Key storage settings:

```bash
# SQLite (default)
GUARDUS_STORAGE_TYPE=sqlite
GUARDUS_STORAGE_PATH=./guardus.db

# MySQL
GUARDUS_STORAGE_TYPE=mysql
GUARDUS_STORAGE_PATH="guardus:pass@tcp(127.0.0.1:3306)/guardus?parseTime=true"

# Memory (no persistence)
GUARDUS_STORAGE_TYPE=memory
```

## 11. Common tasks

### Add a new probe type

1. Implement probe logic in `internal/client/` (e.g., `client.NewHTTPProbe()`)
2. Add config fields to `internal/domain/endpoint/endpoint.go`
3. Add validation in `ValidateAndSetDefaults()`
4. Wire into watchdog in `internal/watchdog/endpoint.go`
5. Update frontend form in `web/app/src/components/EndpointForm.jsx`

### Add a new alert provider

1. Create provider struct in `internal/alerting/provider/<name>/`
2. Add field to `internal/alerting/config.go` (e.g., `Slack *slack.AlertingProvider`)
3. Add case to `GetAlertingProviderByAlertType()` switch
4. Add env var parsing in `internal/config/env.go` (parse JSON config)
5. Update `env.example` with new variable

### Add a new storage backend

See Section 4 (Storage architecture) for the dialect pattern.

### Modify the SPA

```bash
cd web/app
npm run dev          # Dev server with hot reload
# Edit src/App.jsx or components
npm run build        # Production build
make build           # Rebuild Go binary with new static assets
```
