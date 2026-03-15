# with-websocket Feature Demo

`reference/with-websocket` is a **non-canonical feature demo**.

It shows how to add `x/websocket` (WebSocket server with hub) to a service that
follows the same bootstrap structure as `reference/standard-service`.

**This is not the canonical app layout.** See `reference/standard-service` for that.

## What it demonstrates

- Wiring an `x/websocket` server into the app constructor
- Registering WebSocket routes via `ws.RegisterRoutes`
- Keeping the bootstrap shape (config → app → routes → start) identical to the canonical path

## Design constraints

- depends on the same stable roots as `reference/standard-service`
- also imports `x/websocket` for the server (intentional — this is a feature demo)
- keeps WebSocket wiring in `internal/app/app.go`, not in `main.go`
- keeps route registration explicit in `internal/app/routes.go`

## Configuration

| Variable     | Default  | Description                        |
|--------------|----------|------------------------------------|
| `APP_ADDR`   | `:8084`  | Listen address                     |
| `WS_SECRET`  | required | JWT secret (min 32 bytes)          |

## Run it

```bash
WS_SECRET=changeme-replace-with-32-byte-key go run ./reference/with-websocket
```

Connect a WebSocket client to `ws://localhost:8084/ws`.
