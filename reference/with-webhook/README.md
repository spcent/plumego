# with-webhook Scenario Reference

`reference/with-webhook` is a **non-canonical scenario reference**.

It shows how to add `x/messaging/webhook` inbound webhook receivers (GitHub and Stripe) to a
service that follows the same bootstrap structure as `reference/standard-service`.

**This is not the canonical app layout.** See `reference/standard-service` for that.

## What it demonstrates

- Wiring an `x/messaging/webhook` inbound receiver and an `x/messaging/pubsub` broker into the app constructor
- Registering inbound webhook routes via `webhook.RegisterRoutes`
- Verifying HMAC signatures and forwarding events to the in-process broker
- Keeping the bootstrap shape (`main.run` → `app.Start(ctx)`) aligned with the canonical path
- Loading app config with the same precedence as the canonical service:
  `Defaults < .env < process env < flags`

## Design constraints

- depends on the same stable roots as `reference/standard-service`
- also imports `x/messaging/webhook` and `x/messaging/pubsub` (intentional — this is a scenario reference)
- keeps webhook wiring in `internal/app/app.go`, not in `main.go`
- keeps route registration explicit in `internal/app/routes.go`
- keeps process signal ownership in `main.go`; `internal/app` only reacts to the caller-owned context

## Configuration

| Variable                | Default            | Description                      |
|-------------------------|--------------------|----------------------------------|
| `APP_ADDR`              | `:8085`            | Listen address                   |
| `GITHUB_WEBHOOK_SECRET` | (required for GH)  | GitHub HMAC secret               |
| `STRIPE_WEBHOOK_SECRET` | (required for STP) | Stripe HMAC secret               |

## Run it

```bash
cd reference/with-webhook
GITHUB_WEBHOOK_SECRET=mysecret go run .
```

Send a test GitHub webhook to `http://localhost:8085/webhooks/github`.
