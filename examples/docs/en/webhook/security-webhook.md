# Security and Webhook modules

Plumego supports inbound verification and outbound delivery management through explicit webhook components.

## Inbound webhooks (GitHub / Stripe)
Mount the inbound component via `core.WithComponent(...)`:

```go
bus := pubsub.New()

app := core.New(
    core.WithComponent(webhook.NewWebhookInComponent(webhook.WebhookInConfig{
        Enabled:           true,
        Pub:               bus,
        GitHubSecret:      os.Getenv("GITHUB_WEBHOOK_SECRET"),
        StripeSecret:      os.Getenv("STRIPE_WEBHOOK_SECRET"),
        MaxBodyBytes:      1 << 20,
        StripeTolerance:   5 * time.Minute,
        TopicPrefixGitHub: "in.github.",
        TopicPrefixStripe: "in.stripe.",
    }, bus, nil)),
)
```

Default inbound endpoints:
- `POST /webhooks/github`
- `POST /webhooks/stripe`

## Outbound webhook management
Mount outbound management as a component:

```go
store := webhook.NewMemStore()
svc := webhook.NewService(store, webhook.ConfigFromEnv())

app := core.New(
    core.WithComponent(webhook.NewWebhookOutComponent(webhook.WebhookOutConfig{
        Enabled:          true,
        Service:          svc,
        TriggerToken:     os.Getenv("WEBHOOK_TRIGGER_TOKEN"),
        BasePath:         "/webhooks",
        IncludeStats:     true,
        DefaultPageLimit: 50,
    })),
)
```

The outbound component exposes target CRUD, delivery listing/detail, replay, and trigger endpoints under `BasePath`.

## Generic signature verification
For custom providers, use `x/webhook` verification helpers directly:

```go
result, err := webhook.VerifyHMAC(r, webhook.HMACConfig{
    Secret:   []byte(os.Getenv("WEBHOOK_SECRET")),
    Header:   "X-Signature",
    Prefix:   "sha256=",
    Encoding: webhook.EncodingHex,
})
if err != nil {
    http.Error(w, "invalid signature", webhook.HTTPStatus(err))
    return
}
_ = result.Body
```

## Security middleware helpers
Use middleware subpackages explicitly:

- `auth.SimpleAuth(token)`
- `security.SecurityHeaders(nil)`
- `ratelimit.AbuseGuard(...)`

## Operational notes
- Never log raw signatures, webhook secrets, or private keys.
- Keep inbound handlers short; publish to Pub/Sub for async processing.
- Validate secret config during boot and fail fast.
