# AI Agent Gateway Examples

> **Package**: `github.com/spcent/plumego/ai` | **Complete working examples**

## Example 1: Production AI Chat Service

Complete streaming chat service with sessions, rate limiting, and circuit breaking.

```go
package main

import (
    "io"
    "log"
    "os"
    "time"
    "net/http"
    "encoding/json"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/ai/provider"
    "github.com/spcent/plumego/ai/session"
    "github.com/spcent/plumego/ai/sse"
    "github.com/spcent/plumego/ai/ratelimit"
    "github.com/spcent/plumego/ai/circuitbreaker"
    "github.com/spcent/plumego/ai/resilience"
)

func main() {
    // Initialize provider with resilience
    claude := provider.NewClaude(
        provider.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
    )

    cb := circuitbreaker.New(
        circuitbreaker.WithFailureThreshold(5),
        circuitbreaker.WithTimeout(30*time.Second),
    )

    retrier := resilience.NewRetrier(
        resilience.WithMaxAttempts(3),
        resilience.WithExponentialBackoff(time.Second, 30*time.Second),
        resilience.WithRetryOn(provider.IsTransientError),
    )

    sessionMgr := session.NewManager(
        session.NewInMemoryStorage(),
        session.WithMaxTokens(100000),
        session.WithDefaultTTL(24*time.Hour),
        session.WithAutoTrim(true),
    )

    limiter := ratelimit.NewTokenLimiter(
        ratelimit.WithTokensPerMinute(50000),
        ratelimit.WithRequestsPerMinute(30),
    )

    app := core.New(core.WithAddr(":8080"))

    api := app.Router().Group("/api")
    api.Use(limiter.Middleware(extractUserID))

    api.Post("/chat", handleChat(claude, sessionMgr, cb, retrier))
    api.Get("/sessions", handleListSessions(sessionMgr))
    api.Delete("/sessions/:id", handleDeleteSession(sessionMgr))

    if err := app.Boot(); err != nil {
        log.Fatal(err)
    }
}

func handleChat(claude provider.Provider, sessionMgr *session.Manager,
    cb *circuitbreaker.CircuitBreaker, retrier *resilience.Retrier) http.HandlerFunc {

    return func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            SessionID string `json:"session_id"`
            Message   string `json:"message"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        // Check circuit breaker
        if !cb.Allow() {
            w.Header().Set("Retry-After", "30")
            http.Error(w, "AI service temporarily unavailable", http.StatusServiceUnavailable)
            return
        }

        // Get session
        sess, err := sessionMgr.GetOrCreate(r.Context(), req.SessionID)
        if err != nil {
            http.Error(w, "Session error", http.StatusInternalServerError)
            return
        }

        // Add user message
        sessionMgr.AddMessage(r.Context(), sess.ID, provider.Message{
            Role:    "user",
            Content: req.Message,
        })

        // Create SSE stream
        stream, err := sse.NewStream(r.Context(), w,
            sse.WithKeepAlive(20*time.Second),
        )
        if err != nil {
            http.Error(w, "Streaming not supported", http.StatusInternalServerError)
            return
        }
        defer stream.Close()

        // Call LLM with retry
        var reader *provider.StreamReader
        _, err = retrier.Do(r.Context(), func() (struct{}, error) {
            var e error
            reader, e = claude.CompleteStream(r.Context(), &provider.CompletionRequest{
                Model:     "claude-opus-4-6",
                System:    "You are a helpful assistant.",
                Messages:  sess.Messages,
                MaxTokens: 2048,
                Stream:    true,
            })
            return struct{}{}, e
        })

        if err != nil {
            cb.RecordFailure()
            stream.SendError("AI service error")
            return
        }
        defer reader.Close()

        cb.RecordSuccess()

        // Stream tokens
        var fullResponse strings.Builder
        for {
            delta, err := reader.Next()
            if err == io.EOF {
                break
            }
            if err != nil {
                stream.SendError("Stream interrupted")
                return
            }
            if delta.Text != "" {
                fullResponse.WriteString(delta.Text)
                stream.Send(sse.Event{
                    Event: "token",
                    Data:  fmt.Sprintf(`{"text":%q}`, delta.Text),
                })
            }
        }

        // Save response to session
        sessionMgr.AddMessage(r.Context(), sess.ID, provider.Message{
            Role:    "assistant",
            Content: fullResponse.String(),
        })

        stream.Send(sse.Event{
            Event: "done",
            Data:  fmt.Sprintf(`{"session_id":%q}`, sess.ID),
        })
    }
}
```

## Example 2: Research Agent with Tools

```go
func handleResearch(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Query string `json:"query"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Define tools
    registry := tool.NewRegistry()
    registry.Register(tool.New("search_web", "Search the web",
        tool.Schema{
            Type: "object",
            Properties: map[string]tool.Property{
                "query": {Type: "string", Description: "Search query"},
            },
            Required: []string{"query"},
        },
        func(ctx context.Context, input map[string]any) (string, error) {
            results, _ := webSearch.Search(ctx, input["query"].(string))
            out, _ := json.Marshal(results)
            return string(out), nil
        },
    ))

    registry.Register(tool.New("get_page", "Fetch webpage content",
        tool.Schema{
            Type: "object",
            Properties: map[string]tool.Property{
                "url": {Type: "string", Description: "URL to fetch"},
            },
            Required: []string{"url"},
        },
        func(ctx context.Context, input map[string]any) (string, error) {
            content, _ := httpFetcher.Get(ctx, input["url"].(string))
            return content, nil
        },
    ))

    // Initial message
    messages := []provider.Message{
        {Role: "user", Content: req.Query},
    }

    // Run agentic loop
    result, err := runAgentLoop(r.Context(), messages, registry, 10)
    if err != nil {
        http.Error(w, "Research failed", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"result": result})
}

func runAgentLoop(ctx context.Context, messages []provider.Message,
    tools *tool.Registry, maxIter int) (string, error) {

    for i := 0; i < maxIter; i++ {
        resp, err := claude.Complete(ctx, &provider.CompletionRequest{
            Model:      "claude-opus-4-6",
            Messages:   messages,
            Tools:      tools.ProviderTools(),
            ToolChoice: provider.ToolChoiceAuto,
            MaxTokens:  4096,
        })
        if err != nil {
            return "", err
        }

        // Add assistant message
        messages = append(messages, provider.Message{
            Role: "assistant", Content: resp.RawContent,
        })

        if resp.StopReason == "end_turn" {
            for _, block := range resp.Content {
                if block.Type == "text" {
                    return block.Text, nil
                }
            }
            return "", nil
        }

        // Execute tools
        var toolResults []provider.ContentPart
        for _, block := range resp.Content {
            if block.Type != "tool_use" {
                continue
            }
            result, err := tools.Execute(ctx, block.Name, block.Input)
            if err != nil {
                result = "Error: " + err.Error()
            }
            toolResults = append(toolResults, provider.ContentPart{
                Type:      "tool_result",
                ToolUseID: block.ID,
                Content:   result,
            })
        }

        messages = append(messages, provider.Message{
            Role: "user", Parts: toolResults,
        })
    }

    return "", errors.New("max iterations reached")
}
```

## Example 3: Document Q&A with Semantic Cache

```go
func setupDocumentQA() http.HandlerFunc {
    // Embedding + caching setup
    embedder := embedding.NewOpenAI(
        embedding.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
        embedding.WithModel("text-embedding-3-small"),
    )

    cache := semanticcache.New(
        semanticcache.WithEmbeddingProvider(embedder),
        semanticcache.WithVectorStore(semanticcache.NewInMemoryVectorStore()),
        semanticcache.WithSimilarityThreshold(0.93),
        semanticcache.WithTTL(24*time.Hour),
    )

    cachedClaude := semanticcache.Wrap(claude, cache)

    return func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            Question string `json:"question"`
            DocID    string `json:"doc_id"`
        }
        json.NewDecoder(r.Body).Decode(&req)

        // Fetch document context
        doc, _ := docStore.Get(req.DocID)

        resp, err := cachedClaude.Complete(r.Context(), &provider.CompletionRequest{
            Model: "claude-opus-4-6",
            System: fmt.Sprintf("Answer questions based on this document:\n\n%s", doc.Content),
            Messages: []provider.Message{
                {Role: "user", Content: req.Question},
            },
            MaxTokens: 1024,
        })
        if err != nil {
            http.Error(w, "AI error", http.StatusInternalServerError)
            return
        }

        json.NewEncoder(w).Encode(map[string]any{
            "answer":    resp.Content[0].Text,
            "from_cache": resp.Metadata["cache_hit"] == "true",
        })
    }
}
```

## Example 4: Multi-Provider with Fallback

```go
type MultiProvider struct {
    primary   provider.Provider
    secondary provider.Provider
    breakers  map[string]*circuitbreaker.CircuitBreaker
}

func (m *MultiProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
    // Try primary
    if m.breakers["primary"].Allow() {
        resp, err := m.primary.Complete(ctx, req)
        if err == nil {
            m.breakers["primary"].RecordSuccess()
            return resp, nil
        }
        m.breakers["primary"].RecordFailure()
        log.Warnf("Primary provider failed: %v", err)
    }

    // Try secondary
    if m.breakers["secondary"].Allow() {
        resp, err := m.secondary.Complete(ctx, req)
        if err == nil {
            m.breakers["secondary"].RecordSuccess()
            return resp, nil
        }
        m.breakers["secondary"].RecordFailure()
    }

    return nil, errors.New("all providers unavailable")
}

// Usage
multi := &MultiProvider{
    primary:   provider.NewClaude(provider.WithAPIKey(claudeKey)),
    secondary: provider.NewOpenAI(provider.WithAPIKey(openaiKey)),
    breakers: map[string]*circuitbreaker.CircuitBreaker{
        "primary":   circuitbreaker.New(circuitbreaker.WithFailureThreshold(3)),
        "secondary": circuitbreaker.New(circuitbreaker.WithFailureThreshold(3)),
    },
}
```

## Related Documentation

- [README](README.md) — Module overview
- [Provider Interface](provider.md) — LLM providers
- [Session Management](session.md) — Conversation history
- [SSE Streaming](sse.md) — Real-time streaming
- [Tool Framework](tool.md) — Agent tools
