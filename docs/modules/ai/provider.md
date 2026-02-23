# LLM Provider Interface

> **Package**: `github.com/spcent/plumego/ai/provider` | **Supported**: Claude, OpenAI

## Overview

The `provider` package defines a unified interface for LLM providers, enabling seamless switching between AI models without changing application code.

## Provider Interface

```go
type Provider interface {
    // Name returns the provider identifier ("claude", "openai")
    Name() string

    // Complete sends a blocking completion request
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

    // CompleteStream returns a streaming reader for token-by-token responses
    CompleteStream(ctx context.Context, req *CompletionRequest) (*StreamReader, error)

    // ListModels returns available models for this provider
    ListModels(ctx context.Context) ([]Model, error)

    // GetModel returns details for a specific model
    GetModel(ctx context.Context, modelID string) (*Model, error)

    // CountTokens estimates token count for text
    CountTokens(text string) (int, error)
}
```

## Supported Providers

### Claude (Anthropic)

```go
import "github.com/spcent/plumego/ai/provider"

claude := provider.NewClaude(
    provider.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
    provider.WithModel("claude-opus-4-6"),      // Default model
    provider.WithMaxRetries(3),
    provider.WithTimeout(60 * time.Second),
)
```

**Available models**:
- `claude-opus-4-6` — Most capable, best for complex tasks
- `claude-sonnet-4-5-20250929` — Balanced performance/cost
- `claude-haiku-4-5-20251001` — Fastest, lowest cost

### OpenAI

```go
openai := provider.NewOpenAI(
    provider.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
    provider.WithModel("gpt-4"),
    provider.WithOrganization("org-xyz"),
    provider.WithBaseURL("https://api.openai.com/v1"), // Custom endpoint
)
```

**Available models**:
- `gpt-4` — Most capable GPT model
- `gpt-4-turbo` — Faster GPT-4 variant
- `gpt-3.5-turbo` — Economical option

## CompletionRequest

```go
req := &provider.CompletionRequest{
    Model: "claude-opus-4-6",

    // Chat messages
    Messages: []provider.Message{
        {Role: "user", Content: "Explain Go interfaces"},
    },

    // System prompt (optional)
    System: "You are a helpful Go programming assistant.",

    // Generation parameters
    MaxTokens:   2048,
    Temperature: 0.7,    // 0.0 = deterministic, 1.0 = creative
    TopP:        0.95,
    TopK:        50,

    // Stop conditions
    StopSequences: []string{"</answer>"},

    // Function calling
    Tools: []provider.Tool{...},
    ToolChoice: provider.ToolChoiceAuto,

    // Streaming
    Stream: true,

    // Tracking metadata
    Metadata: map[string]string{
        "user_id":    "user-123",
        "session_id": "sess-456",
    },
}
```

## CompletionResponse

```go
resp, err := claude.Complete(ctx, req)
if err != nil {
    return err
}

// Text content
for _, block := range resp.Content {
    switch block.Type {
    case "text":
        fmt.Println(block.Text)
    case "tool_use":
        fmt.Printf("Tool: %s, Input: %v\n", block.Name, block.Input)
    }
}

// Usage statistics
fmt.Printf("Input tokens: %d\n", resp.Usage.InputTokens)
fmt.Printf("Output tokens: %d\n", resp.Usage.OutputTokens)
fmt.Printf("Stop reason: %s\n", resp.StopReason) // "end_turn", "max_tokens", "tool_use"
```

## Streaming

```go
reader, err := claude.CompleteStream(ctx, &provider.CompletionRequest{
    Model:    "claude-opus-4-6",
    Messages: messages,
    Stream:   true,
})
if err != nil {
    return err
}
defer reader.Close()

for {
    delta, err := reader.Next()
    if err == io.EOF {
        break  // Stream complete
    }
    if err != nil {
        return err  // Stream error
    }

    switch delta.Type {
    case "text_delta":
        fmt.Print(delta.Text)  // Incremental text
    case "tool_use":
        // Tool call started
    case "message_stop":
        // Final message stats available
    }
}
```

## Message Format

```go
// Text message
msg := provider.Message{
    Role:    "user",   // "user", "assistant", "system"
    Content: "Hello!",
}

// Multi-part message with image
msg := provider.Message{
    Role: "user",
    Parts: []provider.ContentPart{
        {Type: "text", Text: "What's in this image?"},
        {Type: "image", Image: &provider.Image{
            MediaType: "image/jpeg",
            Data:      base64ImageData,
        }},
    },
}

// Tool result message
msg := provider.Message{
    Role: "user",
    Parts: []provider.ContentPart{
        {
            Type:      "tool_result",
            ToolUseID: toolCall.ID,
            Content:   "Tool executed successfully",
        },
    },
}
```

## Custom Provider

Implement the `Provider` interface for any LLM backend:

```go
type MistralProvider struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func (p *MistralProvider) Name() string { return "mistral" }

func (p *MistralProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
    // Build Mistral API request
    body, _ := json.Marshal(map[string]any{
        "model":       req.Model,
        "messages":    convertMessages(req.Messages),
        "max_tokens":  req.MaxTokens,
        "temperature": req.Temperature,
    })

    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        p.baseURL+"/chat/completions", bytes.NewReader(body))
    httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := p.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result mistralResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return convertResponse(result), nil
}

func (p *MistralProvider) CompleteStream(ctx context.Context, req *provider.CompletionRequest) (*provider.StreamReader, error) {
    // Implement SSE-based streaming...
    return nil, nil
}
```

## Multi-Provider Manager

```go
// Manager with automatic fallback
mgr := provider.NewManager()
mgr.Register("claude", claude, provider.WithPrimary(true))
mgr.Register("openai", openai, provider.WithFallback(true))

// Routes to primary; falls back on error
resp, err := mgr.Complete(ctx, req)
```

## Error Handling

```go
resp, err := claude.Complete(ctx, req)
if err != nil {
    switch {
    case provider.IsRateLimitError(err):
        // Retry after backoff
        retryAfter := provider.RetryAfter(err)
        time.Sleep(retryAfter)
    case provider.IsContextLengthError(err):
        // Trim conversation history
        req.Messages = trimMessages(req.Messages)
    case provider.IsAuthError(err):
        // Invalid API key
        log.Error("invalid API key")
    default:
        log.Errorf("provider error: %v", err)
    }
}
```

## Model Information

```go
// List available models
models, err := claude.ListModels(ctx)
for _, m := range models {
    fmt.Printf("Model: %s, Context: %d tokens, Cost: $%.4f/1k tokens\n",
        m.ID, m.ContextWindow, m.InputCostPer1K)
}

// Get specific model info
model, err := claude.GetModel(ctx, "claude-opus-4-6")
fmt.Printf("Max output: %d tokens\n", model.MaxOutputTokens)
```

## Configuration Reference

| Option | Description | Default |
|--------|-------------|---------|
| `WithAPIKey(key)` | API authentication key | Required |
| `WithModel(id)` | Default model ID | Provider default |
| `WithMaxRetries(n)` | Retry attempts on error | `3` |
| `WithTimeout(d)` | Request timeout | `60s` |
| `WithBaseURL(url)` | Custom API endpoint | Provider default |
| `WithOrganization(id)` | Organization ID (OpenAI) | None |

## Related Documentation

- [Session Management](session.md) — Conversation history
- [SSE Streaming](sse.md) — Real-time streaming
- [Token Counting](tokenizer.md) — Usage tracking
- [Circuit Breaker](circuit-breaker.md) — Fault tolerance
