# Inbound Webhooks

> **Package**: `github.com/spcent/plumego/net/webhookin` | **Providers**: GitHub, Stripe, Custom

For application-facing imports, prefer `github.com/spcent/plumego/x/webhook`. This page documents the current legacy implementation package.

## Overview

The `webhookin` package receives and verifies webhook events from external services. All webhooks are verified using HMAC signatures before processing, preventing unauthorized event injection.

## Quick Start

### GitHub Webhooks

```go
import "github.com/spcent/plumego/net/webhookin"

handler := webhookin.NewGitHub(
    webhookin.WithSecret(os.Getenv("GITHUB_WEBHOOK_SECRET")),
    webhookin.OnPush(func(ctx context.Context, event *webhookin.PushEvent) error {
        log.Printf("Push to %s by %s", event.Repository.Name, event.Pusher.Name)
        return deployPipeline.Trigger(ctx, event.Repository.Name, event.After)
    }),
    webhookin.OnPullRequest(func(ctx context.Context, event *webhookin.PullRequestEvent) error {
        if event.Action == "opened" {
            return reviewer.AssignReviewers(ctx, event.PullRequest.Number)
        }
        return nil
    }),
    webhookin.OnRelease(func(ctx context.Context, event *webhookin.ReleaseEvent) error {
        return notifier.Notify(ctx, "New release: "+event.Release.TagName)
    }),
)

app.Post("/webhooks/github", handler)
```

### Stripe Webhooks

```go
stripeHandler := webhookin.NewStripe(
    webhookin.WithStripeSecret(os.Getenv("STRIPE_WEBHOOK_SECRET")),
    webhookin.WithToleranceSeconds(300), // 5 minute timestamp tolerance
    webhookin.OnPaymentSucceeded(func(ctx context.Context, event *webhookin.PaymentEvent) error {
        return orderService.FulfillOrder(ctx, event.PaymentIntent.Metadata["order_id"])
    }),
    webhookin.OnPaymentFailed(func(ctx context.Context, event *webhookin.PaymentEvent) error {
        return orderService.MarkPaymentFailed(ctx, event.PaymentIntent.Metadata["order_id"])
    }),
    webhookin.OnSubscriptionCanceled(func(ctx context.Context, event *webhookin.SubscriptionEvent) error {
        return subscriptionService.Deactivate(ctx, event.Subscription.ID)
    }),
)

app.Post("/webhooks/stripe", stripeHandler)
```

## GitHub Events

### Supported Events

| Event | Handler | Triggered When |
|-------|---------|----------------|
| `push` | `OnPush` | Commits pushed to branch |
| `pull_request` | `OnPullRequest` | PR opened, closed, merged |
| `release` | `OnRelease` | Release published |
| `issues` | `OnIssue` | Issue created, updated, closed |
| `workflow_run` | `OnWorkflowRun` | GitHub Actions workflow completed |
| `*` | `OnAny` | Any event type |

### Push Event Handler

```go
webhookin.OnPush(func(ctx context.Context, event *webhookin.PushEvent) error {
    // Available fields:
    event.Ref             // "refs/heads/main"
    event.Before         // Previous commit SHA
    event.After          // New commit SHA
    event.Repository.Name // Repository name
    event.Repository.FullName // "owner/repo"
    event.Pusher.Name    // Username
    event.Commits        // List of commits
    event.HeadCommit.Message // Latest commit message

    // Deploy when main branch is pushed
    if event.Ref == "refs/heads/main" {
        return ci.Deploy(ctx, event.Repository.Name, event.After)
    }
    return nil
})
```

## Stripe Events

### Supported Events

| Event | Handler | Triggered When |
|-------|---------|----------------|
| `payment_intent.succeeded` | `OnPaymentSucceeded` | Payment completed |
| `payment_intent.payment_failed` | `OnPaymentFailed` | Payment declined |
| `customer.subscription.created` | `OnSubscriptionCreated` | New subscription |
| `customer.subscription.deleted` | `OnSubscriptionCanceled` | Subscription ended |
| `invoice.paid` | `OnInvoicePaid` | Invoice payment received |
| `charge.refunded` | `OnChargeRefunded` | Refund processed |

### Payment Event

```go
webhookin.OnPaymentSucceeded(func(ctx context.Context, event *webhookin.PaymentEvent) error {
    event.PaymentIntent.ID           // "pi_xxx"
    event.PaymentIntent.Amount       // Amount in cents
    event.PaymentIntent.Currency     // "usd"
    event.PaymentIntent.Metadata     // Custom metadata
    event.PaymentIntent.CustomerID   // Stripe customer ID

    orderID := event.PaymentIntent.Metadata["order_id"]
    return orderService.FulfillOrder(ctx, orderID)
})
```

## Custom Webhooks

Implement webhook reception for any service with HMAC-SHA256 signatures:

```go
handler := webhookin.NewCustom(
    webhookin.WithHMAC(os.Getenv("WEBHOOK_SECRET"), "X-Webhook-Signature"),
    webhookin.WithHandler(func(ctx context.Context, body []byte, headers http.Header) error {
        var event MyEvent
        if err := json.Unmarshal(body, &event); err != nil {
            return err
        }
        return processEvent(ctx, &event)
    }),
)
```

## Signature Verification

### GitHub (HMAC-SHA256)

GitHub sends `X-Hub-Signature-256: sha256=<hex>`:

```go
// Verification is automatic with NewGitHub()
// The secret must match what's configured in GitHub repository settings
```

### Stripe (Timestamp + Signature)

Stripe sends `Stripe-Signature: t=<timestamp>,v1=<signature>`:

```go
// Timestamp tolerance prevents replay attacks
webhookin.WithToleranceSeconds(300) // Max 5 minutes old
```

## IP Allowlisting

Restrict webhooks to known provider IP ranges:

```go
// Allow only GitHub webhook IPs
handler := webhookin.NewGitHub(
    webhookin.WithSecret(secret),
    webhookin.WithIPAllowlist([]string{
        "192.30.252.0/22",
        "185.199.108.0/22",
        "140.82.112.0/20",
    }),
)

// Stripe IP ranges
stripeHandler := webhookin.NewStripe(
    webhookin.WithStripeSecret(secret),
    webhookin.WithStripeIPAllowlist(), // Use built-in Stripe IP list
)
```

## Deduplication

Prevent processing the same webhook twice:

```go
handler := webhookin.NewGitHub(
    webhookin.WithSecret(secret),
    webhookin.WithDeduplication(dedupStore, 24*time.Hour),
)
// Duplicate delivery within 24h is silently skipped
```

## Error Handling

```go
webhookin.OnPush(func(ctx context.Context, event *webhookin.PushEvent) error {
    if err := deployPipeline.Trigger(ctx, event.Repository.Name); err != nil {
        // Return error → 500 response → provider will retry
        return fmt.Errorf("deployment trigger failed: %w", err)
    }
    return nil // 200 OK → provider marks as delivered
})

// Configure error response behavior
webhookin.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
    log.Errorf("webhook error: %v", err)
    http.Error(w, "processing failed", http.StatusInternalServerError)
})
```

## Testing

```go
func TestGitHubWebhook(t *testing.T) {
    secret := "test-secret"
    handler := webhookin.NewGitHub(
        webhookin.WithSecret(secret),
        webhookin.OnPush(func(ctx context.Context, event *webhookin.PushEvent) error {
            assert.Equal(t, "main", event.Repository.Name)
            return nil
        }),
    )

    body := `{"ref": "refs/heads/main", "repository": {"name": "main"}}`

    // Compute valid signature
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(body))
    sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

    req := httptest.NewRequest("POST", "/webhooks/github", strings.NewReader(body))
    req.Header.Set("X-Hub-Signature-256", sig)
    req.Header.Set("X-GitHub-Event", "push")

    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
}
```

## Related Documentation

- [Outbound Webhooks](../webhookout/README.md) — Sending webhooks
- [Security: Input](../../security/input.md) — Input validation
