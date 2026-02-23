# Quota Management

> **Package**: `github.com/spcent/plumego/tenant` | **Feature**: Resource quotas per tenant

## Overview

Quota management enforces resource limits per tenant, preventing any single tenant from consuming disproportionate resources. Quotas can limit API requests, storage, users, and custom business metrics.

## QuotaManager Interface

```go
type QuotaManager interface {
    CheckQuota(ctx context.Context, tenantID, resource string, amount int) (QuotaDecision, error)
    RecordUsage(ctx context.Context, tenantID, resource string, amount int) error
    GetUsage(ctx context.Context, tenantID, resource string) (int, error)
}

type QuotaDecision struct {
    TenantID  string
    Resource  string
    Allowed   bool
    Used      int
    Limit     int
    Remaining int
    ResetAt   time.Time
}
```

## Setup

### In-Memory Quota Manager

```go
configMgr := tenant.NewInMemoryConfigManager()
quotaMgr := tenant.NewInMemoryQuotaManager(configMgr)

// Integrate with app
app := core.New(
    core.WithTenantConfigManager(configMgr),
    core.WithTenantMiddleware(core.TenantMiddlewareOptions{
        HeaderName:   "X-Tenant-ID",
        QuotaManager: quotaMgr,
        Hooks: tenant.Hooks{
            OnQuota: func(ctx context.Context, d tenant.QuotaDecision) {
                log.Warn("Quota exceeded",
                    "tenant", d.TenantID,
                    "resource", d.Resource,
                    "used", d.Used,
                    "limit", d.Limit,
                )
            },
        },
    }),
)
```

## Checking Quotas

### Request Quota Check

```go
func handleAPIRequest(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())

    // Check request quota
    decision, err := quotaMgr.CheckQuota(r.Context(), tenantID, "requests_per_minute", 1)
    if err != nil {
        http.Error(w, "Quota check failed", http.StatusInternalServerError)
        return
    }

    if !decision.Allowed {
        w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", decision.Limit))
        w.Header().Set("X-RateLimit-Remaining", "0")
        w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", decision.ResetAt.Unix()))
        w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(decision.ResetAt).Seconds())))
        http.Error(w, "Request quota exceeded", http.StatusTooManyRequests)
        return
    }

    // Record usage
    quotaMgr.RecordUsage(r.Context(), tenantID, "requests_per_minute", 1)

    // Handle request...
}
```

### Storage Quota Check

```go
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())

    // Get file size
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    fileSizeMB := int(header.Size / (1024 * 1024))

    // Check storage quota
    decision, err := quotaMgr.CheckQuota(r.Context(), tenantID, "storage_mb", fileSizeMB)
    if err != nil || !decision.Allowed {
        http.Error(w, fmt.Sprintf("Storage quota exceeded. Used: %d MB, Limit: %d MB",
            decision.Used, decision.Limit), http.StatusForbidden)
        return
    }

    // Save file and record usage
    if err := saveFile(file, header.Filename); err != nil {
        http.Error(w, "Upload failed", http.StatusInternalServerError)
        return
    }

    quotaMgr.RecordUsage(r.Context(), tenantID, "storage_mb", fileSizeMB)

    json.NewEncoder(w).Encode(map[string]string{"status": "uploaded"})
}
```

### User Quota Check

```go
func handleCreateUser(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())

    // Check user quota before creating
    decision, err := quotaMgr.CheckQuota(r.Context(), tenantID, "users", 1)
    if err != nil || !decision.Allowed {
        http.Error(w, fmt.Sprintf("User limit reached (%d/%d). Upgrade your plan.",
            decision.Used, decision.Limit), http.StatusForbidden)
        return
    }

    // Create user
    userID, err := createUser(r.Context(), tenantID, req)
    if err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

    // Record usage
    quotaMgr.RecordUsage(r.Context(), tenantID, "users", 1)

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
}
```

## Quota Resources

### Standard Resources

| Resource | Unit | Typical Limits |
|----------|------|---------------|
| `requests_per_minute` | requests | 60–10,000 |
| `requests_per_day` | requests | 1,000–1,000,000 |
| `users` | count | 3–unlimited |
| `storage_mb` | megabytes | 100–unlimited |
| `api_calls` | count | varies |
| `exports` | count | 10–unlimited |

### Custom Resources

```go
// Check custom business metric
decision, err := quotaMgr.CheckQuota(ctx, tenantID, "ai_tokens", tokensUsed)
if !decision.Allowed {
    return errors.New("AI token quota exceeded")
}
quotaMgr.RecordUsage(ctx, tenantID, "ai_tokens", tokensUsed)
```

## Usage Dashboard

### Get Current Usage

```go
func handleGetUsage(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())

    resources := []string{"requests_per_minute", "users", "storage_mb"}
    usage := make(map[string]interface{})

    for _, resource := range resources {
        used, _ := quotaMgr.GetUsage(r.Context(), tenantID, resource)
        config, _ := configMgr.GetTenantConfig(r.Context(), tenantID)

        limit := getLimit(config.Quota, resource)
        usage[resource] = map[string]interface{}{
            "used":      used,
            "limit":     limit,
            "remaining": limit - used,
            "percent":   float64(used) / float64(limit) * 100,
        }
    }

    json.NewEncoder(w).Encode(usage)
}
```

**Response**:
```json
{
  "requests_per_minute": {"used": 450, "limit": 1000, "remaining": 550, "percent": 45.0},
  "users": {"used": 23, "limit": 100, "remaining": 77, "percent": 23.0},
  "storage_mb": {"used": 2048, "limit": 51200, "remaining": 49152, "percent": 4.0}
}
```

## Quota Response Headers

Add standard headers to all responses:

```go
func addQuotaHeaders(w http.ResponseWriter, decision tenant.QuotaDecision) {
    w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", decision.Limit))
    w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", decision.Remaining))
    w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", decision.ResetAt.Unix()))
}
```

## Best Practices

### 1. Check Before Processing

```go
// ✅ Good: Check quota before expensive operations
decision, _ := quotaMgr.CheckQuota(ctx, tenantID, "ai_tokens", estimatedTokens)
if !decision.Allowed {
    return ErrQuotaExceeded
}
// Now do the expensive AI call
```

### 2. Return Helpful Errors

```go
if !decision.Allowed {
    resetIn := int(time.Until(decision.ResetAt).Seconds())
    http.Error(w, fmt.Sprintf(
        "Quota exceeded. Used: %d/%d. Resets in %d seconds. Consider upgrading your plan.",
        decision.Used, decision.Limit, resetIn,
    ), http.StatusTooManyRequests)
}
```

### 3. Alert Before Hitting Limits

```go
// Warn at 80% usage
if float64(decision.Used)/float64(decision.Limit) > 0.8 {
    notifications.SendQuotaWarning(tenantID, decision.Resource, decision.Used, decision.Limit)
}
```

## Related Documentation

- [Tenant Overview](README.md) — Multi-tenancy overview
- [Configuration](configuration.md) — Quota configuration
- [Rate Limiting](ratelimit.md) — Request rate limiting
- [Examples](examples.md) — Complete quota examples
