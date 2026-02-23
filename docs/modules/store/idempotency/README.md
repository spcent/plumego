# Idempotency Module

> **Package**: `github.com/spcent/plumego/store/idempotency` | **Feature**: Duplicate request prevention

## Overview

The `idempotency` package prevents duplicate processing of API requests. When a client retries a request (network failure, timeout), the same operation is not executed twice — the cached response from the first successful execution is returned instead.

**Common use cases**:
- Payment processing (prevent double charges)
- Order creation (prevent duplicate orders)
- Email sending (prevent duplicate emails)
- Financial transfers

## How It Works

```
Client sends request with Idempotency-Key: key-abc
                │
                ▼
         Check idempotency store
                │
        ┌───────┴───────┐
     Not found       Found (duplicate)
        │                  │
        ▼                  ▼
   Process request    Return cached response
        │
        ▼
   Store response with key-abc
        │
        ▼
   Return response to client
```

## Quick Start

```go
import "github.com/spcent/plumego/store/idempotency"

// Create idempotency store (KV-backed)
store := idempotency.NewKVStore(kvStore,
    idempotency.WithTTL(24*time.Hour), // Keys expire after 24h
    idempotency.WithKeyHeader("Idempotency-Key"),
)

// Apply middleware to payment endpoint
app.Post("/api/payments", idempotency.Middleware(store)(handlePayment))
```

## Middleware

The idempotency middleware:
1. Reads `Idempotency-Key` header
2. Checks if key has been seen before
3. If yes: returns cached response
4. If no: processes request and caches response

```go
// Apply to specific routes
app.Post("/api/orders", idempotency.Middleware(store)(handleCreateOrder))
app.Post("/api/payments", idempotency.Middleware(store)(handlePayment))
app.Post("/api/subscriptions", idempotency.Middleware(store)(handleSubscribe))

// Response headers set by middleware
Idempotent-Replayed: true  // Present if this is a replayed response
```

## Storage Backends

### KV-backed Store

```go
kvStore, _ := kv.Open("./data/idempotency.kv")
store := idempotency.NewKVStore(kvStore,
    idempotency.WithTTL(24 * time.Hour),
)
```

### SQL-backed Store

```go
store := idempotency.NewSQLStore(database,
    idempotency.WithTable("idempotency_keys"),
    idempotency.WithTTL(7 * 24 * time.Hour), // 7 days
)
```

**Schema**:
```sql
CREATE TABLE idempotency_keys (
    idempotency_key VARCHAR(255) PRIMARY KEY,
    request_hash    VARCHAR(64)  NOT NULL,
    response_status INT          NOT NULL,
    response_body   BLOB         NOT NULL,
    created_at      TIMESTAMP    DEFAULT NOW(),
    expires_at      TIMESTAMP    NOT NULL,

    INDEX idx_idempotency_expires (expires_at)
);
```

## Client Usage

Clients should generate a unique idempotency key per logical operation and reuse it on retries:

```javascript
// Generate once, reuse on retry
const idempotencyKey = crypto.randomUUID();

async function createPayment(amount) {
    const response = await fetch('/api/payments', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Idempotency-Key': idempotencyKey,  // Same key on retry
        },
        body: JSON.stringify({ amount }),
    });
    return response.json();
}
```

## Manual Usage (Without Middleware)

For fine-grained control:

```go
func handlePayment(w http.ResponseWriter, r *http.Request) {
    idempotencyKey := r.Header.Get("Idempotency-Key")
    if idempotencyKey == "" {
        http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
        return
    }

    // Check for existing result
    result, err := idempotencyStore.Get(r.Context(), idempotencyKey)
    if err == nil {
        // Replay cached response
        w.Header().Set("Idempotent-Replayed", "true")
        w.WriteHeader(result.Status)
        w.Write(result.Body)
        return
    }

    // Process payment
    payment, err := paymentService.Charge(r.Context(), chargeRequest)
    if err != nil {
        http.Error(w, "Payment failed", http.StatusUnprocessableEntity)
        return
    }

    // Build response
    responseBody, _ := json.Marshal(payment)
    status := http.StatusCreated

    // Cache response (even on success or known failure)
    idempotencyStore.Set(r.Context(), idempotencyKey, idempotency.Result{
        Status: status,
        Body:   responseBody,
    })

    w.WriteHeader(status)
    w.Write(responseBody)
}
```

## Request Hash Validation

Detect when a client uses the same idempotency key with different request bodies (which is a client error):

```go
store := idempotency.NewKVStore(kvStore,
    idempotency.WithRequestValidation(true), // Hash request body
)

// If key exists but request body differs:
// Returns 422 Unprocessable Entity with error:
// {"error": "Idempotency key used with different request body"}
```

## Concurrency Handling

The idempotency store handles concurrent duplicate requests (two simultaneous retries):

```go
// Store uses locking to prevent race conditions
store := idempotency.NewKVStore(kvStore,
    idempotency.WithLocking(true),           // Distributed lock
    idempotency.WithLockTimeout(30*time.Second), // Max processing time
)

// If duplicate arrives while first is still processing:
// Returns 409 Conflict:
// {"error": "Request is being processed", "retry_after": 1}
```

## Cleanup

```go
// Scheduled cleanup of expired keys
scheduler.AddCron("idempotency-cleanup", "0 3 * * *", // 3 AM daily
    func(ctx context.Context) error {
        return idempotencyStore.DeleteExpired(ctx)
    },
)
```

## Best Practices

### 1. Scope Keys to User/Tenant

```go
// Prevent key collisions across users
scopedKey := tenantID + ":" + userID + ":" + clientKey
```

### 2. Set Appropriate TTL

```go
// Match your retry window
// If clients retry within 24h, set TTL to at least 24h
idempotency.WithTTL(24 * time.Hour)
```

### 3. Cache Both Success and Failure

```go
// Cache client errors too (4xx responses)
// This prevents reprocessing invalid requests
if status >= 400 && status < 500 {
    idempotencyStore.Set(ctx, key, idempotency.Result{
        Status: status,
        Body:   errorBody,
    })
}
```

### 4. Return 422 for Key Conflicts

```go
// If same key used with different body: unprocessable entity
if requestHashMismatch {
    http.Error(w, "Idempotency key conflict", http.StatusUnprocessableEntity)
    return
}
```

## Related Documentation

- [KV Store](../kv/README.md) — Storage backend
- [Database](../db/README.md) — SQL storage backend
- [Scheduler](../../scheduler/README.md) — Cleanup scheduling
