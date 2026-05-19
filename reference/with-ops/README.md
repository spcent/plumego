# with-ops Scenario Reference

`reference/with-ops` is a non-canonical scenario reference.

It shows how to mount protected `x/observability/ops` routes and stable request-observability
middleware without treating debug tools as production defaults.

`x/observability/ops` and `x/observability` remain experimental until promotion evidence is
complete.

## What It Demonstrates

- protected `/ops/*` routes using a static admin token
- stable request ID, recovery, access log, and HTTP metrics middleware
- app-local `/metrics` route exposing in-process metrics stats
- no `x/observability/devtools` mounting

## Routes

- `GET /`
- `GET /metrics`
- `GET /ops`
- `GET /ops/queue?queue=primary`

Set `OPS_TOKEN` and pass it as `Authorization: Bearer <token>` for `/ops/*`.

## Run

```bash
cd reference/with-ops
OPS_TOKEN=local-admin-token go run .
```
