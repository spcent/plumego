# Production Service Reference

`reference/production-service` shows a production-oriented Plumego application
shape. It extends `reference/standard-service` with explicit security and
request-observability wiring while keeping the canonical bootstrap visible.

It is an application reference, not a framework layer. Do not copy it as a
hidden production bundle.

## What It Demonstrates

- app-local configuration in `internal/config`
- deployment environment surfaced with `APP_ENV`
- explicit middleware order in `internal/app/app.go`
- visible route registration in `internal/app/routes.go`
- request IDs, recovery, body limits, timeout, security headers, abuse guard,
  tracing hook, HTTP metrics, and access logs
- stable-root-only production baseline; optional `x/*` capabilities remain
  explicit add-ons

## Routes

- `GET /` service metadata
- `GET /healthz` liveness, public by default for load balancers and probes
- `GET /readyz` readiness, public by default for load balancers and probes
- `GET /api/status` production profile summary
- `GET /api/profile` protected tenant-aware profile read using
  `Authorization: Bearer <APP_API_TOKEN>` and `X-Tenant-ID`
- `GET /ops/metrics` in-process request metric stats, protected by
  `Authorization: Bearer <OPS_TOKEN>`

`APP_API_TOKEN` is intentionally read from the environment and not given a
fallback. If it is unset, protected API routes fail closed with
`401 Unauthorized`.

`OPS_TOKEN` is intentionally read from the environment and not given a fallback.
If it is unset, `/ops/metrics` fails closed with `401 Unauthorized`.

The profile route uses `x/tenant/resolve` to attach the tenant ID to request
context, then reads from an app-local in-memory store. The store is intentionally
small and standard-library-only; real services should replace it with their own
persistence layer while preserving the explicit route, auth, and tenant wiring.

`x/devtools` is not mounted. If local debug routes are needed, wire them
explicitly and protect them outside production defaults. Do not expose
`/_debug/*` as a replacement for production ops routes.

## Run

```bash
go run ./reference/production-service
```

Useful environment variables:

- `APP_ADDR`
- `APP_ENV`
- `APP_SERVICE_NAME`
- `APP_API_TOKEN`
- `APP_BODY_LIMIT_BYTES`
- `APP_REQUEST_TIMEOUT`
- `APP_RATE_LIMIT`
- `APP_RATE_BURST`
- `OPS_TOKEN`

Example protected metrics request:

```bash
OPS_TOKEN=local-admin-token go run ./reference/production-service
curl -H 'Authorization: Bearer local-admin-token' http://127.0.0.1:8080/ops/metrics
```

Example protected tenant API request:

```bash
APP_API_TOKEN=local-api-token go run ./reference/production-service
curl \
  -H 'Authorization: Bearer local-api-token' \
  -H 'X-Tenant-ID: tenant-a' \
  http://127.0.0.1:8080/api/profile
```

## Deployment And Storage Notes

`/api/status` exposes the reference deployment, security, storage, API, and ops
policies as implemented behavior. It does not expose token values.

Use deployment-specific secret management to provide `APP_API_TOKEN` and
`OPS_TOKEN`; this reference intentionally has no hard-coded fallback for those
tokens. Treat `APP_ENV` as an operator label only, not as a switch that changes
security behavior.

The in-memory profile store is app-local in `internal/app`. To use a durable
store, replace `profileStore` behind `App.Profiles` with an application-owned
repository while keeping:

- `contract.WriteResponse` / `contract.WriteError` response paths
- `middleware/auth` bearer checks on protected API and ops routes
- `x/tenant/resolve` for tenant context extraction
- explicit route registration in `internal/app/routes.go`

Do not move persistence wiring into `core` or make devtools part of the
production default path.
