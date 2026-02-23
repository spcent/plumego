# Circuit Breaker for LLM Calls

> **Package**: `github.com/spcent/plumego/ai/circuitbreaker` | **Feature**: Fault tolerance

## Overview

The `circuitbreaker` package prevents cascading failures when LLM providers experience outages or high latency. The circuit breaker monitors error rates and automatically stops forwarding requests when a provider is unhealthy.

## States

```
         failures > threshold
CLOSED ────────────────────► OPEN
  ▲                            │
  │        reset timeout       │
  └────────────────────────────┘
              HALF-OPEN
```

| State | Behavior |
|-------|----------|
| **Closed** | Requests pass through; errors are counted |
| **Open** | Requests fail immediately with `ErrCircuitOpen` |
| **Half-Open** | Single test request; success → Closed, failure → Open |

## Quick Start

```go
import "github.com/spcent/plumego/ai/circuitbreaker"

cb := circuitbreaker.New(
    circuitbreaker.WithFailureThreshold(5),          // Open after 5 failures
    circuitbreaker.WithSuccessThreshold(2),          // Close after 2 successes (half-open)
    circuitbreaker.WithTimeout(30 * time.Second),    // Stay open for 30s
    circuitbreaker.WithHalfOpenRequests(1),          // Test with 1 request
)

// Wrap provider calls
resp, err := cb.Call(ctx, func() (*provider.CompletionResponse, error) {
    return claude.Complete(ctx, req)
})

if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
    http.Error(w, "AI service temporarily unavailable", http.StatusServiceUnavailable)
    return
}
```

## Configuration

```go
cb := circuitbreaker.New(
    // Failure threshold: consecutive failures before opening
    circuitbreaker.WithFailureThreshold(5),

    // Success threshold: consecutive successes in half-open before closing
    circuitbreaker.WithSuccessThreshold(2),

    // Timeout: how long to stay open before trying half-open
    circuitbreaker.WithTimeout(30 * time.Second),

    // Count window: rolling window for failure counting
    circuitbreaker.WithCountWindow(10 * time.Second),

    // Error rate: open if error rate exceeds this % (alternative to count)
    circuitbreaker.WithErrorRate(0.5), // 50% error rate

    // Minimum requests before rate-based opening applies
    circuitbreaker.WithMinRequests(10),
)
```

## Middleware Integration

```go
// Apply to AI route group
ai := app.Group("/api/ai")
ai.Use(cb.Middleware())

// Handler doesn't need to know about circuit breaker
ai.Post("/chat", handleChat)
```

### Middleware with Custom Error Response

```go
func circuitBreakerMiddleware(cb *circuitbreaker.CircuitBreaker) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !cb.Allow() {
                w.Header().Set("Content-Type", "application/json")
                w.Header().Set("Retry-After", "30")
                w.WriteHeader(http.StatusServiceUnavailable)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "AI service temporarily unavailable",
                    "code":  "CIRCUIT_OPEN",
                })
                return
            }

            rr := &responseRecorder{ResponseWriter: w}
            next.ServeHTTP(rr, r)

            // Record result
            if rr.statusCode >= 500 {
                cb.RecordFailure()
            } else {
                cb.RecordSuccess()
            }
        })
    }
}
```

## Per-Provider Circuit Breakers

```go
type ProviderPool struct {
    providers map[string]provider.Provider
    breakers  map[string]*circuitbreaker.CircuitBreaker
}

func NewProviderPool() *ProviderPool {
    return &ProviderPool{
        providers: map[string]provider.Provider{
            "claude": claudeProvider,
            "openai": openaiProvider,
        },
        breakers: map[string]*circuitbreaker.CircuitBreaker{
            "claude": circuitbreaker.New(circuitbreaker.WithFailureThreshold(3)),
            "openai": circuitbreaker.New(circuitbreaker.WithFailureThreshold(3)),
        },
    }
}

func (p *ProviderPool) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
    // Try primary provider
    if p.breakers["claude"].Allow() {
        resp, err := p.providers["claude"].Complete(ctx, req)
        if err != nil {
            p.breakers["claude"].RecordFailure()
            log.Warnf("claude failed, trying openai: %v", err)
        } else {
            p.breakers["claude"].RecordSuccess()
            return resp, nil
        }
    }

    // Fallback to secondary
    if p.breakers["openai"].Allow() {
        resp, err := p.providers["openai"].Complete(ctx, req)
        if err != nil {
            p.breakers["openai"].RecordFailure()
            return nil, fmt.Errorf("all providers failed")
        }
        p.breakers["openai"].RecordSuccess()
        return resp, nil
    }

    return nil, circuitbreaker.ErrCircuitOpen
}
```

## State Change Callbacks

```go
cb := circuitbreaker.New(
    circuitbreaker.WithFailureThreshold(5),
    circuitbreaker.WithOnOpen(func() {
        log.Warn("Circuit opened - AI provider unhealthy")
        alerting.Send("AI_CIRCUIT_OPEN", "Claude API circuit breaker opened")
        metrics.CircuitBreakerState.Set(0) // 0 = open
    }),
    circuitbreaker.WithOnClose(func() {
        log.Info("Circuit closed - AI provider healthy")
        alerting.Resolve("AI_CIRCUIT_OPEN")
        metrics.CircuitBreakerState.Set(1) // 1 = closed
    }),
    circuitbreaker.WithOnHalfOpen(func() {
        log.Info("Circuit half-open - testing AI provider")
    }),
)
```

## Health Endpoint

```go
app.Get("/health/ai", func(w http.ResponseWriter, r *http.Request) {
    state := cb.State()
    status := http.StatusOK

    if state == circuitbreaker.StateOpen {
        status = http.StatusServiceUnavailable
    }

    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]any{
        "ai_provider": map[string]any{
            "state":          state.String(), // "closed", "open", "half_open"
            "failure_count":  cb.FailureCount(),
            "last_failure":   cb.LastFailure(),
            "next_attempt":   cb.NextAttempt(),
        },
    })
})
```

## Timeout Integration

Combine circuit breaker with request timeouts:

```go
resp, err := cb.Call(ctx, func() (*provider.CompletionResponse, error) {
    // Apply per-request timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    resp, err := claude.Complete(ctx, req)
    if errors.Is(err, context.DeadlineExceeded) {
        // Timeouts count as failures
        return nil, err
    }
    return resp, err
})
```

## Metrics

```go
// Expose circuit breaker metrics to Prometheus
breaker_state{provider="claude"}     1.0  // 1=closed, 0=open
breaker_failures_total{provider="claude"}  42
breaker_successes_total{provider="claude"} 1250
breaker_open_duration_seconds{provider="claude"} 120
```

## Related Documentation

- [Provider Interface](provider.md) — LLM providers
- [Resilience](resilience.md) — Retry policies
- [Rate Limiting](rate-limit.md) — Request throttling
