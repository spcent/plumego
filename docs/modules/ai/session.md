# Session Management

> **Package**: `github.com/spcent/plumego/ai/session` | **Feature**: Conversation history

## Overview

The `session` package manages AI conversation sessions, maintaining message history, context window limits, and per-session metadata. Sessions enable multi-turn conversations by persisting the exchange between user and AI.

## Quick Start

```go
import "github.com/spcent/plumego/ai/session"

// Create manager with in-memory storage
mgr := session.NewManager(
    session.NewInMemoryStorage(),
    session.WithMaxTokens(100000),       // Context window limit
    session.WithDefaultTTL(24*time.Hour), // Session expiry
    session.WithAutoTrim(true),           // Auto-trim when near limit
)

// Create or retrieve a session
sess, err := mgr.GetOrCreate(ctx, sessionID)

// Add a user message
err = mgr.AddMessage(ctx, sess.ID, provider.Message{
    Role:    "user",
    Content: "Tell me about Go generics",
})

// Get messages for LLM request
messages := sess.Messages

// Update after AI response
err = mgr.AddMessage(ctx, sess.ID, provider.Message{
    Role:    "assistant",
    Content: aiResponse,
})
```

## Session Structure

```go
type Session struct {
    ID        string            // Unique session identifier
    TenantID  string            // Multi-tenant isolation
    UserID    string            // User association
    AgentID   string            // AI agent identifier
    Model     string            // LLM model for this session
    Messages  []provider.Message // Conversation history
    Context   map[string]any    // Arbitrary session context
    Usage     tokenizer.TokenUsage // Cumulative token usage
    CreatedAt time.Time
    UpdatedAt time.Time
    ExpiresAt time.Time
    Metadata  map[string]string // Custom metadata
}
```

## Manager Options

```go
mgr := session.NewManager(storage,
    session.WithMaxTokens(200000),          // 200k token context
    session.WithMaxMessages(1000),          // Max messages before compaction
    session.WithDefaultTTL(7*24*time.Hour), // 1 week expiry
    session.WithAutoTrim(true),             // Trim old messages automatically
    session.WithTokenizer(customTokenizer), // Custom token counter
)
```

## Session Lifecycle

### Create

```go
// Auto-generate ID
sess, err := mgr.Create(ctx, session.Options{
    UserID:   "user-123",
    TenantID: "tenant-abc",
    Model:    "claude-opus-4-6",
    Metadata: map[string]string{
        "source": "web-chat",
    },
})
```

### Get or Create

```go
// Typical usage: resume if exists, create otherwise
sess, err := mgr.GetOrCreate(ctx, sessionID)
if err != nil {
    return err
}
```

### Add Messages

```go
// Add user message
err = mgr.AddMessage(ctx, sess.ID, provider.Message{
    Role:    "user",
    Content: userInput,
})

// Add assistant response
err = mgr.AddMessage(ctx, sess.ID, provider.Message{
    Role:    "assistant",
    Content: assistantResponse,
})
```

### Session Context

```go
// Store arbitrary data in session
err = mgr.SetContext(ctx, sess.ID, "user_preference", "detailed_responses")
err = mgr.SetContext(ctx, sess.ID, "current_topic", "kubernetes")

// Retrieve context
sess, _ := mgr.Get(ctx, sessionID)
preference := sess.Context["user_preference"].(string)
```

### Delete

```go
err = mgr.Delete(ctx, sess.ID)
```

## Context Window Management

When conversation history grows large, the manager automatically trims older messages while preserving context.

### Auto-Trim

```go
mgr := session.NewManager(storage,
    session.WithMaxTokens(100000),
    session.WithAutoTrim(true),
)

// When messages exceed 100k tokens, older messages are removed
// System message and recent context are always preserved
```

### Manual Compaction

```go
// Summarize old messages to free context
summary, err := claude.Complete(ctx, &provider.CompletionRequest{
    Model: "claude-haiku-4-5-20251001",
    Messages: []provider.Message{
        {Role: "user", Content: "Summarize this conversation: " + conversationText},
    },
    MaxTokens: 500,
})

// Replace old messages with summary
err = mgr.SetMessages(ctx, sess.ID, []provider.Message{
    {Role: "system", Content: "Previous conversation summary: " + summary.Content[0].Text},
})
```

## Storage Backends

### In-Memory (Development/Testing)

```go
storage := session.NewInMemoryStorage()
mgr := session.NewManager(storage)
```

### Redis (Production)

```go
import "github.com/spcent/plumego/store/cache"

redisClient := cache.NewRedisClient(os.Getenv("REDIS_URL"))
storage := session.NewRedisStorage(redisClient,
    session.WithKeyPrefix("ai:sessions:"),
)
mgr := session.NewManager(storage)
```

### Custom Storage

```go
type Storage interface {
    Get(ctx context.Context, sessionID string) (*Session, error)
    Set(ctx context.Context, session *Session) error
    Delete(ctx context.Context, sessionID string) error
    List(ctx context.Context, userID string) ([]*Session, error)
    Cleanup(ctx context.Context) error // Remove expired sessions
}
```

## Token Usage Tracking

```go
sess, _ := mgr.Get(ctx, sessionID)

fmt.Printf("Total tokens used: %d\n", sess.Usage.Total)
fmt.Printf("Input tokens: %d\n", sess.Usage.Input)
fmt.Printf("Output tokens: %d\n", sess.Usage.Output)

// Estimated cost (if tracked)
cost := float64(sess.Usage.Total) / 1000 * 0.015 // $0.015 per 1k tokens
fmt.Printf("Estimated cost: $%.4f\n", cost)
```

## Multi-Turn Conversation Pattern

```go
func handleChat(w http.ResponseWriter, r *http.Request) {
    var req struct {
        SessionID string `json:"session_id"`
        Message   string `json:"message"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Get or create session
    sess, err := sessionMgr.GetOrCreate(r.Context(), req.SessionID)
    if err != nil {
        http.Error(w, "Session error", http.StatusInternalServerError)
        return
    }

    // Add user message
    userMsg := provider.Message{Role: "user", Content: req.Message}
    sessionMgr.AddMessage(r.Context(), sess.ID, userMsg)

    // Build request with full conversation history
    resp, err := claude.Complete(r.Context(), &provider.CompletionRequest{
        Model:    "claude-opus-4-6",
        System:   "You are a helpful assistant.",
        Messages: sess.Messages,
        MaxTokens: 2048,
    })
    if err != nil {
        http.Error(w, "AI error", http.StatusInternalServerError)
        return
    }

    assistantText := resp.Content[0].Text

    // Save assistant response
    sessionMgr.AddMessage(r.Context(), sess.ID, provider.Message{
        Role:    "assistant",
        Content: assistantText,
    })

    json.NewEncoder(w).Encode(map[string]any{
        "session_id": sess.ID,
        "response":   assistantText,
        "usage":      sess.Usage,
    })
}
```

## Session API Endpoints

```go
// List user sessions
app.Get("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
    userID := auth.UserIDFromContext(r.Context())
    sessions, _ := sessionMgr.List(r.Context(), userID)
    json.NewEncoder(w).Encode(sessions)
})

// Delete session
app.Delete("/api/sessions/:id", func(w http.ResponseWriter, r *http.Request) {
    sessionID := r.PathValue("id")
    sessionMgr.Delete(r.Context(), sessionID)
    w.WriteHeader(http.StatusNoContent)
})

// Get session history
app.Get("/api/sessions/:id/messages", func(w http.ResponseWriter, r *http.Request) {
    sess, err := sessionMgr.Get(r.Context(), r.PathValue("id"))
    if err != nil {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(sess.Messages)
})
```

## Best Practices

### 1. Bound Session Lifetime

```go
// Always set TTL to prevent memory leaks
session.WithDefaultTTL(24 * time.Hour)
```

### 2. Enable Auto-Trim

```go
// Prevent context overflow errors
session.WithAutoTrim(true),
session.WithMaxTokens(80000), // Leave buffer below model limit
```

### 3. Session Isolation

```go
// Always include tenant/user for multi-tenant apps
sess, _ := mgr.Create(ctx, session.Options{
    TenantID: tenantID,
    UserID:   userID,
})
```

### 4. Periodic Cleanup

```go
// Run cleanup job to remove expired sessions
scheduler.AddCron("session-cleanup", "0 * * * *", func(ctx context.Context) error {
    return sessionMgr.Cleanup(ctx)
})
```

## Related Documentation

- [Provider Interface](provider.md) — LLM providers
- [SSE Streaming](sse.md) — Real-time streaming with sessions
- [Token Counting](tokenizer.md) — Token usage tracking
