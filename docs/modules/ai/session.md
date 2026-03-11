# Session Management

> **Package**: `github.com/spcent/plumego/ai/session`

## Overview

The current `session` package manages persisted conversation state with explicit storage, explicit configuration, and explicit token-window trimming.

## Quick Start

```go
mgr := session.NewManager(
    session.NewMemoryStorage(),
    session.WithConfig(session.Config{
        DefaultTTL:  24 * time.Hour,
        MaxMessages: 1000,
        MaxTokens:   100000,
        AutoTrim:    true,
    }),
)

sess, err := mgr.Create(ctx, session.CreateOptions{Model: "claude-3-sonnet-20240229"})
if err != nil {
    return err
}

if err := mgr.AppendMessage(ctx, sess.ID, provider.NewTextMessage(provider.RoleUser, "Tell me about Go generics")); err != nil {
    return err
}

messages, err := mgr.GetActiveContext(ctx, sess.ID, 100000)
if err != nil {
    return err
}
_ = messages
```

## Session shape

`Session` stores:

- identity fields: `ID`, `TenantID`, `UserID`, `AgentID`, `Model`
- history: `Messages`
- per-session state: `Context`, `Metadata`
- usage and lifecycle: `Usage`, `CreatedAt`, `UpdatedAt`, `ExpiresAt`

## Main operations

### Create and load

```go
sess, err := mgr.Create(ctx, session.CreateOptions{
    UserID:   "user-123",
    TenantID: "tenant-abc",
    Model:    "claude-3-sonnet-20240229",
})

loaded, err := mgr.Get(ctx, sess.ID)
```

### Append messages

```go
_ = mgr.AppendMessage(ctx, sess.ID, provider.NewTextMessage(provider.RoleUser, "question"))
_ = mgr.AppendMessage(ctx, sess.ID, provider.NewTextMessage(provider.RoleAssistant, "answer"))
```

### Token-window selection and trimming

```go
messages, err := mgr.GetActiveContext(ctx, sess.ID, 4000)
if err != nil {
    return err
}

if err := mgr.TrimMessages(ctx, sess.ID, 4000); err != nil {
    return err
}
_ = messages
```

### Context values

```go
_ = mgr.SetContext(ctx, sess.ID, "topic", "kubernetes")
value, err := mgr.GetContext(ctx, sess.ID, "topic")
if err != nil {
    return err
}
_ = value
```

### Delete

```go
if err := mgr.Delete(ctx, sess.ID); err != nil {
    return err
}
```

## Storage interface

Current storage contract:

```go
type Storage interface {
    Save(ctx context.Context, session *Session) error
    Load(ctx context.Context, sessionID string) (*Session, error)
    Delete(ctx context.Context, sessionID string) error
}
```

The built-in implementation is `session.NewMemoryStorage()`.
