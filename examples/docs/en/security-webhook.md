# Security and webhook modules

The **security** and webhook packages handle inbound signature verification and outbound delivery management.

## Inbound webhooks
- Configure secrets and limits with `core.WithWebhookIn` (GitHub and Stripe are supported). Options include `GitHubSecret`, `StripeSecret`, `StripeTolerance`, `MaxBodyBytes`, and topic prefixes.
- Requests are validated against HMAC signatures; failures surface as `signature mismatch` in logs.
- Valid events are published to the Pub/Sub bus using prefixed topics such as `in.github.*` and `in.stripe.*` for downstream consumers.

## Outbound webhooks
- Enable delivery management via `core.WithWebhookOut` (trigger token, base path, include-stats flag, default pagination limits).
- API endpoints manage targets, trigger deliveries, and replay failures; mount them on the router when the component is enabled.
- Backed by the in-memory store by default; can be replaced with custom store implementations if persistence is required.

## Operational guidance
- Keep secrets outside source control; load from environment variables (`GITHUB_WEBHOOK_SECRET`, `STRIPE_WEBHOOK_SECRET`, `WEBHOOK_TRIGGER_TOKEN`).
- Set `MaxBodyBytes` conservatively to prevent memory exhaustion from large webhook payloads.
- Combine inbound webhooks with Pub/Sub and WebSocket to broadcast events to clients in real time.

## Example wiring
```go
app := core.New(
    core.WithWebhookIn(core.WebhookInConfig{
        GitHubSecret:      os.Getenv("GITHUB_WEBHOOK_SECRET"),
        StripeSecret:      os.Getenv("STRIPE_WEBHOOK_SECRET"),
        MaxBodyBytes:      1 << 20,
        StripeTolerance:   5 * time.Minute,
        TopicPrefixGitHub: "in.github.",
        TopicPrefixStripe: "in.stripe.",
    }),
    core.WithWebhookOut(core.WebhookOutConfig{
        TriggerToken: os.Getenv("WEBHOOK_TRIGGER_TOKEN"),
        BasePath:     "/webhooks",
    }),
)
```

## Where to look in the repo
- `security/webhook` for signature validators and helpers.
- `core/webhook_in.go` and `core/webhook_out.go` for component wiring and routing.
- `env.example` for environment variable names tied to webhook configuration.
