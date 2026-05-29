# guardus

Feature-equivalent v1 reimplementation of [Gatus](https://github.com/TwiN/gatus)
backend on top of [plumego](https://github.com/spcent/plumego) stable roots and
extensions.

This use-case sits alongside `use-cases/gatus` (the original Gatus drop-in) to
validate plumego in monitoring/observability product use cases.

## Status

v1 backend complete: HTTP/TCP/ICMP/DNS/TLS/STARTTLS/gRPC/WebSocket probes,
ExternalEndpoint push, SQLite + memory storage, Basic Auth, the five v1 alert
providers (custom/email/slack/telegram/discord), maintenance windows,
connectivity checker, Prometheus metrics, /health, and embedded Vue SPA. See
[AGENTS.md](./AGENTS.md) for v1 scope.

## Build

```bash
make build
make run
```

## Configuration

Deployment settings come from environment variables (`GUARDUS_*`). See
[`env.example`](./env.example) for the full surface — copy it, edit values,
and `source` (or set the variables in your orchestrator) before `make run`.

Endpoint definitions live in the storage backend (SQLite by default at
`./guardus.db`). On each startup, if `GUARDUS_BOOTSTRAP_FILE` points at
an existing JSON document, every entry is upserted by key into the store, then
the watchdog runs against whatever the store now holds:

```json
{
  "endpoints": [
    {
      "name": "front-end",
      "url": "https://example.org",
      "interval": "5m",
      "conditions": ["[STATUS] == 200"]
    }
  ]
}
```

The bootstrap file is optional — once endpoints are in the database they
persist across restarts. An admin API for managing endpoints at runtime is
the next milestone; for now, edits made directly to the DB take effect after
restart.

## Frontend

The SPA is rewritten in React + Vite (source in `web/app/`). Run:

```bash
make frontend-build   # installs Node deps and builds into web/static/
make build            # Go binary embeds web/static/ via //go:embed
```

The dev server proxies `/api`, `/health`, `/metrics` to `http://localhost:8080`
so `npm run dev` works against a running backend.

Suites and OIDC are intentionally omitted (out of v1 scope).
