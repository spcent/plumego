# Security and Webhook modules

Plumego includes helpers for inbound webhook verification, outbound delivery management, and lightweight auth utilities.

## Inbound webhooks
`core.WithWebhookIn` mounts verified receivers for GitHub and Stripe.

```go
app := core.New(core.WithWebhookIn(core.WebhookInConfig{
    Enabled:           true,
    Pub:               bus, // optional: publish events to the in-process bus
    GitHubSecret:      config.GetString("GITHUB_WEBHOOK_SECRET", "dev-github"),
    StripeSecret:      config.GetString("STRIPE_WEBHOOK_SECRET", "whsec_dev"),
    MaxBodyBytes:      1 << 20,
    StripeTolerance:   5 * time.Minute,
    TopicPrefixGitHub: "in.github.",
    TopicPrefixStripe: "in.stripe.",
}))
```

- GitHub: HMAC signature validation using the shared secret.
- Stripe: timestamp tolerance + signature validation.
- Optional Pub/Sub publication lets you decouple processing from request lifecycles.
- `MaxBodyBytes` protects against oversized payloads.

### Generic HMAC verification (replay + IP allowlist)

Use the generic verifier for custom providers:

```go
allowlist, _ := webhookin.NewIPAllowlist([]string{"203.0.113.0/24"})
nonceStore := webhookin.NewMemoryNonceStore(10 * time.Minute)

result, err := webhookin.VerifyHMAC(r, webhookin.HMACConfig{
    Secret:   []byte("shared-secret"),
    Header:   "X-Signature",
    Prefix:   "sha256=",
    Encoding: webhookin.EncodingHex,
    Replay: webhookin.HMACReplayConfig{
        TimestampHeader: "X-Timestamp",
        NonceHeader:     "X-Nonce",
        Tolerance:       5 * time.Minute,
        NonceStore:      nonceStore,
    },
    IPAllowlist: allowlist,
})
if err != nil {
    http.Error(w, "invalid signature", webhookin.HTTPStatus(err))
    return
}
payload := result.Body
```

The verifier supports:
- HMAC-SHA256/512
- Timestamp tolerance + nonce replay protection
- Optional IP allowlist with CIDR support
- Structured error codes via `webhookin.ErrorCodeOf(err)`

### End-to-end inbound handler

```go
allowlist, _ := webhookin.NewIPAllowlist([]string{"203.0.113.0/24"})
nonceStore := webhookin.NewMemoryNonceStore(10 * time.Minute)

mux := http.NewServeMux()
mux.Handle("/webhooks/inbound", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    result, err := webhookin.VerifyHMAC(r, webhookin.HMACConfig{
        Secret:   []byte(os.Getenv("WEBHOOK_SECRET")),
        Header:   "X-Signature",
        Prefix:   "sha256=",
        Encoding: webhookin.EncodingHex,
        Replay: webhookin.HMACReplayConfig{
            TimestampHeader: "X-Timestamp",
            NonceHeader:     "X-Nonce",
            Tolerance:       5 * time.Minute,
            NonceStore:      nonceStore,
        },
        IPAllowlist: allowlist,
    })
    if err != nil {
        http.Error(w, "invalid signature", webhookin.HTTPStatus(err))
        return
    }

    // process result.Body ...
    w.WriteHeader(http.StatusOK)
}))

log.Fatal(http.ListenAndServe(":8080", mux))
```

For a minimal runnable version, see `examples/webhook_example.go`.

## Outbound webhooks
`core.WithWebhookOut` wires the outbound delivery service (backed by an in-memory store in the example, extensible via `webhookout.Service`).

```go
store := webhookout.NewMemStore()
svc := webhookout.NewService(store, webhookout.ConfigFromEnv())
app := core.New(core.WithWebhookOut(core.WebhookOutConfig{
    Enabled:          true,
    Service:          svc,
    TriggerToken:     config.GetString("WEBHOOK_TRIGGER_TOKEN", "dev-token"),
    BasePath:         "/webhooks",
    IncludeStats:     true,
    DefaultPageLimit: 50,
}))
svc.Start(context.Background())
defer svc.Stop()
```

Features:
- CRUD for webhook targets and secrets via HTTP endpoints under `BasePath`.
- Delivery attempts with retry, replay, and optional stats exposure.
- Trigger token protects mutation endpoints; mount additional middleware (auth, rate limit) as needed.

## Simple auth helpers
- `middleware.SimpleAuth("token")`: checks `Authorization: Bearer <token>`.
- `middleware.APIKey(header, value)`: enforces a specific header-based API key.

Use these for admin consoles or internal endpoints (metrics, webhook management). Pair with TLS and rotate secrets via environment variables.

## Operational tips
- Never log raw webhook signatures or payload secrets.
- Keep inbound handlers fast; publish to Pub/Sub or queue for async processing instead of blocking webhook senders.
- Validate secrets early (during boot) so misconfigurations fail fast.

## Where to look in the repo
- `core/webhook.go`: inbound configuration wiring.
- `net/webhookout/`: outbound delivery service, store implementations, and HTTP handlers.
- `middleware/auth.go`: simple bearer/API-key middleware helpers.
- `examples/reference/main.go`: end-to-end wiring for inbound + outbound webhooks with Pub/Sub fan-out.
