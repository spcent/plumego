# API Reference

This document provides a comprehensive API reference for the plumego web toolkit.

## Table of Contents

- [Middleware](#middleware)
- [Security](#security)
- [WebSocket](#websocket)
- [Webhook](#webhook)
- [PubSub](#pubsub)
- [Observability](#observability)

---

## Middleware

### Middleware Chain

The middleware package provides composable HTTP middleware functions.

#### Available Middleware

- **Logging**: Request/response logging
- **Recovery**: Panic recovery
- **CORS**: Cross-Origin Resource Sharing
- **Gzip**: Response compression
- **Rate Limit**: Request rate limiting
- **Timeout**: Request timeout
- **Auth**: Authentication helpers
- **Security Headers**: Security-related HTTP headers
- **Debug Errors**: Error debugging
- **Response**: Response formatting

#### Example Usage

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware"
)

app := core.New()

// Add middleware to the chain
app.Use(middleware.Logging())
app.Use(middleware.Recovery())
app.Use(middleware.CORS(middleware.CORSConfig{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
}))
app.Use(middleware.Gzip())
app.Use(middleware.RateLimit(middleware.RateLimitConfig{
    RequestsPerSecond: 100,
}))
app.Use(middleware.Timeout(5 * time.Second))
app.Use(middleware.SecurityHeaders())

// Add route-specific middleware
app.Get("/api/users", middleware.Auth(), handler)
```

### Authentication Middleware

#### BasicAuth

Basic authentication middleware.

```go
import "github.com/spcent/plumego/middleware"

auth := middleware.BasicAuth(middleware.BasicAuthConfig{
    Username: "admin",
    Password: "secret",
})

app.Use(auth)
```

#### JWTAuth

JWT authentication middleware.

```go
import "github.com/spcent/plumego/middleware"

auth := middleware.JWTAuth(middleware.JWTAuthConfig{
    Secret: "your-secret-key",
    Header: "Authorization",
    Prefix: "Bearer ",
})

app.Use(auth)
```

### CORS Middleware

CORS (Cross-Origin Resource Sharing) middleware.

```go
import "github.com/spcent/plumego/middleware"

config := middleware.CORSConfig{
    AllowOrigins: []string{"https://example.com", "https://app.example.com"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders: []string{"Content-Type", "Authorization"},
    ExposeHeaders: []string{"X-Request-ID"},
    AllowCredentials: true,
    MaxAge: 86400,
}

app.Use(middleware.CORS(config))
```

### Rate Limiting Middleware

Rate limiting middleware with multiple algorithms.

```go
import "github.com/spcent/plumego/middleware"

// Token bucket rate limiting
config := middleware.RateLimitConfig{
    Algorithm:     middleware.TokenBucket,
    RequestsPerSecond: 100,
    BurstSize:     200,
    Key:           middleware.RemoteIPKey,
}

app.Use(middleware.RateLimit(config))
```

### Timeout Middleware

Request timeout middleware.

```go
import "github.com/spcent/plumego/middleware"

app.Use(middleware.Timeout(30 * time.Second))
```

### Security Headers Middleware

Security headers middleware.

```go
import "github.com/spcent/plumego/middleware"

config := middleware.SecurityHeadersConfig{
    XFrameOptions:         "DENY",
    XContentTypeOptions:   "nosniff",
    XSSProtection:         "1; mode=block",
    HSTS:                  "max-age=31536000; includeSubDomains",
    CSP:                   "default-src 'self'",
    ReferrerPolicy:        "strict-origin-when-cross-origin",
    PermissionsPolicy:     "geolocation=(), microphone=()",
}

app.Use(middleware.SecurityHeaders(config))
```

---

## Security

### Webhook Signature Verification

Verify webhook signatures for security.

```go
import "github.com/spcent/plumego/security"

// Verify GitHub webhook signature
signature := r.Header.Get("X-Hub-Signature-256")
payload, _ := io.ReadAll(r.Body)

valid := security.VerifyGitHubWebhook(payload, signature, "your-secret")
if !valid {
    http.Error(w, "Invalid signature", http.StatusUnauthorized)
    return
}
```

### Input Validation

Validate user input to prevent injection attacks.

```go
import "github.com/spcent/plumego/security/input"

// Validate email
email := "user@example.com"
if !input.ValidateEmail(email) {
    return errors.New("invalid email")
}

// Validate URL
url := "https://example.com"
if !input.ValidateURL(url) {
    return errors.New("invalid URL")
}

// Sanitize HTML
html := "<script>alert('xss')</script>"
clean := input.SanitizeHTML(html)
```

### Rate Limiting (Security)

Prevent abuse with rate limiting.

```go
import "github.com/spcent/plumego/security/abuse"

limiter := abuse.NewLimiter(abuse.Config{
    MaxRequests: 100,
    Window:      time.Minute,
    Burst:       20,
})

// Check if request is allowed
if !limiter.Allow(r.RemoteAddr) {
    http.Error(w, "Too many requests", http.StatusTooManyRequests)
    return
}
```

### Password Strength

Validate password strength.

```go
import "github.com/spcent/plumego/security/password"

// Check password strength
strength := password.CheckStrength("MySecurePassword123!")
if strength < password.Strong {
    return errors.New("password too weak")
}

// Hash password
hashed, err := password.Hash("MySecurePassword123!")
if err != nil {
    return err
}

// Verify password
valid := password.Verify("MySecurePassword123!", hashed)
```

### JWT Token

Generate and verify JWT tokens.

```go
import "github.com/spcent/plumego/security/jwt"

// Generate token
token, err := jwt.Generate(jwt.Payload{
    UserID:   "123",
    Email:    "user@example.com",
    ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
}, "your-secret-key")

// Verify token
payload, err := jwt.Verify(token, "your-secret-key")
if err != nil {
    // Invalid token
}
```

---

## WebSocket

### WebSocket Hub

Manage WebSocket connections with room-based broadcasting.

```go
import "github.com/spcent/plumego/net/websocket"

// Create hub with default configuration
hub := websocket.NewHub(4, 1024)
defer hub.Stop()

// Or with custom configuration
config := websocket.HubConfig{
    WorkerCount:            8,
    JobQueueSize:           2048,
    MaxConnections:         50000,
    MaxRoomConnections:     500,
    EnableDebugLogging:     true,
    EnableMetrics:          true,
    RejectOnQueueFull:      true,
    MaxConnectionRate:      200,
    EnableSecurityMetrics:  true,
}
hub := websocket.NewHubWithConfig(config)
defer hub.Stop()
```

### WebSocket Connection

Handle WebSocket connections.

```go
import "github.com/spcent/plumego/net/websocket"

// Upgrade HTTP to WebSocket
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := websocket.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Upgrade failed: %v", err)
        return
    }
    defer conn.Close()

    // Join to room
    err = hub.TryJoin("chat-room", conn)
    if err != nil {
        log.Printf("Failed to join room: %v", err)
        return
    }
    defer hub.Leave("chat-room", conn)

    // Read messages
    for {
        op, data, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Read error: %v", err)
            return
        }

        // Broadcast to room
        hub.BroadcastRoom("chat-room", op, data)
    }
}
```

### Broadcasting Messages

Broadcast messages to rooms.

```go
import "github.com/spcent/plumego/net/websocket"

// Broadcast to a room
hub.BroadcastRoom("chat-room", websocket.OpcodeText, []byte("Hello everyone!"))

// Broadcast to all rooms
hub.BroadcastAll(websocket.OpcodeText, []byte("System announcement"))

// Send to specific connection
err := conn.WriteMessage(websocket.OpcodeText, []byte("Hello"))
```

### Room Management

Manage rooms and connections.

```go
import "github.com/spcent/plumego/net/websocket"

// Check if room has capacity
if err := hub.CanJoin("chat-room"); err != nil {
    // Room is full
}

// Get room count
count := hub.GetRoomCount("chat-room")
fmt.Printf("Room has %d connections\n", count)

// Get all rooms
rooms := hub.GetRooms()
for _, room := range rooms {
    fmt.Printf("Room: %s, Connections: %d\n", room, hub.GetRoomCount(room))
}

// Get connections in room
connections := hub.GetConnectionsInRoom("chat-room")
for _, conn := range connections {
    conn.WriteMessage(websocket.OpcodeText, []byte("Hello"))
}
```

### Metrics

Monitor WebSocket hub metrics.

```go
import "github.com/spcent/plumego/net/websocket"

metrics := hub.Metrics()
fmt.Printf("Active: %d, Rooms: %d\n", metrics.ActiveConnections, metrics.Rooms)
```

### Security Events

Monitor security events.

```go
import "github.com/spcent/plumego/net/websocket"

events := hub.SecurityEvents()
go func() {
    for event := range events {
        log.Printf("Security event: %s - %v", event.Type, event.Details)
    }
}()
```

---

## Webhook

### Webhook Service

Manage webhook delivery with retry logic.

```go
import "github.com/spcent/plumego/net/webhookout"

// Create store
store := webhookout.NewMemStore()

// Create service with default configuration
config := webhookout.DefaultConfig()
service := webhookout.NewService(store, config)
defer service.Stop()

// Start service
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go service.Start(ctx)
```

### Webhook Targets

Manage webhook targets.

```go
import "github.com/spcent/plumego/net/webhookout"

// Create target
target := webhookout.Target{
    Name:    "Payment Service",
    URL:     "https://api.example.com/webhooks",
    Secret:  "secret-key-here",
    Events:  []string{"payment.success", "payment.failed"},
    Enabled: true,
    Headers: map[string]string{
        "X-API-Version": "v1",
    },
    TimeoutMs:     5000,
    MaxRetries:    3,
    BackoffBaseMs: 1000,
    BackoffMaxMs:  30000,
    RetryOn429:    boolPtr(true),
}

created, err := service.CreateTarget(context.Background(), target)
if err != nil {
    log.Fatalf("Failed to create target: %v", err)
}

// Update target
enabled := false
patch := webhookout.TargetPatch{
    Enabled: &enabled,
}
updated, err := service.UpdateTarget(context.Background(), created.ID, patch)

// Get target
target, ok := service.GetTarget(context.Background(), created.ID)
if !ok {
    // Target not found
}

// List targets
enabled := true
filter := webhookout.TargetFilter{
    Enabled: &enabled,
    Event:   "payment.success",
}
targets, err := service.ListTargets(context.Background(), filter)
```

### Triggering Events

Trigger webhook events to all subscribed targets.

```go
import "github.com/spcent/plumego/net/webhookout"

event := webhookout.Event{
    Type: "payment.success",
    Data: map[string]any{
        "amount":      100.00,
        "currency":    "USD",
        "customer_id": "cust-456",
        "order_id":    "ord-789",
    },
    Meta: map[string]any{
        "source": "api",
    },
}

count, err := service.TriggerEvent(context.Background(), event)
if err != nil {
    log.Fatalf("Failed to trigger event: %v", err)
}
log.Printf("Delivered to %d targets", count)
```

### Delivery Management

Manage webhook deliveries.

```go
import "github.com/spcent/plumego/net/webhookout"

// Get delivery
delivery, ok := service.GetDelivery(context.Background(), "del-123")
if !ok {
    // Delivery not found
}

// List deliveries
filter := webhookout.DeliveryFilter{
    Limit: 100,
}
deliveries, err := service.ListDeliveries(context.Background(), filter)

// Replay failed delivery
newDelivery, err := service.ReplayDelivery(context.Background(), "del-123")
if err != nil {
    // Handle error
}
```

### Metrics

Monitor webhook delivery metrics.

```go
import "github.com/spcent/plumego/net/webhookout"

metrics := service.Metrics()
fmt.Printf("Enqueued: %d, SentOK: %d, SentFail: %d\n",
    metrics.Enqueued, metrics.SentOK, metrics.SentFail)
```

### Configuration

Configure webhook service.

```go
import "github.com/spcent/plumego/net/webhookout"

config := webhookout.Config{
    Enabled:             true,
    QueueSize:           4096,
    Workers:             16,
    DrainMax:            10 * time.Second,
    DropPolicy:          webhookout.BlockWithLimit,
    BlockWait:           100 * time.Millisecond,
    DefaultTimeout:      10 * time.Second,
    DefaultMaxRetries:   8,
    BackoffBase:         1 * time.Second,
    BackoffMax:          60 * time.Second,
    RetryOn429:          true,
    AllowPrivateNetwork: false,
}

// Validate configuration
if err := config.Validate(); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}

// Load from environment
config := webhookout.ConfigFromEnv()
```

---

## PubSub

### In-Process PubSub

Lightweight in-process publish/subscribe system.

```go
import "github.com/spcent/plumego/pubsub"

// Create pubsub instance
ps := pubsub.New()
defer ps.Close()
```

### Subscriptions

Subscribe to topics.

```go
import "github.com/spcent/plumego/pubsub"

// Subscribe to a topic
sub, err := ps.Subscribe("user.created", pubsub.DefaultSubOptions())
if err != nil {
    log.Fatalf("Failed to subscribe: %v", err)
}
defer sub.Cancel()

// Receive messages
for msg := range sub.C() {
    fmt.Printf("Received: %v\n", msg.Data)
}
```

### Pattern Subscriptions

Subscribe to topic patterns.

```go
import "github.com/spcent/plumego/pubsub"

// Subscribe to all user events
sub, err := ps.SubscribePattern("user.*", pubsub.DefaultSubOptions())
if err != nil {
    log.Fatalf("Failed to subscribe: %v", err)
}
defer sub.Cancel()

// Will receive messages from:
// - user.created
// - user.updated
// - user.deleted
for msg := range sub.C() {
    fmt.Printf("Pattern match: %v\n", msg.Data)
}
```

### Publishing

Publish messages to topics.

```go
import "github.com/spcent/plumego/pubsub"

msg := pubsub.Message{
    Data: map[string]any{
        "user_id": "123",
        "email":   "user@example.com",
    },
}

err := ps.Publish("user.created", msg)
if err != nil {
    log.Fatalf("Failed to publish: %v", err)
}

// Fire-and-forget
err = ps.PublishAsync("user.created", msg)
```

### Subscription Options

Configure subscription behavior.

```go
import "github.com/spcent/plumego/pubsub"

opts := pubsub.SubOptions{
    BufferSize:   32,
    Policy:       pubsub.BlockWithTimeout,
    BlockTimeout: 100 * time.Millisecond,
}

sub, err := ps.Subscribe("topic", opts)
```

### Backpressure Policies

Handle slow subscribers with backpressure policies.

```go
import "github.com/spcent/plumego/pubsub"

// Drop oldest messages when buffer is full
opts := pubsub.SubOptions{
    BufferSize: 16,
    Policy:     pubsub.DropOldest,
}

// Drop newest messages when buffer is full
opts := pubsub.SubOptions{
    BufferSize: 16,
    Policy:     pubsub.DropNewest,
}

// Block until timeout
opts := pubsub.SubOptions{
    BufferSize:   16,
    Policy:       pubsub.BlockWithTimeout,
    BlockTimeout: 50 * time.Millisecond,
}

// Close subscription when buffer is full
opts := pubsub.SubOptions{
    BufferSize: 16,
    Policy:     pubsub.CloseSubscriber,
}
```

### Metrics

Monitor pubsub metrics.

```go
import "github.com/spcent/plumego/pubsub"

snapshot := ps.Snapshot()
for topic, metrics := range snapshot.Topics {
    fmt.Printf("Topic: %s\n", topic)
    fmt.Printf("  Publish total: %d\n", metrics.PublishTotal)
    fmt.Printf("  Delivered total: %d\n", metrics.DeliveredTotal)
    fmt.Printf("  Subscribers: %d\n", metrics.SubscribersGauge)
}
```

### Topic Management

Manage topics and subscriptions.

```go
import "github.com/spcent/plumego/pubsub"

// Get subscriber count
count := ps.GetSubscriberCount("user.created")
fmt.Printf("Topic has %d subscribers\n", count)

// Get pattern subscriber count
count = ps.GetPatternSubscriberCount("user.*")
fmt.Printf("Pattern has %d subscribers\n", count)

// List topics
topics := ps.ListTopics()
for _, topic := range topics {
    fmt.Printf("Topic: %s\n", topic)
}

// List patterns
patterns := ps.ListPatterns()
for _, pattern := range patterns {
    fmt.Printf("Pattern: %s\n", pattern)
}
```

### Metrics Collector

Integrate with metrics collectors.

```go
import (
    "github.com/spcent/plumego/pubsub"
    "github.com/spcent/plumego/metrics"
)

collector := metrics.NewPrometheusCollector("myapp")
ps.SetMetricsCollector(collector)
```

---

## Observability

### OpenTelemetry Tracer

OpenTelemetry-compatible tracing without external dependencies.

```go
import "github.com/spcent/plumego/metrics"

tracer := metrics.NewOpenTelemetryTracer("my-service")
defer tracer.Clear()

// Use in HTTP middleware
mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    ctx, span := tracer.Start(context.Background(), r)
    defer span.End(metrics.RequestMetrics{
        Status:  http.StatusOK,
        Bytes:   100,
        TraceID: span.TraceID(),
    })
    // ... handle request ...
})
```

### Prometheus Collector

Prometheus-compatible metrics collector.

```go
import "github.com/spcent/plumego/metrics"

collector := metrics.NewPrometheusCollector("myapp")
defer collector.Clear()

// Expose metrics endpoint
http.Handle("/metrics", collector.Handler())

// Record metrics
collector.ObserveHTTP(context.Background(), "GET", "/api/users", 200, 100, 50*time.Millisecond)
```

### Metrics

Monitor application metrics.

```go
import "github.com/spcent/plumego/metrics"

// Prometheus collector
collector := metrics.NewPrometheusCollector("myapp")
stats := collector.GetStats()
fmt.Printf("Total requests: %d, Average latency: %.3f\n",
    stats.TotalRequests, stats.AverageLatency)

// OpenTelemetry tracer
tracer := metrics.NewOpenTelemetryTracer("my-service")
spanStats := tracer.GetSpanStats()
fmt.Printf("Total spans: %d, Error spans: %d\n",
    spanStats.TotalSpans, spanStats.ErrorSpans)
```

### Health Checks

Health and readiness endpoints.

```go
import "github.com/spcent/plumego/health"

// Create health checker
checker := health.NewChecker()

// Add checks
checker.AddCheck("database", func() error {
    return db.Ping()
})

checker.AddCheck("cache", func() error {
    return cache.Ping()
})

// HTTP handler
http.Handle("/health", checker.Handler())
```

---

## Configuration

### Environment Variables

Configure applications using environment variables.

```go
import "github.com/spcent/plumego/config"

// Get values
port := config.GetString("PORT", "8080")
debug := config.GetBool("DEBUG", false)
timeout := config.GetDurationMs("TIMEOUT_MS", 5000)

// Load from .env file
config.LoadEnvFile(".env")
```

### Validation

Validate configuration values.

```go
import "github.com/spcent/plumego/config"

type Config struct {
    Port     int    `validate:"required,min=1,max=65535"`
    Host     string `validate:"required,hostname"`
    Timeout  int    `validate:"required,min=1"`
}

config := Config{
    Port:    8080,
    Host:    "localhost",
    Timeout: 5000,
}

if err := config.Validate(); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}
```

---

## Error Handling

### Error Utilities

Error handling utilities.

```go
import "github.com/spcent/plumego/contract"

// Create error with context
err := contract.Errorf("user_not_found", "User %s not found", userID)

// Check error type
if errors.Is(err, contract.ErrNotFound) {
    // Handle not found
}

// Add context to error
err = contract.WithContext(err, map[string]any{
    "user_id": userID,
    "action":  "get_user",
})
```

---

## Context

### Extended Context

Extended context with trace information.

```go
import "github.com/spcent/plumego/contract"

// Get trace ID from context
traceID := contract.TraceIDFromContext(ctx)

// Get span ID from context
spanID := contract.SpanIDFromContext(ctx)

// Get trace context
traceCtx := contract.TraceContextFromContext(ctx)
if traceCtx != nil {
    fmt.Printf("Trace: %s, Span: %s\n", traceCtx.TraceID, traceCtx.SpanID)
}
```

---

## Examples

### Complete Webhook Example

See [`examples/webhook_example.go`](examples/webhook_example.go) for a complete webhook example.

### Complete PubSub Example

See [`examples/pubsub_example.go`](examples/pubsub_example.go) for a complete pubsub example.

---

## Best Practices

### 1. Use Context for Cancellation

Always use context for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Use ctx in your operations
```

### 2. Handle Errors Gracefully

Always check and handle errors:

```go
if err != nil {
    log.Printf("Error: %v", err)
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    return
}
```

### 3. Close Resources

Always close resources when done:

```go
defer conn.Close()
defer rows.Close()
defer file.Close()
```

### 4. Use Connection Pools

For database connections, use connection pools:

```go
db, err := sql.Open("postgres", dsn)
if err != nil {
    log.Fatal(err)
}
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 5. Monitor Metrics

Monitor application metrics in production:

```go
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        metrics := service.Metrics()
        log.Printf("Metrics: %+v", metrics)
    }
}()
```

### 6. Use Graceful Shutdown

Always use graceful shutdown:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go service.Start(ctx)

// Wait for shutdown signal
<-ctx.Done()
service.Stop()
```

---

## License

This documentation is part of the plumego project and is licensed under the same terms as the project itself.
