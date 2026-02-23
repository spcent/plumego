# AI Agent Gateway Module

> **Package Path**: `github.com/spcent/plumego/ai` | **Stability**: Medium | **Priority**: P2

## Overview

The `ai/` package provides an AI agent gateway for Plumego applications. It offers a unified interface for multiple LLM providers, real-time streaming via Server-Sent Events, conversation session management, token counting, function calling, semantic caching, and fault-tolerance patterns.

**Key Features**:
- **Unified Provider Interface**: Single API for Claude, OpenAI, and custom providers
- **SSE Streaming**: Real-time token-by-token responses via Server-Sent Events
- **Session Management**: Conversation history with context window control
- **Token Counting**: Usage tracking and quota enforcement
- **Function Calling**: Tool framework for agent actions
- **Semantic Cache**: Embedding-based response deduplication
- **Circuit Breaker**: Fault tolerance for LLM calls
- **Orchestration**: Multi-step AI workflow coordination
- **Rate Limiting**: Per-tenant/per-user AI quota management

## Subpackages

| Package | Description |
|---------|-------------|
| `ai/provider` | Unified LLM provider interface (Claude, OpenAI) |
| `ai/session` | Conversation session management |
| `ai/sse` | Server-Sent Events streaming |
| `ai/streaming` | Streaming response processing |
| `ai/tokenizer` | Token counting and quota management |
| `ai/tool` | Function calling framework |
| `ai/semanticcache` | Embedding-based response caching |
| `ai/llmcache` | Exact-match LLM response caching |
| `ai/circuitbreaker` | Circuit breaker for LLM calls |
| `ai/orchestration` | AI workflow orchestration |
| `ai/prompt` | Prompt management and templating |
| `ai/ratelimit` | AI endpoint rate limiting |
| `ai/resilience` | Error handling and retries |
| `ai/filter` | Request/response filtering |
| `ai/instrumentation` | AI metrics and observability |

## Quick Start

### Basic AI Endpoint

```go
import (
    "github.com/spcent/plumego/ai/provider"
    "github.com/spcent/plumego/ai/session"
    "github.com/spcent/plumego/ai/sse"
)

// Create provider
claude := provider.NewClaude(
    provider.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
    provider.WithModel("claude-opus-4-6"),
)

// Session manager
sessionMgr := session.NewManager(
    session.NewInMemoryStorage(),
    session.WithMaxTokens(100000),
    session.WithDefaultTTL(24*time.Hour),
)

// Streaming chat endpoint
app.Post("/api/chat", func(w http.ResponseWriter, r *http.Request) {
    var req struct {
        SessionID string `json:"session_id"`
        Message   string `json:"message"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Create SSE stream
    stream, err := sse.NewStream(r.Context(), w)
    if err != nil {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }
    defer stream.Close()

    // Get session
    sess, _ := sessionMgr.GetOrCreate(r.Context(), req.SessionID)
    sessionMgr.AddMessage(r.Context(), sess.ID, provider.Message{
        Role:    "user",
        Content: req.Message,
    })

    // Stream completion
    reader, err := claude.CompleteStream(r.Context(), &provider.CompletionRequest{
        Model:    "claude-opus-4-6",
        Messages: sess.Messages,
    })
    if err != nil {
        stream.SendError("completion failed")
        return
    }
    defer reader.Close()

    // Pipe tokens to client
    for {
        delta, err := reader.Next()
        if err == io.EOF {
            break
        }
        stream.Send(sse.Event{Data: delta.Text})
    }
})
```

### Non-Streaming Chat

```go
resp, err := claude.Complete(r.Context(), &provider.CompletionRequest{
    Model: "claude-opus-4-6",
    Messages: []provider.Message{
        {Role: "user", Content: "What is Go?"},
    },
    MaxTokens: 1024,
})
if err != nil {
    http.Error(w, "AI error", http.StatusInternalServerError)
    return
}

json.NewEncoder(w).Encode(map[string]string{
    "response": resp.Content[0].Text,
})
```

## Integration with Core

### Application Setup

```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithAIProvider(claude),
    core.WithSessionManager(sessionMgr),
)
```

### AI Middleware Stack

```go
// Typical AI endpoint middleware
aiMiddleware := middleware.NewChain().
    Use(middleware.RequestID).
    Use(rateLimiter.Middleware()).          // Per-user rate limiting
    Use(tokenBudget.Middleware(10000)).     // Token budget enforcement
    Use(circuitBreaker.Middleware()).       // Fault tolerance
    Apply(aiHandler)
```

## Architecture

```
Client Request
     │
     ▼
Rate Limiter (per-user/tenant)
     │
     ▼
Circuit Breaker (fault tolerance)
     │
     ▼
Session Manager (conversation history)
     │
     ▼
Prompt Builder (system prompt + history)
     │
     ▼
LLM Provider (Claude/OpenAI)
     │
     ▼
Token Counter (usage tracking)
     │
     ▼
SSE Stream → Client (real-time tokens)
```

## Configuration

### Provider Options

```go
// Claude (Anthropic)
claude := provider.NewClaude(
    provider.WithAPIKey("sk-ant-..."),
    provider.WithModel("claude-opus-4-6"),
    provider.WithMaxRetries(3),
    provider.WithTimeout(60*time.Second),
)

// OpenAI
openai := provider.NewOpenAI(
    provider.WithAPIKey("sk-..."),
    provider.WithModel("gpt-4"),
    provider.WithOrganization("org-..."),
)
```

### Multi-Provider Routing

```go
// Route based on request requirements
router := provider.NewRouter()
router.Add("claude", claude, provider.WithPriority(1))
router.Add("openai", openai, provider.WithPriority(2))
router.SetFallback("openai") // Fallback if primary fails
```

## Module Documentation

- **[Provider Interface](provider.md)** — LLM provider abstraction
- **[Session Management](session.md)** — Conversation history
- **[SSE Streaming](sse.md)** — Server-Sent Events
- **[Streaming Processing](streaming.md)** — Stream handling patterns
- **[Token Counting](tokenizer.md)** — Usage tracking
- **[Function Calling](tool.md)** — Agent tools
- **[Semantic Cache](semantic-cache.md)** — Embedding-based caching
- **[Circuit Breaker](circuit-breaker.md)** — Fault tolerance
- **[Orchestration](orchestration.md)** — AI workflows
- **[Rate Limiting](rate-limit.md)** — AI quota management
- **[Resilience](resilience.md)** — Error handling and retries

## Related Documentation

- [Middleware: Observability](../middleware/observability.md) — Request tracing
- [Metrics Module](../metrics/) — Prometheus/OTel metrics
- [Tenant Module](../tenant/) — Multi-tenant AI quotas

## Reference Implementation

See `examples/ai-agent-gateway/` for a complete working AI gateway with:
- Multi-provider routing
- Real-time SSE streaming
- Session-based conversation history
- Token quota management
- Semantic response caching
