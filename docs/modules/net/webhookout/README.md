# Outbound Webhook Delivery

> **Package**: `github.com/spcent/plumego/net/webhookout` | **Feature**: Reliable webhook delivery

## Overview

The `webhookout` package provides reliable outbound webhook delivery with automatic retries, dead-letter queuing, signature signing, and delivery tracking.

## Quick Start

```go
import "github.com/spcent/plumego/net/webhookout"

// Create delivery service
service := webhookout.NewService(webhookout.Config{
    Workers:     8,               // Concurrent delivery workers
    QueueSize:   2048,            // In-flight webhook buffer
    Timeout:     10 * time.Second, // Per-request timeout
    MaxRetries:  5,               // Retry attempts on failure
    RetryDelay:  time.Minute,     // Initial retry delay (exponential backoff)
})
service.Start()
defer service.Stop()

// Register as Plumego runner
app := core.New(
    core.WithRunner(service),
)

// Send webhook
err := service.Send(webhookout.Webhook{
    URL:     "https://customer.example.com/webhook",
    Payload: eventJSON,
    Headers: map[string]string{
        "X-Event-Type": "order.created",
    },
    Secret: "hmac-signing-secret", // Signs payload for verification
})
```

## Webhook Structure

```go
type Webhook struct {
    // Destination
    URL    string

    // Payload (JSON)
    Payload []byte

    // Custom headers
    Headers map[string]string

    // HMAC signing secret (optional)
    // Signs payload as: X-Webhook-Signature: sha256=<hex>
    Secret string

    // Unique event ID for deduplication (optional)
    EventID string

    // Delivery metadata for tracking
    TenantID string
    UserID   string
    EventType string

    // Retry configuration (overrides service defaults)
    MaxRetries int
    RetryDelay time.Duration
}
```

## Configuration

```go
service := webhookout.NewService(webhookout.Config{
    // Concurrency
    Workers:    8,
    QueueSize:  2048,

    // Timeouts
    Timeout:    10 * time.Second,

    // Retry policy
    MaxRetries:     5,
    RetryDelay:     time.Minute,        // Initial delay
    MaxRetryDelay:  30 * time.Minute,   // Maximum delay (exponential cap)
    RetryMultiplier: 2.0,               // Exponential multiplier

    // Dead-letter queue (for permanently failed webhooks)
    DeadLetterStore: dlqStore,

    // Persistence (survive restarts)
    Store: webhookStore,
})
```

## Retry Policy

Failed webhooks are retried with exponential backoff:

| Attempt | Delay |
|---------|-------|
| 1 | 1 minute |
| 2 | 2 minutes |
| 3 | 4 minutes |
| 4 | 8 minutes |
| 5 | 16 minutes |

After all retries fail, the webhook moves to the dead-letter queue.

**Retry conditions**:
- Network errors
- HTTP 5xx responses
- HTTP 429 (rate limited) — respects `Retry-After` header
- Request timeout

**No retry**:
- HTTP 4xx responses (client error, won't succeed on retry)
- Invalid URL

## Signature Signing

When a `Secret` is provided, webhooks are signed for endpoint verification:

```go
service.Send(webhookout.Webhook{
    URL:     endpoint.URL,
    Payload: eventJSON,
    Secret:  endpoint.SigningSecret,
})
// Adds header: X-Webhook-Signature: sha256=<hmac-sha256>

// Timestamp to prevent replay attacks:
// X-Webhook-Timestamp: 1706745600
```

**Recipient verification**:
```go
// Endpoint verifies the signature
mac := hmac.New(sha256.New, []byte(signingSecret))
mac.Write([]byte(timestamp + "." + string(body)))
expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
received := r.Header.Get("X-Webhook-Signature")

if !hmac.Equal([]byte(expected), []byte(received)) {
    http.Error(w, "Invalid signature", http.StatusUnauthorized)
    return
}
```

## Delivery Tracking

```go
service := webhookout.NewService(config,
    webhookout.WithDeliveryLog(func(delivery webhookout.Delivery) {
        log.Infof("webhook %s to %s: status=%d attempt=%d",
            delivery.EventID, delivery.URL, delivery.StatusCode, delivery.Attempt)

        // Store in database for dashboard
        deliveryStore.Record(delivery)
    }),
)
```

```go
type Delivery struct {
    ID         string
    WebhookID  string
    EventID    string
    URL        string
    StatusCode int
    Duration   time.Duration
    Attempt    int
    Error      string
    Timestamp  time.Time
    Success    bool
}
```

## Dead-Letter Queue

Permanently failed webhooks go to the DLQ for manual inspection and replay:

```go
// Configure DLQ storage
dlqStore := webhookout.NewKVDeadLetterStore(kvStore)

service := webhookout.NewService(webhookout.Config{
    DeadLetterStore: dlqStore,
})

// Admin: inspect DLQ
failed, err := dlqStore.List(ctx, 100)
for _, w := range failed {
    fmt.Printf("Failed: %s → %s (%d attempts)\n",
        w.EventID, w.URL, w.Attempts)
}

// Admin: replay a failed webhook
err = service.Replay(ctx, failed[0].ID)

// Admin: bulk replay all DLQ entries
err = service.ReplayAll(ctx, dlqStore)
```

## Multi-Tenant Webhook Distribution

```go
type WebhookEndpointStore interface {
    ListEndpoints(ctx context.Context, tenantID string) ([]Endpoint, error)
}

// Broadcast event to all tenant's registered endpoints
func broadcastEvent(ctx context.Context, tenantID string, event any) error {
    endpoints, err := endpointStore.ListEndpoints(ctx, tenantID)
    if err != nil {
        return err
    }

    eventJSON, _ := json.Marshal(event)
    for _, ep := range endpoints {
        webhookService.Send(webhookout.Webhook{
            URL:       ep.URL,
            Payload:   eventJSON,
            Secret:    ep.SigningSecret,
            TenantID:  tenantID,
            EventType: event.Type(),
        })
    }
    return nil
}
```

## Metrics

```go
// Emitted metrics
webhooks_sent_total{status="success"}
webhooks_sent_total{status="failure"}
webhooks_retried_total
webhooks_dlq_total
webhook_delivery_duration_seconds
webhook_queue_size
```

## Environment Configuration

```yaml
# env.example
WEBHOOK_QUEUE_SIZE=2048     # Outbound queue buffer
WEBHOOK_WORKERS=8           # Delivery worker count
WEBHOOK_TIMEOUT_SECONDS=10  # Per-request timeout
WEBHOOK_MAX_RETRIES=5       # Retry attempts
```

## Related Documentation

- [Inbound Webhooks](../webhookin/README.md) — Receiving webhooks
- [Security](../../security/) — HMAC signature verification
- [Scheduler](../../scheduler/README.md) — Scheduled webhook delivery
