# Resilience Patterns

> **Package**: `github.com/spcent/plumego/ai/resilience` | **Feature**: Error handling and retries

## Overview

The `resilience` package provides retry policies, error classification, and fallback strategies for LLM API calls. LLM APIs are inherently unreliable — they experience rate limits, transient failures, and occasional timeouts. Resilience patterns handle these gracefully.

## Quick Start

```go
import "github.com/spcent/plumego/ai/resilience"

// Wrap LLM calls with retry
retrier := resilience.NewRetrier(
    resilience.WithMaxAttempts(3),
    resilience.WithExponentialBackoff(time.Second, 30*time.Second),
    resilience.WithRetryOn(resilience.IsRetryable),
)

resp, err := retrier.Do(ctx, func() (*provider.CompletionResponse, error) {
    return claude.Complete(ctx, req)
})
```

## Retry Policies

### Exponential Backoff

```go
// Delays: 1s, 2s, 4s, 8s (max 30s)
retrier := resilience.NewRetrier(
    resilience.WithMaxAttempts(4),
    resilience.WithExponentialBackoff(
        1*time.Second,   // Initial delay
        30*time.Second,  // Maximum delay
    ),
)
```

### Exponential with Jitter

Jitter prevents thundering herd when many requests retry simultaneously:

```go
retrier := resilience.NewRetrier(
    resilience.WithMaxAttempts(4),
    resilience.WithExponentialBackoff(time.Second, 30*time.Second),
    resilience.WithJitter(0.25), // ±25% randomization
)
// Delays: ~1s, ~2s, ~4s, ~8s (with random variation)
```

### Linear Backoff

```go
// Delays: 5s, 5s, 5s
retrier := resilience.NewRetrier(
    resilience.WithMaxAttempts(3),
    resilience.WithLinearBackoff(5 * time.Second),
)
```

### Custom Retry Condition

```go
retrier := resilience.NewRetrier(
    resilience.WithMaxAttempts(5),
    resilience.WithExponentialBackoff(time.Second, 60*time.Second),
    resilience.WithRetryOn(func(err error) bool {
        // Only retry transient errors
        if provider.IsRateLimitError(err) {
            return true
        }
        if provider.IsServerError(err) {
            return true
        }
        if errors.Is(err, context.DeadlineExceeded) {
            return false // Don't retry timeouts
        }
        return false
    }),
)
```

## Error Classification

```go
// Check error type
switch {
case provider.IsRateLimitError(err):
    // 429 - back off and retry
    retryAfter := provider.RetryAfter(err)
    time.Sleep(retryAfter)

case provider.IsContextLengthError(err):
    // 400 - context too long, trim messages
    req.Messages = trimOldestMessages(req.Messages)

case provider.IsAuthError(err):
    // 401 - invalid API key, do not retry
    return ErrInvalidAPIKey

case provider.IsServerError(err):
    // 5xx - transient failure, retry with backoff
    return retrier.Do(ctx, callLLM)

case errors.Is(err, context.DeadlineExceeded):
    // Timeout - may retry with shorter request
    return ErrRequestTimeout

case errors.Is(err, context.Canceled):
    // Client disconnected - abort
    return nil
}
```

## Retry-After Handling

When the API returns a `Retry-After` header, respect it:

```go
retrier := resilience.NewRetrier(
    resilience.WithMaxAttempts(3),
    resilience.WithRespectRetryAfter(true), // Use server-provided delay
    resilience.WithMaxRetryAfter(5*time.Minute), // Cap at 5 minutes
)
```

## Fallback Strategies

### Provider Fallback

```go
resp, err := claude.Complete(ctx, req)
if err != nil && provider.IsServerError(err) {
    log.Warnf("Claude unavailable, falling back to OpenAI: %v", err)
    resp, err = openai.Complete(ctx, req) // Fallback provider
}
```

### Model Downgrade

```go
// If large model fails, try smaller model
resp, err := claude.Complete(ctx, &provider.CompletionRequest{
    Model:    "claude-opus-4-6",
    Messages: messages,
    MaxTokens: 4096,
})

if err != nil && provider.IsContextLengthError(err) {
    // Context too large for Opus, try summarized version
    summarized := summarizeMessages(messages)
    resp, err = claude.Complete(ctx, &provider.CompletionRequest{
        Model:    "claude-haiku-4-5-20251001", // Faster/cheaper
        Messages: summarized,
        MaxTokens: 2048,
    })
}
```

### Cached Fallback

```go
resp, err := claude.Complete(ctx, req)
if err != nil {
    // Serve stale cached response during outage
    if cached, ok := cache.GetStale(cacheKey); ok {
        w.Header().Set("X-Served-From", "stale-cache")
        return cached, nil
    }
    return nil, err
}
```

## Bulkhead Pattern

Isolate AI calls from affecting the rest of the application:

```go
// Limit concurrent AI requests via semaphore
semaphore := make(chan struct{}, 20) // Max 20 concurrent

func callAI(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
    select {
    case semaphore <- struct{}{}:
        defer func() { <-semaphore }()
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        return nil, errors.New("AI service at capacity")
    }

    return claude.Complete(ctx, req)
}
```

## Timeout Configuration

```go
// Global timeout for all AI requests
app.Use("/api/ai", middleware.Timeout(60*time.Second))

// Per-model timeouts
timeouts := map[string]time.Duration{
    "claude-opus-4-6":  90 * time.Second, // Larger model, more time
    "claude-haiku-4-5": 30 * time.Second, // Fast model, short timeout
    "gpt-3.5-turbo":    20 * time.Second,
}

func handleChat(w http.ResponseWriter, r *http.Request) {
    model := req.Model
    timeout := timeouts[model]

    ctx, cancel := context.WithTimeout(r.Context(), timeout)
    defer cancel()

    resp, err := claude.Complete(ctx, &provider.CompletionRequest{
        Model: model,
        // ...
    })
    if errors.Is(err, context.DeadlineExceeded) {
        http.Error(w, "Request timed out. Try a shorter prompt.", http.StatusGatewayTimeout)
        return
    }
}
```

## Observability

```go
// Track retry attempts in metrics
retrier := resilience.NewRetrier(
    resilience.WithMaxAttempts(3),
    resilience.WithExponentialBackoff(time.Second, 30*time.Second),
    resilience.WithOnRetry(func(attempt int, err error) {
        aiRetries.WithLabelValues(
            strconv.Itoa(attempt),
            classifyError(err),
        ).Inc()
        log.Warnf("AI retry attempt %d: %v", attempt, err)
    }),
)
```

## Combined Pattern

Production-ready resilience combining circuit breaker, retry, and timeout:

```go
type ResilientAI struct {
    provider provider.Provider
    breaker  *circuitbreaker.CircuitBreaker
    retrier  *resilience.Retrier
}

func (r *ResilientAI) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
    if !r.breaker.Allow() {
        return nil, circuitbreaker.ErrCircuitOpen
    }

    resp, err := r.retrier.Do(ctx, func() (*provider.CompletionResponse, error) {
        reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
        return r.provider.Complete(reqCtx, req)
    })

    if err != nil {
        r.breaker.RecordFailure()
        return nil, err
    }

    r.breaker.RecordSuccess()
    return resp, nil
}
```

## Related Documentation

- [Circuit Breaker](circuit-breaker.md) — State-based fault tolerance
- [Rate Limiting](rate-limit.md) — Quota management
- [Provider Interface](provider.md) — Error types
