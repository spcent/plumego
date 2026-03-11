# AI Agent Gateway Examples

> **Package**: `github.com/spcent/plumego/ai` | **Examples aligned with current API surface**

The previous examples in this document referenced outdated constructors and app options. The examples below use the current provider, session, resilience, and SSE APIs.

## Example 1: Streaming chat endpoint

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"

    "github.com/spcent/plumego/ai/provider"
    "github.com/spcent/plumego/ai/session"
    "github.com/spcent/plumego/ai/sse"
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
            SessionID string `json:"session_id"`
            Message   string `json:"message"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }

        sessID := req.SessionID
        if sessID == "" {
            created, err := sessions.Create(r.Context(), session.CreateOptions{Model: "claude-3-sonnet-20240229"})
            if err != nil {
                http.Error(w, "session error", http.StatusInternalServerError)
                return
            }
            sessID = created.ID
        }

        if err := sessions.AppendMessage(r.Context(), sessID, provider.Message{Role: "user", Content: req.Message}); err != nil {
            http.Error(w, "session error", http.StatusInternalServerError)
            return
        }

        active, err := sessions.GetActiveContext(r.Context(), sessID, 100000)
        if err != nil {
            http.Error(w, "session error", http.StatusInternalServerError)
            return
        }

        reader, err := model.CompleteStream(r.Context(), &provider.CompletionRequest{
            Model:     "claude-3-sonnet-20240229",
            Messages:  active,
            MaxTokens: 2048,
            Stream:    true,
        })
        if err != nil {
            http.Error(w, "provider error", http.StatusBadGateway)
            return
        }
        defer reader.Close()

        stream, err := sse.NewStream(r.Context(), w)
        if err != nil {
            http.Error(w, "streaming not supported", http.StatusInternalServerError)
            return
        }
        defer stream.Close()

        var full string
        for {
            delta, err := reader.Next()
            if err == io.EOF {
                break
            }
            if err != nil {
                http.Error(w, "stream error", http.StatusBadGateway)
                return
            }
            if delta.Text == "" {
                continue
            }

            full += delta.Text
            if err := stream.Send(&sse.Event{Event: "token", Data: fmt.Sprintf(`{"text":%q}`, delta.Text)}); err != nil {
                return
            }
        }

        _ = sessions.AppendMessage(r.Context(), sessID, provider.Message{Role: "assistant", Content: full})
        _ = stream.Send(&sse.Event{Event: "done", Data: fmt.Sprintf(`{"session_id":%q}`, sessID)})
    })
}
```

## Example 2: Resilient provider wrapper

```go
base := provider.NewClaudeProvider(os.Getenv("ANTHROPIC_API_KEY"))
limiter := ratelimit.NewTokenBucketLimiter(30, 0.5)
breaker := circuitbreaker.NewCircuitBreaker("claude", 5, 30*time.Second)

model := resilience.NewResilientProvider(resilience.Config{
    Provider:       base,
    RateLimiter:    limiter,
    CircuitBreaker: breaker,
})
```

## Example 3: Session manager with explicit config

```go
sessions := session.NewManager(
    session.NewMemoryStorage(),
    session.WithConfig(session.Config{
        DefaultTTL:  24 * time.Hour,
        MaxMessages: 200,
        MaxTokens:   100000,
        AutoTrim:    true,
    }),
)
```
