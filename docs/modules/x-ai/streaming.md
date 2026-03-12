# Streaming Response Processing

> **Package**: `github.com/spcent/plumego/x/ai/streaming` | **Feature**: Stream handling

## Overview

The `streaming` package provides utilities for processing LLM streaming responses — including parallel stream aggregation, stream transformation, and testing helpers.

## Quick Start

```go
import "github.com/spcent/plumego/x/ai/streaming"

// Stream and collect full response
handler := streaming.NewStreamHandler(claude)

result, err := handler.StreamAndCollect(ctx, &provider.CompletionRequest{
    Model:    "claude-opus-4-6",
    Messages: messages,
    Stream:   true,
})

fullText := result.Text   // Complete response text
usage := result.Usage     // Token usage
```

## Stream Handler

```go
handler := streaming.NewStreamHandler(claude,
    streaming.WithBufferSize(4096),
    streaming.WithFlushInterval(50*time.Millisecond),
)
```

### Stream and Collect

```go
// Get the full response after streaming completes
result, err := handler.StreamAndCollect(ctx, req)
if err != nil {
    return err
}

fmt.Println(result.Text)              // Complete text
fmt.Println(result.Usage.Total)       // Total tokens
fmt.Println(result.FinishReason)      // "end_turn", "max_tokens"
```

### Stream to Writer

```go
// Write tokens to any io.Writer as they arrive
var buf bytes.Buffer
err := handler.StreamToWriter(ctx, req, &buf)

// Works with HTTP response writer
err = handler.StreamToWriter(ctx, req, w)
```

### Stream with Callbacks

```go
err := handler.Stream(ctx, req, &streaming.Callbacks{
    OnToken: func(text string) {
        // Called for each text token
        fmt.Print(text)
    },
    OnToolCall: func(call provider.ToolCall) {
        // Called when model requests tool use
        result := executeTool(call)
        toolResults = append(toolResults, result)
    },
    OnComplete: func(result streaming.Result) {
        // Called when stream finishes
        log.Printf("Done: %d tokens", result.Usage.Total)
    },
    OnError: func(err error) {
        log.Errorf("Stream error: %v", err)
    },
})
```

## Parallel Streaming

Process multiple LLM requests concurrently:

```go
// Define multiple requests
requests := []*provider.CompletionRequest{
    {Model: "claude-opus-4-6", Messages: []provider.Message{{Role: "user", Content: "Explain async"}}},
    {Model: "claude-opus-4-6", Messages: []provider.Message{{Role: "user", Content: "Explain channels"}}},
    {Model: "claude-opus-4-6", Messages: []provider.Message{{Role: "user", Content: "Explain goroutines"}}},
}

// Execute in parallel with concurrency limit
results, err := streaming.ParallelComplete(ctx, claude, requests,
    streaming.WithConcurrency(3),
    streaming.WithTimeout(30*time.Second),
)

for i, result := range results {
    fmt.Printf("Response %d: %s\n", i, result.Text)
}
```

## Stream Transformation

Transform streaming output before sending to client:

```go
// Filter sensitive content during streaming
filtered := streaming.Transform(reader, func(text string) string {
    return redactor.RemovePII(text)
})

// Add markdown formatting
formatted := streaming.Transform(reader, func(text string) string {
    return markdownRenderer.Render(text)
})

// Pipe transformed stream to SSE
for {
    delta, err := filtered.Next()
    if err == io.EOF {
        break
    }
    sseStream.Send(sse.Event{Data: delta.Text})
}
```

## Streaming with Tool Calls

Handle tool calls mid-stream in an agentic loop:

```go
func streamWithTools(ctx context.Context, messages []provider.Message, w http.ResponseWriter) error {
    sseStream, _ := sse.NewStream(ctx, w)
    defer sseStream.Close()

    for iteration := 0; iteration < 10; iteration++ {
        handler := streaming.NewStreamHandler(claude)
        var toolCalls []provider.ToolCall

        err := handler.Stream(ctx, &provider.CompletionRequest{
            Model:      "claude-opus-4-6",
            Messages:   messages,
            Tools:      registry.ProviderTools(),
            ToolChoice: provider.ToolChoiceAuto,
            MaxTokens:  4096,
        }, &streaming.Callbacks{
            OnToken: func(text string) {
                // Send text tokens to client in real-time
                sseStream.Send(sse.Event{
                    Event: "token",
                    Data:  fmt.Sprintf(`{"text":%q}`, text),
                })
            },
            OnToolCall: func(call provider.ToolCall) {
                toolCalls = append(toolCalls, call)
            },
        })
        if err != nil {
            return err
        }

        // No tool calls = we're done
        if len(toolCalls) == 0 {
            sseStream.Send(sse.Event{Event: "done", Data: "[DONE]"})
            return nil
        }

        // Execute tool calls and continue loop
        var toolResults []provider.ContentPart
        for _, call := range toolCalls {
            result, _ := registry.Execute(ctx, call.Name, call.Input)
            toolResults = append(toolResults, provider.ContentPart{
                Type:      "tool_result",
                ToolUseID: call.ID,
                Content:   result,
            })
            // Notify client of tool execution
            sseStream.Send(sse.Event{
                Event: "tool_use",
                Data:  fmt.Sprintf(`{"tool":%q}`, call.Name),
            })
        }

        messages = append(messages, provider.Message{Role: "user", Parts: toolResults})
    }

    return errors.New("max iterations reached")
}
```

## Testing Helpers

```go
import "github.com/spcent/plumego/x/ai/streaming"

// Create mock streaming provider for tests
mockProvider := streaming.NewMockProvider(
    streaming.MockResponse("Hello, I am Claude!"),
)

// Test streaming handler
handler := streaming.NewStreamHandler(mockProvider)
result, err := handler.StreamAndCollect(ctx, req)

assert.NoError(t, err)
assert.Equal(t, "Hello, I am Claude!", result.Text)
```

### Controlled Token Delivery

```go
// Mock with configurable delays between tokens
mockProvider := streaming.NewMockProvider(
    streaming.MockStreamTokens(
        []string{"Hello", ",", " ", "world", "!"},
        10*time.Millisecond, // Delay between tokens
    ),
)
```

## Buffering and Backpressure

```go
// Configure buffer to handle slow clients
handler := streaming.NewStreamHandler(claude,
    streaming.WithBufferSize(8192),          // 8KB token buffer
    streaming.WithFlushInterval(50*time.Millisecond), // Flush every 50ms
    streaming.WithSlowClientTimeout(5*time.Second),   // Disconnect slow clients
)
```

## Error Recovery

```go
err := handler.Stream(ctx, req, &streaming.Callbacks{
    OnToken: func(text string) {
        // Write to client
        if err := sseStream.Send(sse.Event{Data: text}); err != nil {
            // Client disconnected, cancel context
            cancel()
        }
    },
    OnError: func(err error) {
        if errors.Is(err, context.Canceled) {
            // Client disconnected, expected
            return
        }
        log.Errorf("Stream error: %v", err)
        sseStream.SendError("Stream interrupted")
    },
})
```

## Related Documentation

- [SSE Streaming](sse.md) — SSE protocol implementation
- [Provider Interface](provider.md) — Streaming API
- [Session Management](session.md) — Session-aware streaming
