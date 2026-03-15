# with-gateway Feature Demo

`reference/with-gateway` is a **non-canonical feature demo**.

It shows how to add `x/gateway` (reverse proxy) to a service that follows the
same bootstrap structure as `reference/standard-service`.

**This is not the canonical app layout.** See `reference/standard-service` for that.

## What it demonstrates

- Wiring an `x/gateway` reverse proxy into the app constructor
- Proxying all `/proxy/*` requests to a configurable backend
- Keeping the bootstrap shape (config → app → routes → start) identical to the canonical path

## Design constraints

- depends on the same stable roots as `reference/standard-service`
- also imports `x/gateway` for the proxy (intentional — this is a feature demo)
- keeps gateway wiring in `internal/app/app.go`, not in `main.go`
- keeps route registration explicit in `internal/app/routes.go`

## Configuration

| Variable          | Default                   | Description              |
|-------------------|---------------------------|--------------------------|
| `APP_ADDR`        | `:8083`                   | Listen address           |
| `GATEWAY_BACKEND` | `http://localhost:9090`   | Backend URL to proxy to  |

## Run it

```bash
GATEWAY_BACKEND=http://localhost:9090 go run ./reference/with-gateway
```
