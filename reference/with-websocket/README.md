# with-websocket

`reference/with-websocket` shows how to add WebSocket support to a service that
follows the same bootstrap structure as `reference/standard-service`.

**This is a non-canonical scenario reference.** See `reference/standard-service`
for the canonical app layout.

## Purpose

Demonstrate WebSocket lifecycle management using `x/websocket`: upgrade, hub
registration, read loop, safe write path, ping/pong heartbeat, and graceful
shutdown. The endpoint echoes every message back to the sender.

## What it demonstrates

- HTTP route to WebSocket upgrade via `ws.RegisterRoutes`
- Explicit connection lifecycle: upgrade → hub registration → read loop → writer goroutine
- Read loop with frame validation (1 MB default message-size limit, UTF-8 enforcement,
  control-character rejection)
- Safe write path via `conn.WriteMessage` → bounded send queue → `writerPump` goroutine
- Automatic ping/pong heartbeat with pong-wait timeout that closes stale connections
- Graceful shutdown: `WS.Shutdown` closes all connections and drains the job queue
  before the HTTP server stops, driven by the caller-owned `context.Context`
- Signal-driven shutdown (`SIGINT`/`SIGTERM`) via `signal.NotifyContext`
- Config-precedence chain: `Defaults < .env < process env < flags`

## What it intentionally excludes

- **Authentication**: `AllowUnauthenticated: true` is set for simplicity; JWT
  authentication requires one config change (see [Extending](#extending))
- **Room routing**: messages are echoed directly back to the sender; no broadcast
  or room fanout
- **Persistence, databases, Redis, or external state**
- **Multi-node or pub/sub coordination**
- **TLS**: configure `core.AppConfig.TLS` in production

## Configuration

| Variable     | Default  | Description                             |
|--------------|----------|-----------------------------------------|
| `APP_ADDR`   | `:8084`  | Listen address                          |
| `WS_SECRET`  | required | JWT signing secret (min 32 bytes)       |
| `APP_DEBUG`  | `false`  | Enable verbose request/response logging |

`WS_SECRET` is required even though the demo runs with `AllowUnauthenticated: true`.
This ensures the secret is already set when you enable JWT authentication.

Generate a safe value:

```bash
openssl rand -base64 32
```

## Run

```bash
cd reference/with-websocket
WS_SECRET=$(openssl rand -base64 32) go run .
```

Or with a `.env` file:

```bash
cp env.example .env
# Edit .env: set WS_SECRET to a value of 32 bytes or more
go run .
```

## Manual test

### Health check

```bash
curl http://localhost:8084/healthz
```

Expected:

```json
{"data":{"status":"ok","service":"with-websocket","timestamp":"..."}}
```

### WebSocket echo

Install [websocat](https://github.com/vi/websocat) or
[wscat](https://github.com/websockets/wscat) as optional CLI tools.

**websocat:**

```bash
websocat ws://localhost:8084/ws
```

**wscat:**

```bash
wscat -c ws://localhost:8084/ws
```

Type a message and press Enter. The server echoes it back.

**Example session:**

```
> hello
< hello
> {"event":"ping"}
< {"event":"ping"}
```

Text and binary frames are both echoed verbatim.

**Verify upgrade rejection (non-WebSocket request):**

```bash
curl -i http://localhost:8084/ws
```

Expected: `400 Bad Request` with `{"error":{"code":"WEBSOCKET_BAD_UPGRADE",...}}`.

## Connection lifecycle

```
Client                              Server
  │── GET /ws (Upgrade) ──────────>│  handshake: origin check, key validation
  │                                │  hub.tryJoin → registered in "default" room
  │<── 101 Switching Protocols ────│  writer goroutine + pong monitor started
  │── text frame ─────────────────>│  read loop → OnMessage → WriteMessage
  │<── text frame (echo) ──────────│  writer goroutine drains send queue
  │     (idle)                     │
  │<── ping ───────────────────────│  writerPump sends periodic ping
  │── pong ─────────────────────────│  pongMonitor resets timeout
  │                                │
  │  (SIGTERM received)            │  hub.Shutdown closes all connections
  │<── TCP close ───────────────────│  read loop returns io.EOF, cleanup goroutine
  │                                │  calls hub.RemoveConn; HTTP server stops
```

## Production notes

- **Enable JWT auth**: set `AllowUnauthenticated: false` in `internal/app/app.go`
  and ensure `WS_SECRET` is a strong secret. Clients must send
  `Authorization: Bearer <jwt>`.
- **Restrict origins**: replace `AllowAllOrigins: true` with
  `AllowedOrigins: []string{"https://your-domain.com"}` to prevent cross-origin
  WebSocket connections from browsers.
- **Message size**: `DefaultMessageValidationConfig` enforces a 1 MB limit and
  rejects control characters and invalid UTF-8. Adjust via `wsCfg.MessageValidation`
  in `internal/app/app.go`.
- **Backpressure**: a slow client with a full send queue triggers `SendBehavior`
  (default `SendBlock` with timeout). Use `SendDrop` or `SendClose` under load.
- **Connection limits**: set `MaxRoomRegistrations` and `MaxConnectionRate` on
  `HubConfig` to protect against connection floods.
- **Reverse proxy (nginx)**: add these headers to pass the upgrade:
  ```nginx
  proxy_http_version 1.1;
  proxy_set_header Upgrade $http_upgrade;
  proxy_set_header Connection "upgrade";
  ```
- **TLS**: set `core.AppConfig.TLS.Enabled = true` with a certificate pair.

## Extending

**Enable JWT authentication:**

1. Set `wsCfg.AllowUnauthenticated = false` in `internal/app/app.go`.
2. Remove or restrict `wsCfg.AllowAllOrigins`.
3. Clients mint a JWT signed with `WS_SECRET` and send it as
   `Authorization: Bearer <token>`.
4. See `x/websocket/auth.go` for `NewSimpleHS256TokenAuth` and `TokenAuthenticator`.

**Broadcast from outside the connection handler:**

1. Store `a.WS.Hub()` in your service layer.
2. Call `hub.BroadcastAll(websocket.OpcodeText, payload)` or
   `hub.BroadcastRoom(room, ...)`.
3. Do not share the hub across process boundaries without an explicit message broker.

**Add room-based fanout:**

Remove the `OnMessage` handler from `wsCfg` in `app.go`. Without `OnMessage`,
`RegisterRoutes` automatically uses `ServeRoomFanoutWS`, which broadcasts each
client message to everyone in the same room (selected via `?room=` query parameter,
defaulting to `"default"`).
