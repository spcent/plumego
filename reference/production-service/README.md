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
- `GET /healthz` liveness
- `GET /readyz` readiness
- `GET /api/status` production profile summary
- `GET /ops/metrics` in-process request metric stats for the example

`x/devtools` is not mounted. If local debug routes are needed, wire them
explicitly and protect them outside production defaults.

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
