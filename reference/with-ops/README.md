# with-ops Feature Demo

`reference/with-ops` is a non-canonical feature demo.

It shows how to mount protected `x/ops` routes and stable request-observability
middleware without treating debug tools as production defaults.

`x/ops` and `x/observability` remain experimental until promotion evidence is
complete.

## What It Demonstrates

- protected `/ops/*` routes using a static admin token
- stable request ID, recovery, access log, and HTTP metrics middleware
- app-local `/metrics` route exposing in-process metrics stats
- no `x/devtools` mounting

## Routes

- `GET /`
- `GET /metrics`
- `GET /ops`
- `GET /ops/queue?queue=primary`

Set `OPS_TOKEN` and pass it as `Authorization: Bearer <token>` for `/ops/*`.

## Run

```bash
OPS_TOKEN=local-admin-token go run ./reference/with-ops
```
