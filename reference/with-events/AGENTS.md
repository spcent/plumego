# AGENTS.md - reference/with-events

Operational guide for agents working in `reference/with-events`.

This file is intentionally short. The repository root `AGENTS.md` remains
authoritative; when a task becomes architectural, security-sensitive, or
cross-module, fall back to the root workflow and the matching
`specs/task-routing.yaml` entry.

## 1. Minimal Context

For changes confined to this directory, load only:

- repository root `AGENTS.md`, if not already loaded
- this file
- the touched Go files and their tests

Read `reference/standard-service` first — this service has the same core shape
and extends it with event-driven messaging.

## 2. Purpose And Boundaries

`reference/with-events` demonstrates the publish/subscribe pattern using an
in-process broker from `x/messaging`. A domain handler publishes events; a
consumer subscribes and delivers them as outbound webhooks. No external MQ is
required — the broker is wired entirely in-process.

Hard rules:

- `x/messaging` is the only allowed `x/*` import.
- No new third-party dependencies.
- No hidden globals, `init()` registration, or reflection routing.
- Keep `main.go` thin: load config, construct app, register routes, start.
- Keep middleware wiring explicit in `internal/app/app.go`.
- Keep all routes explicit in `internal/app/routes.go`.
- Preserve `func(http.ResponseWriter, *http.Request)` handler shape.
- Use `contract.WriteResponse` for success and `contract.WriteError` with
  `contract.NewErrorBuilder()` for errors.
- Consumers must be started in `app.Start()` (or equivalent), not in `init()`.

## 3. Package Ownership

- `main.go`: process entrypoint only.
- `internal/config`: config loading, defaults, and environment (WEBHOOK_TARGET_URL).
- `internal/app/app.go`: middleware chain (requestid → recovery), consumer startup,
  graceful shutdown.
- `internal/app/routes.go`: route table — POST /orders and GET /orders/:id.
- `internal/handler/order`: order creation handler and OrderConsumer event listener.
- `internal/handler/orderwebhook`: webhook sender that delivers events to the target URL.

Dependency direction:

```text
main.go → config → app → handler/order, handler/orderwebhook
handler/order → contract, router, x/messaging
handler/orderwebhook → x/messaging
```

Handlers must not import `internal/app` or `internal/config`.

## 4. Change Patterns

Add a new event type:

1. Define the event struct in the relevant domain package.
2. Publish it via the broker in the handler where the domain action occurs.
3. Add a consumer that subscribes to the new topic.
4. Start the consumer in `app.Start()` (or hook into the lifecycle via context).
5. Add tests for both the publisher and the consumer.

Add a new domain endpoint:

1. Implement the handler in `internal/handler/<domain>`.
2. Register the route in `routes.go`.
3. Add focused handler tests.

Change the webhook target:

1. Set `WEBHOOK_TARGET_URL` in `.env` or the process environment.
2. The sender reads this from config at startup — no code change needed.

## 5. Validation

For docs-only changes:

```bash
git diff --check
```

For Go changes inside this service:

```bash
cd reference/with-events && go test -race -timeout 30s ./...
go run ./internal/checks/dependency-rules
```

## 6. Review Focus

When reviewing or optimizing this service, check:

- Consumers started in `app.Start()`, not at package init or in handlers.
- No event payload contains credentials or PII that should not be logged.
- Idempotency stores are checked before processing to prevent duplicate delivery.
- `x/messaging` only — no other `x/*` packages.
- Response envelope uses `contract.WriteResponse` / `contract.WriteError`.
