# AI Rate Limiting

> **Package**: `github.com/spcent/plumego/ai/ratelimit` | **Feature**: Quota management

## Overview

The `ratelimit` package provides AI-specific rate limiting based on token consumption, request counts, and concurrent connections — going beyond simple request-per-second limits to enable fair usage enforcement across users and tenants.

## Rate Limit Strategies

| Strategy | Best For |
|----------|----------|
| Requests per minute | Simple API throttling |
| Tokens per minute | Cost-based fair usage |
| Tokens per day | Daily budget enforcement |
| Concurrent requests | Resource protection |

## Quick Start

```go
import "github.com/spcent/plumego/ai/ratelimit"

// Token-based rate limiter
limiter := ratelimit.NewTokenLimiter(
    ratelimit.WithTokensPerMinute(100000),  // 100k tokens/min per user
    ratelimit.WithTokensPerDay(1000000),    // 1M tokens/day per user
    ratelimit.WithRequestsPerMinute(60),    // 60 requests/min per user
    ratelimit.WithStorage(redisStore),      // Distributed storage
)

// Apply to AI routes
ai := app.Router().Group("/api/ai")
ai.Use(limiter.Middleware(extractUserID))
```

## Token-Based Limiting

```go
func handleChat(w http.ResponseWriter, r *http.Request) {
    userID := auth.UserIDFromContext(r.Context())

    var req ChatRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Estimate input tokens
    inputTokens, _ := tokenizer.CountTokens(req.Message)
    maxOutput := 2048

    // Check budget before sending to LLM
    decision, err := limiter.CheckTokenBudget(r.Context(), userID, inputTokens+maxOutput)
    if err != nil || !decision.Allowed {
        w.Header().Set("X-RateLimit-Limit", strconv.Itoa(decision.Limit))
        w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
        w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(decision.ResetAt.Unix(), 10))
        w.Header().Set("Retry-After", strconv.Itoa(int(time.Until(decision.ResetAt).Seconds())))
        http.Error(w, "Token quota exceeded", http.StatusTooManyRequests)
        return
    }

    // Call LLM
    resp, err := claude.Complete(r.Context(), &provider.CompletionRequest{
        Model:    "claude-opus-4-6",
        Messages: buildMessages(sess, req.Message),
        MaxTokens: maxOutput,
    })

    // Record actual usage
    limiter.RecordUsage(r.Context(), userID, tokenizer.TokenUsage{
        Input:  resp.Usage.InputTokens,
        Output: resp.Usage.OutputTokens,
        Total:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
    })

    json.NewEncoder(w).Encode(ChatResponse{Response: resp.Content[0].Text})
}
```

## Per-Plan Limits

```go
var planLimits = map[string]ratelimit.Config{
    "free": {
        TokensPerMinute:  10000,
        TokensPerDay:     100000,
        RequestsPerMinute: 10,
        MaxConcurrent:    2,
    },
    "pro": {
        TokensPerMinute:  100000,
        TokensPerDay:     1000000,
        RequestsPerMinute: 60,
        MaxConcurrent:    10,
    },
    "enterprise": {
        TokensPerMinute:  1000000,
        TokensPerDay:     0,            // Unlimited
        RequestsPerMinute: 600,
        MaxConcurrent:    100,
    },
}

func aiRateLimiter(plan string) func(http.Handler) http.Handler {
    config := planLimits[plan]
    limiter := ratelimit.New(config, ratelimit.WithStorage(redis))

    return limiter.Middleware(func(r *http.Request) string {
        return auth.UserIDFromContext(r.Context())
    })
}
```

## Concurrent Request Limiting

```go
// Prevent too many simultaneous LLM calls per user
concurrentLimiter := ratelimit.NewConcurrentLimiter(
    ratelimit.WithMaxConcurrent(5), // Max 5 simultaneous AI requests per user
)

func handleStream(w http.ResponseWriter, r *http.Request) {
    userID := auth.UserIDFromContext(r.Context())

    // Acquire slot
    if !concurrentLimiter.Acquire(r.Context(), userID) {
        http.Error(w, "Too many concurrent requests", http.StatusTooManyRequests)
        return
    }
    defer concurrentLimiter.Release(userID)

    // Stream AI response...
}
```

## Rate Limit Headers

```go
// Standard rate limit response headers
X-RateLimit-Limit: 100000      // Total token budget per window
X-RateLimit-Remaining: 45231   // Remaining tokens
X-RateLimit-Reset: 1706745600  // Unix timestamp when limit resets
Retry-After: 3600              // Seconds until retry (on 429)

// AI-specific headers
X-Tokens-Used: 1523            // Tokens used in this request
X-Tokens-Remaining: 98477      // Remaining in current window
```

## Storage Backends

### In-Memory (Single Instance)

```go
limiter := ratelimit.New(config,
    ratelimit.WithInMemoryStorage(),
)
```

### Redis (Distributed)

```go
redisClient := redis.NewClient(os.Getenv("REDIS_URL"))
limiter := ratelimit.New(config,
    ratelimit.WithRedisStorage(redisClient,
        ratelimit.WithKeyPrefix("ai:ratelimit:"),
    ),
)
```

## Graceful Degradation

```go
// When rate limited, return cached or default response
func handleWithFallback(w http.ResponseWriter, r *http.Request) {
    userID := auth.UserIDFromContext(r.Context())

    decision, _ := limiter.Check(r.Context(), userID)
    if !decision.Allowed {
        // Try to serve from cache
        if cached, ok := responseCache.Get(r.URL.String()); ok {
            w.Header().Set("X-Served-From", "cache")
            json.NewEncoder(w).Encode(cached)
            return
        }

        // Inform client of limit
        w.Header().Set("Retry-After", strconv.Itoa(int(decision.RetryAfter.Seconds())))
        http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
        return
    }

    // Normal AI processing...
}
```

## Admin: View Usage

```go
app.Get("/admin/usage/:userID", adminAuth(func(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("userID")

    usage, err := limiter.GetUsage(r.Context(), userID)
    if err != nil {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(map[string]any{
        "user_id":              userID,
        "tokens_used_today":    usage.DailyTokens,
        "tokens_used_minute":   usage.MinuteTokens,
        "requests_used_minute": usage.MinuteRequests,
        "reset_at":             usage.ResetAt,
    })
}))
```

## Integration with Tenant Module

```go
// Use tenant quota as AI rate limit
func aiQuotaMiddleware(quotaMgr tenant.QuotaManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenantID := tenant.TenantIDFromContext(r.Context())

            // Check AI token quota
            decision, err := quotaMgr.CheckQuota(r.Context(), tenantID, "ai_tokens", 1000)
            if err != nil || !decision.Allowed {
                http.Error(w, "AI token quota exceeded", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Related Documentation

- [Token Counting](tokenizer.md) — Accurate token measurement
- [Circuit Breaker](circuit-breaker.md) — Failure protection
- [Tenant Module](../tenant/quota.md) — Tenant-level quotas
