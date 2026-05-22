# with-websocket Scenario Reference

`reference/with-websocket` is a **non-canonical scenario reference**.

It shows how to add `x/websocket` (WebSocket server with hub) to a service that
follows the same bootstrap structure as `reference/standard-service`.

**This is not the canonical app layout.** See `reference/standard-service` for that.

## What it demonstrates

- Wiring an `x/websocket` server into the app constructor
- Registering WebSocket routes via `ws.RegisterRoutes`
- Keeping the bootstrap shape (`main.run` → `app.Start(ctx)`) aligned with the canonical path
- Loading app config with the same precedence as the canonical service:
  `Defaults < .env < process env < flags`

## Design constraints

- depends on the same stable roots as `reference/standard-service`
- also imports `x/websocket` for the server (intentional — this is a scenario reference)
- keeps WebSocket wiring in `internal/app/app.go`, not in `main.go`
- keeps route registration explicit in `internal/app/routes.go`
- keeps process signal ownership in `main.go`; `internal/app` shuts down WebSocket before HTTP when the caller-owned context is canceled

## Configuration

| Variable     | Default  | Description                        |
|--------------|----------|------------------------------------|
| `APP_ADDR`   | `:8084`  | Listen address                     |
| `WS_SECRET`  | required | JWT secret (min 32 bytes)          |

`WS_SECRET` is intentionally not exposed as a command-line flag. Set it through
the environment or a local `.env` file.

## Run it

```bash
cd reference/with-websocket
WS_SECRET=changeme-replace-with-32-byte-key go run .
```

Connect a WebSocket client to `ws://localhost:8084/ws`.
