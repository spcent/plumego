# with-webhook Feature Demo

`reference/with-webhook` is a **non-canonical feature demo**.

It shows how to add `x/webhook` inbound webhook receivers (GitHub and Stripe) to a
service that follows the same bootstrap structure as `reference/standard-service`.

**This is not the canonical app layout.** See `reference/standard-service` for that.

## What it demonstrates

- Wiring an `x/webhook` inbound receiver and an `x/pubsub` broker into the app constructor
- Registering inbound webhook routes via `webhook.RegisterRoutes`
- Verifying HMAC signatures and forwarding events to the in-process broker
- Keeping the bootstrap shape (config → app → routes → start) identical to the canonical path

## Design constraints

- depends on the same stable roots as `reference/standard-service`
- also imports `x/webhook` and `x/pubsub` (intentional — this is a feature demo)
- keeps webhook wiring in `internal/app/app.go`, not in `main.go`
- keeps route registration explicit in `internal/app/routes.go`

## Configuration

| Variable                | Default            | Description                      |
|-------------------------|--------------------|----------------------------------|
| `APP_ADDR`              | `:8085`            | Listen address                   |
| `GITHUB_WEBHOOK_SECRET` | (required for GH)  | GitHub HMAC secret               |
| `STRIPE_WEBHOOK_SECRET` | (required for STP) | Stripe HMAC secret               |

## Run it

```bash
GITHUB_WEBHOOK_SECRET=mysecret go run ./reference/with-webhook
```

Send a test GitHub webhook to `http://localhost:8085/webhooks/github`.
