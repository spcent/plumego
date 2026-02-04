# AI Agent Gateway Example

This example demonstrates how to build a code agent gateway using Plumego's new AI capabilities (Phase 1).

## Features

- ✅ **SSE (Server-Sent Events)**: Stream real-time AI responses
- ✅ **LLM Provider Abstraction**: Unified interface for Claude and OpenAI
- ✅ **Session Management**: Maintain conversation context
- ✅ **Token Counting**: Track and limit token usage
- ✅ **Tool Framework**: Execute functions from AI models

## Quick Start

### 1. Set API Key (Optional)

```bash
export CLAUDE_API_KEY="your-api-key"
```

If not set, the gateway will run in mock mode for demonstration.

### 2. Run the Server

```bash
go run examples/ai-agent-gateway/main.go
```

The server starts on `http://localhost:8080`

### 3. Try the API

#### Create a Session

```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"model": "claude-3-opus-20240229"}'
```

Response:
```json
{
  "session_id": "sess_1234567890",
  "model": "claude-3-opus-20240229",
  "created_at": "2024-02-04T10:00:00Z"
}
```

#### Send a Message

```bash
curl -X POST http://localhost:8080/api/sessions/{session_id}/messages \
  -H "Content-Type: application/json" \
  -d '{"message": "Calculate 5 + 3"}'
```

#### Stream Responses (SSE)

```bash
curl -N http://localhost:8080/api/sessions/{session_id}/stream
```

Output:
```
event: message
data: {"type": "welcome", "text": "Connected to AI Agent Gateway"}

event: chunk
data: {"index": 0, "text": "Hello! "}

event: chunk
data: {"index": 1, "text": "I'm your "}

...
```

#### List Available Tools

```bash
curl http://localhost:8080/api/tools
```

Response:
```json
{
  "tools": [
    {
      "name": "echo",
      "description": "Echoes back the input message",
      "parameters": { ... }
    },
    {
      "name": "calculator",
      "description": "Performs basic arithmetic operations",
      "parameters": { ... }
    },
    {
      "name": "get_timestamp",
      "description": "Returns the current timestamp",
      "parameters": { ... }
    }
  ],
  "count": 3
}
```

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       v
┌─────────────────────────┐
│   HTTP/SSE Endpoints    │
├─────────────────────────┤
│  Session Manager        │  ← Manages conversations
├─────────────────────────┤
│  Provider Manager       │  ← Routes to LLM providers
├─────────────────────────┤
│  Tool Registry          │  ← Function calling
├─────────────────────────┤
│  Token Counter          │  ← Track usage
└─────────────────────────┘
       │
       v
┌───────────────────────────┐
│  Claude / OpenAI / etc.   │
└───────────────────────────┘
```

## Code Walkthrough

### Session Creation

```go
sessionMgr := session.NewManager(
    session.NewMemoryStorage(),
    session.WithTokenizer(tokenizer.NewClaudeTokenizer("claude-3")),
)

sess, err := sessionMgr.Create(ctx, session.CreateOptions{
    Model: "claude-3-opus-20240229",
})
```

### Provider Setup

```go
providerMgr := provider.NewManager()
claudeProvider := provider.NewClaudeProvider(apiKey)
providerMgr.Register(claudeProvider)
```

### Tool Registration

```go
toolRegistry := tool.NewRegistry()
toolRegistry.Register(tool.NewEchoTool())
toolRegistry.Register(tool.NewCalculatorTool())
toolRegistry.Register(tool.NewTimestampTool())
```

### SSE Streaming

```go
app.Get("/stream", sse.Handle(func(s *sse.Stream) error {
    if err := s.SendData("Hello"); err != nil {
        return err
    }
    return s.SendJSON("done", `{"status": "complete"}`)
}))
```

## Production Considerations

This is a Phase 1 demonstration. For production use, add:

- [ ] Authentication & authorization
- [ ] Persistent storage (PostgreSQL/Redis)
- [ ] Rate limiting per tenant
- [ ] Content filtering
- [ ] Distributed sessions
- [ ] Monitoring & metrics
- [ ] Error handling & retries
- [ ] Tool sandboxing

## Next Steps

Phase 2 will add:
- Prompt template engine
- Multi-model routing
- Intelligent caching
- Agent orchestration
- Advanced tool sandboxing

## License

MIT
