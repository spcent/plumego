# Net Package Collection

This directory contains production-ready networking utilities built with Go standard library only.

## Overview

```
net/
├── http/           # HTTP client with retry, middleware, and backoff
├── mq/             # Message broker adapters (in-proc + hooks)
├── webhookin/      # Incoming webhook verification (GitHub, Stripe)
├── webhookout/     # Outgoing webhook delivery system
└── websocket/      # WebSocket server (modular architecture)
```

## HTTP Client (`net/http`)

### Features
- **Retry Logic**: Exponential backoff with jitter
- **Middleware**: Chainable request/response interceptors
- **Timeout Control**: Global and per-request timeouts
- **Multiple Policies**: Timeout, status code, composite policies
- **HTTP Methods**: GET, POST, PUT, DELETE with JSON support
- **Consistent Errors**: GET/POST/PUT/DELETE return errors on non-2xx responses

### Quick Start

```go
import "github.com/spcent/plumego/net/http"

// Create client
client := http.New(
    http.WithTimeout(5 * time.Second),
    http.WithRetryCount(3),
    http.WithRetryPolicy(http.StatusCodeRetryPolicy{Codes: []int{500, 502, 503}}),
    http.WithRetryCheck(func(req *http.Request) bool {
        return req.Method == http.MethodGet || req.Method == http.MethodPut || req.Method == http.MethodDelete
    }),
    http.WithMiddleware(http.Logging),
)

// GET request
body, err := client.Get(ctx, "https://api.example.com/data")

// POST JSON
data := map[string]string{"name": "John"}
body, err := client.PostJson(ctx, "https://api.example.com/users", data)

// Per-request options
body, err := client.Get(ctx, url, 
    http.WithRequestTimeout(10*time.Second),
    http.WithHeader("Authorization", "Bearer token"),
)
```

### Retry Policies

```go
// Timeout only
policy := http.TimeoutRetryPolicy{}

// Specific status codes
policy := http.StatusCodeRetryPolicy{Codes: []int{500, 502, 503}}

// Composite (OR logic)
policy := http.CompositeRetryPolicy{
    Policies: []http.RetryPolicy{
        http.TimeoutRetryPolicy{},
        http.StatusCodeRetryPolicy{Codes: []int{500}},
    },
}

// Always retry
policy := http.AlwaysRetryPolicy{}
```

### Middleware

```go
// Logging middleware (built-in)
client := http.New(http.WithMiddleware(http.Logging))

// Custom middleware
func AuthMiddleware(next http.RoundTripperFunc) http.RoundTripperFunc {
    return func(req *http.Request) (*http.Response, error) {
        req.Header.Set("Authorization", "Bearer token")
        return next(req)
    }
}

client := http.New(http.WithMiddleware(AuthMiddleware))

// Multiple middleware (executed in order)
client := http.New(
    http.WithMiddleware(http.Logging),
    http.WithMiddleware(AuthMiddleware),
    http.WithMiddleware(http.Metrics),
)
```

## Message Queue (`net/mq`)

### Features
- **In-Process Broker**: Lightweight adapter around `pubsub.PubSub`
- **Panic Recovery**: Safe guards around publish/subscribe/close
- **Metrics Hooks**: Observe broker activity and timing

### Quick Start

```go
import "github.com/spcent/plumego/net/mq"

broker := mq.NewInProcBroker(nil)
defer broker.Close()

sub, _ := broker.Subscribe(context.Background(), "events", mq.DefaultSubOptions())
defer sub.Cancel()

_ = broker.Publish(context.Background(), "events", mq.Message{ID: "1", Data: "payload"})
```

## Webhook In (`net/webhookin`)

### Features
- **Deduplication**: TTL-based idempotency gate
- **GitHub Verification**: X-Hub-Signature-256 validation
- **Stripe Verification**: Stripe-Signature header validation
- **Zero Dependencies**: Pure standard library

### Deduplication

```go
import "github.com/spcent/plumego/net/webhookin"

// Create deduper with 10 minute TTL
deduper := webhookin.NewDeduper(10 * time.Minute)

// Check if request was already processed
if deduper.SeenBefore(requestID) {
    // Already processed, return 200 OK
    return
}

// Process request...
```

### GitHub Webhook Verification

```go
func handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
    secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
    body, err := webhookin.VerifyGitHub(r, secret, 10<<20) // 10MB max
    if err != nil {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Process verified webhook
    var event github.Event
    json.Unmarshal(body, &event)
    // ...
}
```

### Stripe Webhook Verification

```go
func handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
    secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
    
    opts := webhookin.StripeVerifyOptions{
        MaxBody:   10 << 20, // 10MB
        Tolerance: 5 * time.Minute,
    }
    
    body, err := webhookin.VerifyStripe(r, secret, opts)
    if err != nil {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Process verified webhook
    var event stripe.Event
    json.Unmarshal(body, &event)
    // ...
}
```

## Webhook Out (`net/webhookout`)

### Features
- **Reliable Delivery**: Retry with exponential backoff
- **Dead Letter Queue**: Failed deliveries tracked
- **Replay Support**: Redeliver any failed delivery
- **Target Management**: Dynamic target configuration
- **Metrics**: Real-time delivery statistics
- **SSRF Protection**: Private network blocking

### Architecture

```
Event → Target Matching → Delivery Creation → Queue → Worker → HTTP Request
         ↓                    ↓                    ↓
    Store Target        Store Delivery       Retry/Backoff
```

### Quick Start

```go
import (
    "github.com/spcent/plumego/net/webhookout"
    "github.com/spcent/plumego/net/webhookout/store" // or your own store
)

// Create service
cfg := webhookout.Config{
    Enabled:          true,
    Workers:          4,
    QueueSize:        1000,
    DefaultTimeout:   10 * time.Second,
    DefaultMaxRetries: 3,
    BackoffBase:      1 * time.Second,
    BackoffMax:       60 * time.Second,
    RetryOn429:       true,
    AllowPrivateNetwork: false,
    DrainMax:         30 * time.Second,
}

store := store.NewMemoryStore() // or PostgreSQL, Redis, etc.
service := webhookout.NewService(store, cfg)

// Start workers
ctx := context.Background()
service.Start(ctx)
defer service.Stop()

// Create target
target := webhookout.Target{
    Name:    "My App",
    URL:     "https://myapp.com/webhook",
    Secret:  "super-secret-key",
    Events:  []string{"user.created", "order.paid"},
    Enabled: true,
    Headers: map[string]string{
        "X-API-Version": "v1",
    },
}

target, err := service.CreateTarget(ctx, target)

// Trigger event
event := webhookout.Event{
    Type: "user.created",
    Data: map[string]any{
        "user_id":   12345,
        "email":     "user@example.com",
        "timestamp": time.Now(),
    },
    Meta: map[string]any{
        "source": "web-app",
    },
}

count, err := service.TriggerEvent(ctx, event)
// Returns: number of targets that received the event
```

### Delivery Management

```go
// List deliveries
filter := webhookout.DeliveryFilter{
    Status: webhookout.DeliveryFailed,
    Limit:  50,
}
deliveries, err := service.ListDeliveries(ctx, filter)

// Get specific delivery
delivery, ok := service.GetDelivery(ctx, deliveryID)

// Replay failed delivery
newDelivery, err := service.ReplayDelivery(ctx, deliveryID)
```

### Target Management

```go
// Update target
patch := webhookout.TargetPatch{
    Enabled: boolPtr(false),
    URL:     stringPtr("https://new-url.com/webhook"),
}
target, err := service.UpdateTarget(ctx, targetID, patch)

// List targets
targets, err := service.ListTargets(ctx, webhookout.TargetFilter{
    Enabled: boolPtr(true),
    Event:   "user.created",
})
```

### Retry Logic

```go
// Automatic retry based on:
// - HTTP status codes (5xx, 429 if enabled)
// - Network errors
// - Timeout errors

// Backoff formula: base * 2^attempt * jitter[0.5, 1.5]
// Example: 1s * 2^1 * 1.2 = 2.4s
```

### Metrics

```go
metrics := service.Metrics()
fmt.Printf(`
Enqueued:  %d
Dropped:   %d
Sent OK:   %d
Sent Fail: %d
Retried:   %d
Dead:      %d
`, metrics.Enqueued, metrics.Dropped, metrics.SentOK, 
   metrics.SentFail, metrics.Retried, metrics.Dead)
```

### Security Features

```go
// 1. SSRF Protection (private IP blocking)
// Enabled by default, can be disabled:
cfg.AllowPrivateNetwork = true

// 2. Signature Verification
// Each webhook includes:
// X-Webhook-Id: target ID
// X-Webhook-Event: event type
// X-Webhook-Delivery: delivery ID
// X-Webhook-Timestamp: unix timestamp
// X-Webhook-Signature: v1=<hmac>

// 3. Replay Protection
// Timestamp tolerance check (configurable)

// 4. Payload Integrity
// Each delivery stores exact payload for replay
```

## WebSocket (`net/websocket`)

See [net/websocket/README.md](websocket/README.md) for detailed documentation.

## Performance Considerations

### HTTP Client
- **Connection Pool**: Reuse connections via `http.Client`
- **Timeout**: Set appropriate timeouts (default 3s)
- **Retry**: Use exponential backoff to avoid thundering herd
- **Middleware**: Keep middleware chain lightweight

### Webhook In
- **Deduplication**: TTL should match your processing window
- **Body Size**: Limit body size to prevent memory exhaustion
- **Verification**: Always verify signatures before processing

### Webhook Out
- **Workers**: 1-2 per CPU core for I/O bound work
- **Queue Size**: 100-1000 per worker for backpressure
- **Backoff**: Base 1-5s, Max 60-300s for production
- **Drain**: Allow 30-60s for graceful shutdown

## Testing

```bash
# Test all packages
go test ./net/... -v

# HTTP client tests
go test ./net/http -v -cover

# Webhook tests
go test ./net/webhookin -v
go test ./net/webhookout -v

# WebSocket tests
go test ./net/websocket -v
```

## Production Checklist

### HTTP Client
- [ ] Set appropriate timeouts
- [ ] Configure retry policies
- [ ] Add authentication middleware
- [ ] Monitor retry rates
- [ ] Set connection pool limits

### Webhook In
- [ ] Use strong secrets (32+ bytes)
- [ ] Verify all signatures
- [ ] Implement deduplication
- [ ] Limit body sizes
- [ ] Log verification failures

### Webhook Out
- [ ] Configure workers based on load
- [ ] Set queue size for backpressure
- [ ] Enable SSRF protection
- [ ] Monitor delivery metrics
- [ ] Implement alerting for dead deliveries
- [ ] Use persistent storage (not just memory)
- [ ] Set appropriate timeouts

### WebSocket
- [ ] Configure connection limits
- [ ] Set message size limits
- [ ] Implement authentication
- [ ] Monitor connection counts
- [ ] Use proper TLS in production

## License

All packages follow the project's license terms.
