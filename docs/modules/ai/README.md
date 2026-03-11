# AI Agent Gateway Module

> **Package Path**: `github.com/spcent/plumego/ai` | **Stability**: Medium | **Priority**: P2

## Overview

The `ai/` tree provides provider adapters, streaming helpers, session management, token counting, tools, resilience wrappers, and related building blocks.

Current stable entry points documented here:

- `ai/provider`: unified provider interface with Claude/OpenAI adapters
- `ai/session`: conversation/session storage and token-window trimming
- `ai/sse`: SSE response streaming helpers
- `ai/ratelimit`: token bucket rate limiter primitives
- `ai/resilience`: provider wrapper with rate limiting and circuit breaking

There is no current `core.WithAIProvider(...)` or `core.WithSessionManager(...)` integration option. Wire AI dependencies explicitly in your handlers/services.

## Minimal chat wiring

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"

    "github.com/spcent/plumego/ai/provider"
    "github.com/spcent/plumego/ai/session"
    "github.com/spcent/plumego/core"
)

func main() {
    model := provider.NewClaudeProvider(os.Getenv("ANTHROPIC_API_KEY"))
    sessions := session.NewManager(session.NewMemoryStorage())

    app := core.New(core.WithAddr(":8080"))
    if err := app.Router().AddRoute(http.MethodPost, "/chat", handleChat(model, sessions)); err != nil {
        log.Fatal(err)
    }

    if err := app.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}

func handleChat(model provider.Provider, sessions *session.Manager) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            Message string `json:"message"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }

        sess, err := sessions.Create(r.Context(), session.CreateOptions{Model: "claude-3-sonnet-20240229"})
        if err != nil {
            http.Error(w, "session error", http.StatusInternalServerError)
            return
        }

        _ = sessions.AppendMessage(r.Context(), sess.ID, provider.NewTextMessage(provider.RoleUser, req.Message))
        messages, _ := sessions.GetActiveContext(r.Context(), sess.ID, 100000)

        resp, err := model.Complete(r.Context(), &provider.CompletionRequest{
            Model:     "claude-3-sonnet-20240229",
            Messages:  messages,
            MaxTokens: 512,
        })
        if err != nil {
            http.Error(w, "provider error", http.StatusBadGateway)
            return
        }

        _ = sessions.AppendMessage(r.Context(), sess.ID, provider.NewTextMessage(provider.RoleAssistant, resp.GetText()))
        _ = json.NewEncoder(w).Encode(map[string]string{"response": resp.GetText()})
    })
}
```

## Related pages

- [Provider Interface](provider.md)
- [Session Management](session.md)
- [SSE Streaming](sse.md)
- [Rate Limiting](rate-limit.md)
- [Examples](examples.md)
