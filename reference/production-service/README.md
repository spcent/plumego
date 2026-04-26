# Production Service Reference

`reference/production-service` shows a production-oriented Plumego application
shape. It extends `reference/standard-service` with explicit security and
request-observability wiring while keeping the canonical bootstrap visible.

It is an application reference, not a framework layer. Do not copy it as a
hidden production bundle.

## What It Demonstrates

- app-local configuration in `internal/config`
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
- `GET /ops/metrics` in-process request metric stats, protected by
  `Authorization: Bearer <OPS_TOKEN>`

`OPS_TOKEN` is intentionally read from the environment and not given a fallback.
If it is unset, `/ops/metrics` fails closed with `401 Unauthorized`.

`x/devtools` is not mounted. If local debug routes are needed, wire them
explicitly and protect them outside production defaults. Do not expose
`/_debug/*` as a replacement for production ops routes.

## Run

```bash
go run ./reference/production-service
```

Useful environment variables:

- `APP_ADDR`
- `APP_SERVICE_NAME`
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
