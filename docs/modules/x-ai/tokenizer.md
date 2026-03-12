# Token Counting

> **Package**: `github.com/spcent/plumego/x/ai/tokenizer` | **Feature**: Usage tracking

## Overview

The `tokenizer` package provides token counting for LLM requests and responses. Accurate token counting enables cost estimation, quota enforcement, and context window management.

## Quick Start

```go
import "github.com/spcent/plumego/x/ai/tokenizer"

t := tokenizer.NewSimpleTokenizer("claude-opus-4-6")

// Count tokens in text
count, err := t.CountTokens("Hello, how are you today?")
fmt.Printf("Tokens: %d\n", count) // ~7 tokens

// Count tokens in messages
messages := []provider.Message{
    {Role: "user", Content: "What is Go?"},
    {Role: "assistant", Content: "Go is a systems programming language."},
}
total, err := t.CountMessages(messages)
fmt.Printf("Total tokens: %d\n", total)
```

## Token Usage Tracking

```go
type TokenUsage struct {
    Input  int // Prompt tokens
    Output int // Completion tokens
    Total  int // Input + Output
}
```

### Track Per-Request Usage

```go
resp, err := claude.Complete(ctx, req)
if err != nil {
    return err
}

// Usage from response
usage := tokenizer.TokenUsage{
    Input:  resp.Usage.InputTokens,
    Output: resp.Usage.OutputTokens,
    Total:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
}

// Record for billing/quota
usageStore.Record(userID, usage)
```

### Estimate Before Request

```go
t := tokenizer.NewSimpleTokenizer("claude-opus-4-6")

// Estimate cost before sending
inputTokens, _ := t.CountMessages(messages)
estimatedOutput := 1024 // Expected max output

totalEstimate := inputTokens + estimatedOutput
if totalEstimate > userQuota.RemainingTokens {
    return errors.New("insufficient token quota")
}

// Cost estimate
inputCost := float64(inputTokens) / 1000000 * 15.00   // $15 per 1M input tokens
outputCost := float64(estimatedOutput) / 1000000 * 75.00 // $75 per 1M output tokens
fmt.Printf("Estimated cost: $%.4f\n", inputCost+outputCost)
```

## Context Window Management

Models have token limits for their context window. The tokenizer helps stay within bounds.

```go
const maxContextTokens = 200000 // claude-opus-4-6 limit

func fitInContext(messages []provider.Message, maxTokens int) []provider.Message {
    t := tokenizer.NewSimpleTokenizer("claude-opus-4-6")

    total := 0
    result := make([]provider.Message, 0)

    // Always include system message (first message)
    if len(messages) > 0 && messages[0].Role == "system" {
        tokens, _ := t.CountTokens(messages[0].Content)
        total += tokens
        result = append(result, messages[0])
        messages = messages[1:]
    }

    // Add most recent messages until limit
    for i := len(messages) - 1; i >= 0; i-- {
        tokens, _ := t.CountTokens(messages[i].Content)
        if total+tokens > maxTokens {
            break
        }
        total += tokens
        result = append([]provider.Message{messages[i]}, result...)
    }

    return result
}
```

## Token Budget Middleware

```go
func tokenBudgetMiddleware(budget int, tokenizer tokenizer.Tokenizer) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Read and decode request body
            body, _ := io.ReadAll(r.Body)
            r.Body = io.NopCloser(bytes.NewReader(body))

            var req ChatRequest
            json.Unmarshal(body, &req)

            // Check token count
            tokens, err := tokenizer.CountTokens(req.Message)
            if err != nil || tokens > budget {
                http.Error(w, fmt.Sprintf("Message too long (%d tokens, max %d)", tokens, budget),
                    http.StatusBadRequest)
                return
            }

            // Add token count to context
            ctx := context.WithValue(r.Context(), "input_tokens", tokens)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Model Token Limits

| Model | Context Window | Max Output |
|-------|---------------|------------|
| `claude-opus-4-6` | 200,000 | 4,096 |
| `claude-sonnet-4-5` | 200,000 | 4,096 |
| `claude-haiku-4-5` | 200,000 | 4,096 |
| `gpt-4` | 128,000 | 4,096 |
| `gpt-4-turbo` | 128,000 | 4,096 |
| `gpt-3.5-turbo` | 16,385 | 4,096 |

## Cost Calculation

```go
type ModelPricing struct {
    InputPer1M  float64 // Cost per 1M input tokens
    OutputPer1M float64 // Cost per 1M output tokens
}

var pricing = map[string]ModelPricing{
    "claude-opus-4-6":    {InputPer1M: 15.00, OutputPer1M: 75.00},
    "claude-sonnet-4-5":  {InputPer1M: 3.00, OutputPer1M: 15.00},
    "claude-haiku-4-5":   {InputPer1M: 0.25, OutputPer1M: 1.25},
    "gpt-4":              {InputPer1M: 30.00, OutputPer1M: 60.00},
    "gpt-3.5-turbo":      {InputPer1M: 0.50, OutputPer1M: 1.50},
}

func CalculateCost(model string, usage tokenizer.TokenUsage) float64 {
    p, ok := pricing[model]
    if !ok {
        return 0
    }
    input := float64(usage.Input) / 1_000_000 * p.InputPer1M
    output := float64(usage.Output) / 1_000_000 * p.OutputPer1M
    return input + output
}
```

## Cumulative Usage Tracking

```go
type UsageTracker struct {
    mu     sync.Mutex
    totals map[string]*tokenizer.TokenUsage // key: userID
}

func (t *UsageTracker) Record(userID string, usage tokenizer.TokenUsage) {
    t.mu.Lock()
    defer t.mu.Unlock()

    if t.totals[userID] == nil {
        t.totals[userID] = &tokenizer.TokenUsage{}
    }
    t.totals[userID].Input += usage.Input
    t.totals[userID].Output += usage.Output
    t.totals[userID].Total += usage.Total
}

func (t *UsageTracker) GetUsage(userID string) tokenizer.TokenUsage {
    t.mu.RLock()
    defer t.mu.RUnlock()
    if u := t.totals[userID]; u != nil {
        return *u
    }
    return tokenizer.TokenUsage{}
}
```

## Custom Tokenizer

```go
type CustomTokenizer struct {
    model string
}

func (t *CustomTokenizer) CountTokens(text string) (int, error) {
    // Rough estimate: 1 token ≈ 4 characters
    return len(text) / 4, nil
}

func (t *CustomTokenizer) CountMessages(msgs []provider.Message) (int, error) {
    total := 0
    for _, m := range msgs {
        n, err := t.CountTokens(m.Content)
        if err != nil {
            return 0, err
        }
        total += n + 4 // Per-message overhead
    }
    return total, nil
}
```

## Related Documentation

- [Provider Interface](provider.md) — Token usage in responses
- [Session Management](session.md) — Session token limits
- [Rate Limiting](rate-limit.md) — Token-based quotas
