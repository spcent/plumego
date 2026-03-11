# LLM Provider Interface

> **Package**: `github.com/spcent/plumego/ai/provider`

## Overview

`provider.Provider` is the shared interface implemented by the built-in Claude and OpenAI adapters.

## Interface

```go
type Provider interface {
    Name() string
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    CompleteStream(ctx context.Context, req *CompletionRequest) (*StreamReader, error)
    ListModels(ctx context.Context) ([]Model, error)
    GetModel(ctx context.Context, modelID string) (*Model, error)
    CountTokens(text string) (int, error)
}
```

## Built-in providers

### Claude

```go
claude := provider.NewClaudeProvider(
    os.Getenv("ANTHROPIC_API_KEY"),
    provider.WithClaudeHTTPClient(&http.Client{}),
)
```

### OpenAI

```go
openai := provider.NewOpenAIProvider(
    os.Getenv("OPENAI_API_KEY"),
    provider.WithOpenAIHTTPClient(&http.Client{}),
)
```

Custom base URLs are available with `WithClaudeBaseURL(...)` and `WithOpenAIBaseURL(...)`.

## Request shape

```go
req := &provider.CompletionRequest{
    Model: "claude-3-sonnet-20240229",
    Messages: []provider.Message{
        provider.NewTextMessage(provider.RoleUser, "Explain Go interfaces"),
    },
    System:      "You are a concise Go assistant.",
    MaxTokens:   512,
    Temperature: 0.2,
    Stream:      false,
}
```

## Response handling

```go
resp, err := claude.Complete(ctx, req)
if err != nil {
    return err
}

fmt.Println(resp.GetText())
fmt.Println(resp.StopReason)
fmt.Println(resp.Usage.TotalTokens)
```

## Streaming

```go
reader, err := claude.CompleteStream(ctx, &provider.CompletionRequest{
    Model:    "claude-3-sonnet-20240229",
    Messages: []provider.Message{provider.NewTextMessage(provider.RoleUser, "hello")},
    Stream:   true,
})
if err != nil {
    return err
}
defer reader.Close()

for {
    chunk, err := reader.Next()
    if err == io.EOF {
        break
    }
    if err != nil {
        return err
    }
    if chunk.Delta != nil && chunk.Delta.Text != "" {
        fmt.Print(chunk.Delta.Text)
    }
}
```

## Provider manager

```go
mgr := provider.NewManager()
mgr.Register(claude)
mgr.Register(openai)

resp, err := mgr.Complete(ctx, &provider.CompletionRequest{
    Model: "gpt-4",
    Messages: []provider.Message{
        provider.NewTextMessage(provider.RoleUser, "Summarize this"),
    },
})
```

The default router matches providers by model prefix when possible and otherwise falls back to the first registered provider.
