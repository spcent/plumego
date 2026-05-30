# Standard Service Reference

`reference/standard-service` is Plumego's canonical reference application.

Read this immediately after `docs/start/getting-started.md`. This directory is the
default answer to "what should a Plumego application look like?"

It is the single source of truth for:

- default application layout
- bootstrap flow in `main.go`
- route registration style
- handler and response conventions
- minimal stable-root-only wiring

Design constraints:

- depends only on `core`, `router`, `contract`, `middleware`, `health`, `metrics`, and `log` plus the standard library
- excludes `x/*` capabilities from the canonical learning path
- keeps all route registration explicit in `internal/app/routes.go`
- keeps app-local configuration in `internal/config`

Configuration precedence is explicit and test-covered:
`Defaults()` < `.env` file < process environment < command-line flags.

The middleware stack in `internal/app/app.go` wires the production-relevant
stable-root set in order:

```
requestid → security → cors → recovery → accesslog → bodylimit → httpmetrics → timeout
```

`middleware/securityheaders/headers` applies default security headers (X-Frame-Options,
X-Content-Type-Options, Referrer-Policy) to every response. `middleware/cors`
uses permissive defaults for the reference app; replace with
`cors.StrictDefaultOptions(origins...)` in production. `middleware/httpmetrics`
uses a noop collector; swap in a real `metrics.HTTPObserver` for production
instrumentation.

Production services extend this stack with abuse guard rate limiting, auth
adapters, and tracing as application-local decisions. Do not replace the visible
wiring with a hidden global production bundle.

For operations, keep telemetry, admin, and debug surfaces separate:

- request metrics, tracing, access logs, and request IDs stay in stable middleware
- exporter and adapter wiring belongs in `x/observability`
- protected admin and health orchestration surfaces belong in `x/observability/ops`
- local debug endpoints belong in `x/observability/devtools` and are not mounted by default

Canonical files:

- `main.go`
- `internal/app/app.go`
- `internal/app/routes.go`
- `internal/domain/item/item.go`
- `internal/handler/api.go`
- `internal/handler/health.go`
- `internal/handler/items.go`
- `internal/config/config.go`

## Response shape

Every endpoint uses a single canonical envelope via `contract.WriteResponse` and
`contract.WriteError`. No ad hoc JSON helpers or per-handler response structs.

**Success** (`contract.WriteResponse` — only accepts 2xx status codes):
```json
{
  "data":       { "id": "...", "name": "widget", "description": "...", "created_at": "..." },
  "meta":       { "total": 42, "limit": 20, "offset": 0 },
  "request_id": "01HX..."
}
```
`data` and `meta` use `omitempty`: `meta` is absent when `nil` is passed;
`data` is absent when `nil`. For 204 No Content the body is suppressed entirely
by the contract layer — `writeJSON` writes only the status line.

**Error** (`contract.WriteError`):
```json
{
  "error": {
    "code":     "item.fields.required",
    "message":  "one or more required fields are missing",
    "category": "validation_error",
    "type":     "required_field_missing",
    "details":  { "name": "field is required", "description": "field is required" }
  },
  "request_id": "01HX..."
}
```
`type` is the `ErrorType` string value (e.g. `TypeRequired = "required_field_missing"`),
not the Go identifier. `category` is derived automatically from the type
(`TypeRequired → "validation_error"`, `TypeNotFound → "client_error"`,
`TypeUnauthorized → "auth_error"`, `TypeUnavailable → "server_error"`). `code`
is overridden per handler with a namespaced string (`<resource>.<operation>.<reason>`).
`details` carries per-field annotations; it is omitted when empty.

## Run it

```bash
cd reference/standard-service
go run .
```

The CLI `canonical` scaffold template is expected to teach the same layout and
stable-root-only wiring:

```bash
plumego new myapp --template canonical
```

The scaffold is not a byte-for-byte copy of this directory. It preserves the
same bootstrap, `internal/app`, `internal/handler`, and `internal/config`
runtime shape, while adapting the entrypoint to a generated project layout
(`cmd/app/main.go`) and adding project-local files such as `go.mod`,
`env.example`, `.gitignore`, and `README.md`. Reference-only tests stay in this
directory; generated projects should add their own tests next to changed
behavior.

Canonical next reads after this reference:

1. `docs/README.md`
2. `docs/reference/canonical-style-guide.md`
3. `docs/concepts/agent-first-repo-blueprint.md`
4. `specs/task-routing.yaml`
