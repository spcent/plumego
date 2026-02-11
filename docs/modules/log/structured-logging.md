# Structured Logging

> **Package**: `github.com/spcent/plumego/log` | **Best Practices**

## Overview

Structured logging outputs logs in a machine-readable format (typically JSON) with typed fields. This enables powerful log analysis, searching, and aggregation compared to traditional text-based logs.

**Benefits**:
- **Searchable**: Query by specific fields
- **Aggregatable**: Group and count by dimensions
- **Type-safe**: Preserve data types (numbers, booleans, timestamps)
- **Correlatable**: Link related events via request IDs
- **Alertable**: Trigger alerts based on specific field values

## Traditional vs Structured

### Traditional Text Logging

```go
// ❌ Text logging - hard to parse
log.Println("User user-123 logged in from 192.168.1.1 at 2025-01-15 10:30:00")
log.Println("Request GET /api/users took 125ms with status 200")
```

**Problems**:
- Requires regex parsing
- No type information (125 is a string)
- Inconsistent format
- Hard to search

### Structured Logging

```go
// ✅ Structured logging - easy to parse
log.Info("User logged in",
    "user_id", "user-123",
    "ip", "192.168.1.1",
    "timestamp", time.Now(),
)

log.Info("Request completed",
    "method", "GET",
    "path", "/api/users",
    "duration_ms", 125,
    "status", 200,
)
```

**Output** (JSON):
```json
{"level":"info","message":"User logged in","user_id":"user-123","ip":"192.168.1.1","timestamp":"2025-01-15T10:30:00Z"}
{"level":"info","message":"Request completed","method":"GET","path":"/api/users","duration_ms":125,"status":200}
```

**Benefits**:
- Direct field access (`duration_ms > 100`)
- Preserves types (125 is a number)
- Consistent structure
- Easy aggregation

## Field Naming Conventions

### Use Snake Case

```go
// ✅ Good: snake_case
log.Info("Request", "user_id", "123", "request_id", "req-456")

// ❌ Bad: camelCase
log.Info("Request", "userId", "123", "requestId", "req-456")

// ❌ Bad: PascalCase
log.Info("Request", "UserId", "123", "RequestId", "req-456")
```

### Consistent Field Names

Use the same field names across all logs:

```go
// ✅ Good: Always use "user_id"
log.Info("Login", "user_id", "123")
log.Info("Logout", "user_id", "123")
log.Info("Update profile", "user_id", "123")

// ❌ Bad: Inconsistent names
log.Info("Login", "user_id", "123")
log.Info("Logout", "userId", "123")
log.Info("Update profile", "user", "123")
```

### Standard Fields

Recommended standard field names:

| Field | Type | Description |
|-------|------|-------------|
| `request_id` | string | Unique request identifier |
| `user_id` | string | User identifier |
| `tenant_id` | string | Tenant identifier |
| `session_id` | string | Session identifier |
| `ip` | string | Client IP address |
| `method` | string | HTTP method (GET, POST, etc.) |
| `path` | string | Request path |
| `status` | int | HTTP status code |
| `duration_ms` | int/float | Duration in milliseconds |
| `error` | string | Error message |
| `timestamp` | time.Time | Event timestamp |

## Common Patterns

### 1. Request Logging

```go
func requestLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := getRequestID(r)

        // Create request-scoped logger
        logger := log.With(
            "request_id", requestID,
            "method", r.Method,
            "path", r.URL.Path,
            "remote_addr", r.RemoteAddr,
        )

        // Log request start
        logger.Info("Request started")

        // Wrap response writer to capture status
        wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
        next.ServeHTTP(wrapped, r)

        // Log request completion
        logger.Info("Request completed",
            "status", wrapped.statusCode,
            "duration_ms", time.Since(start).Milliseconds(),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
    w.statusCode = code
    w.ResponseWriter.WriteHeader(code)
}
```

### 2. Database Query Logging

```go
func logQuery(query string, args ...interface{}) func(error, time.Duration) {
    start := time.Now()

    return func(err error, duration time.Duration) {
        fields := []interface{}{
            "query", query,
            "duration_ms", duration.Milliseconds(),
        }

        if len(args) > 0 {
            fields = append(fields, "args", args)
        }

        if err != nil {
            fields = append(fields, "error", err.Error())
            log.Error("Database query failed", fields...)
        } else {
            log.Debug("Database query", fields...)
        }
    }
}

// Usage
logFn := logQuery("SELECT * FROM users WHERE id = ?", userID)
result, err := db.Query(...)
logFn(err, time.Since(start))
```

### 3. Business Event Logging

```go
// User registration
log.Info("User registered",
    "user_id", user.ID,
    "email", user.Email,
    "registration_method", "email", // or "oauth", "sso"
    "source", "web", // or "mobile", "api"
)

// Order placed
log.Info("Order placed",
    "order_id", order.ID,
    "user_id", order.UserID,
    "total_amount", order.Total,
    "currency", order.Currency,
    "items_count", len(order.Items),
)

// Payment processed
log.Info("Payment processed",
    "payment_id", payment.ID,
    "order_id", payment.OrderID,
    "amount", payment.Amount,
    "method", payment.Method, // "card", "paypal", etc.
    "status", payment.Status,
)
```

### 4. Error Context Logging

```go
func processOrder(orderID string) error {
    order, err := db.GetOrder(orderID)
    if err != nil {
        log.Error("Failed to fetch order",
            "order_id", orderID,
            "error", err,
            "operation", "get_order",
        )
        return err
    }

    if err := validateOrder(order); err != nil {
        log.Error("Order validation failed",
            "order_id", orderID,
            "error", err,
            "operation", "validate_order",
            "order_status", order.Status,
        )
        return err
    }

    if err := processPayment(order); err != nil {
        log.Error("Payment processing failed",
            "order_id", orderID,
            "error", err,
            "operation", "process_payment",
            "payment_method", order.PaymentMethod,
            "amount", order.Total,
        )
        return err
    }

    log.Info("Order processed successfully",
        "order_id", orderID,
        "total", order.Total,
    )

    return nil
}
```

### 5. Performance Monitoring

```go
func trackPerformance(operation string) func() {
    start := time.Now()
    var memBefore runtime.MemStats
    runtime.ReadMemStats(&memBefore)

    return func() {
        var memAfter runtime.MemStats
        runtime.ReadMemStats(&memAfter)

        log.Info("Operation completed",
            "operation", operation,
            "duration_ms", time.Since(start).Milliseconds(),
            "alloc_bytes", memAfter.Alloc-memBefore.Alloc,
        )
    }
}

// Usage
defer trackPerformance("process_batch")()
// ... perform operation ...
```

## Log Aggregation

### ELK Stack (Elasticsearch, Logstash, Kibana)

**Setup**:
1. Output JSON logs to stdout
2. Ship logs with Filebeat or Fluentd
3. Parse JSON in Logstash
4. Store in Elasticsearch
5. Visualize in Kibana

**Query examples** (Kibana):
```
# Find slow requests
duration_ms > 1000 AND status: 200

# Find errors for specific user
level: "error" AND user_id: "user-123"

# Count requests by path
# Visualization: Terms aggregation on path field
```

### AWS CloudWatch Logs Insights

**Query examples**:
```
# Find slow requests
fields @timestamp, request_id, duration_ms
| filter duration_ms > 1000
| sort duration_ms desc

# Error rate by path
fields path
| filter level = "error"
| stats count() by path

# P95 latency by endpoint
fields path, duration_ms
| stats pct(duration_ms, 95) by path
```

### Splunk

**Query examples**:
```
# Find errors in last hour
level="error" earliest=-1h

# Average response time by endpoint
| stats avg(duration_ms) by path

# Unique users per hour
| timechart span=1h dc(user_id)
```

## Correlation

### Request ID Propagation

```go
// Generate request ID
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = generateRequestID()
        }

        // Add to response header
        w.Header().Set("X-Request-ID", requestID)

        // Add to context
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Use in logs
logger := log.With("request_id", getRequestID(r))
logger.Info("Processing request")
```

### Distributed Tracing

```go
// Propagate trace context
func traceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        traceID := r.Header.Get("X-Trace-ID")
        spanID := generateSpanID()

        logger := log.With(
            "trace_id", traceID,
            "span_id", spanID,
        )

        ctx := context.WithValue(r.Context(), "logger", logger)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Best Practices

### 1. Log at Appropriate Levels

```go
// DEBUG: Detailed diagnostic info (development only)
log.Debug("Processing item",
    "item_id", item.ID,
    "item_data", item,
)

// INFO: General informational events
log.Info("User logged in", "user_id", userID)

// WARN: Potentially harmful situations
log.Warn("Cache miss, falling back to database",
    "key", cacheKey,
)

// ERROR: Error events that might allow app to continue
log.Error("Failed to send email",
    "user_id", userID,
    "error", err,
)
```

### 2. Include Context

```go
// ✅ Good: Includes context
log.Error("Payment failed",
    "payment_id", payment.ID,
    "order_id", payment.OrderID,
    "user_id", payment.UserID,
    "amount", payment.Amount,
    "error", err,
)

// ❌ Bad: Missing context
log.Error("Payment failed", "error", err)
```

### 3. Don't Log Sensitive Data

```go
// ❌ Never log these
log.Info("Login", "password", password)
log.Info("Card payment", "card_number", cardNumber)
log.Info("API call", "api_key", apiKey)

// ✅ Log identifiers only
log.Info("Login", "user_id", userID)
log.Info("Card payment", "last4", last4Digits)
log.Info("API call", "client_id", clientID)
```

### 4. Use Consistent Structure

```go
// ✅ Good: Consistent structure for all requests
log.Info("Request started", "method", "GET", "path", "/api/users", "status", 200, "duration_ms", 50)
log.Info("Request started", "method", "POST", "path", "/api/orders", "status", 201, "duration_ms", 150)

// ❌ Bad: Inconsistent structure
log.Info("GET /api/users completed in 50ms with status 200")
log.Info("Order created", "path", "/api/orders", "time", 150)
```

### 5. Aggregate Similar Events

```go
// ❌ Bad: Log every iteration
for _, item := range items {
    log.Info("Processing item", "item_id", item.ID)
}

// ✅ Good: Log summary
log.Info("Processing batch",
    "batch_size", len(items),
    "first_item", items[0].ID,
    "last_item", items[len(items)-1].ID,
)
```

## Performance Tips

### Conditional Debug Logging

```go
// ❌ Slow: Always allocates
log.Debug(fmt.Sprintf("User: %+v", user))

// ✅ Fast: Only if debug enabled
if log.IsDebugEnabled() {
    log.Debug("User data", "user", user)
}
```

### Reuse Loggers

```go
// ✅ Efficient: Create once, reuse
serviceLogger := log.With("service", "payment", "version", "1.0")

for _, payment := range payments {
    serviceLogger.Info("Processing payment", "id", payment.ID)
}
```

## Testing

### Verify Log Output

```go
func TestLogging(t *testing.T) {
    var buf bytes.Buffer
    log.SetOutput(&buf)
    defer log.SetOutput(os.Stdout)

    // Execute code that logs
    processOrder("order-123")

    // Verify log contains expected fields
    output := buf.String()

    var entry map[string]interface{}
    json.Unmarshal([]byte(output), &entry)

    if entry["order_id"] != "order-123" {
        t.Error("order_id not logged correctly")
    }
}
```

## Related Documentation

- [Log Overview](README.md) — Log module overview
- [Middleware: Observability](../middleware/observability.md) — Request logging
- [Metrics](../metrics/) — Metrics collection

## External Resources

- [The Twelve-Factor App: Logs](https://12factor.net/logs)
- [Structured Logging Best Practices](https://engineering.grab.com/structured-logging)
- [JSON Logging Best Practices](https://www.loggly.com/blog/why-json-is-the-best-application-log-format-and-how-to-switch/)
