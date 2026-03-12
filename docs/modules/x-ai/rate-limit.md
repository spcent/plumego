# AI Rate Limiting

> **Package**: `github.com/spcent/plumego/x/ai/ratelimit`

## Overview

The current rate-limit package exposes a generic token bucket limiter keyed by caller-defined strings.

It does not currently provide token-budget accounting middleware, distributed stores, or per-day AI quota objects.

## Core API

```go
limiter := ratelimit.NewTokenBucketLimiter(30, 0.5)
```

Meaning:

- capacity: `30`
- refill rate: `0.5` tokens per second

## Handler usage

```go
func handleAI(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    allowed, err := limiter.Allow(r.Context(), userID)
    if err != nil {
        http.Error(w, "rate limiter error", http.StatusInternalServerError)
        return
    }
    if !allowed {
        http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
        return
    }

    // Continue with provider call
}
```

## Remaining tokens

```go
remaining, err := limiter.Remaining(ctx, "user-123")
if err != nil {
    return err
}
fmt.Println(remaining)
```

## Resetting state

```go
_ = limiter.Reset(ctx, "user-123")
_ = limiter.ResetAll(ctx)
```

## Config-based construction

```go
limiter := ratelimit.NewTokenBucketLimiterWithConfig(ratelimit.TokenBucketConfig{
    Capacity:   60,
    RefillRate: 1.0,
    CleanupInt: 10 * time.Minute,
})
```

## Notes

- Keys are chosen by your application: user ID, tenant ID, API key, provider:model, etc.
- `resilience.NewResilientProvider(...)` can consume any `ratelimit.RateLimiter` implementation.
