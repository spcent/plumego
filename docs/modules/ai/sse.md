# Server-Sent Events (SSE) Streaming

> **Package**: `github.com/spcent/plumego/ai/sse` | **Feature**: Real-time AI streaming

## Overview

The `sse` package provides Server-Sent Events support for streaming AI responses. SSE enables real-time, token-by-token delivery to web clients over standard HTTP, without requiring WebSockets.

## Quick Start

```go
import "github.com/spcent/plumego/ai/sse"

func handleStream(w http.ResponseWriter, r *http.Request) {
    // Create SSE stream (sets Content-Type: text/event-stream)
    stream, err := sse.NewStream(r.Context(), w)
    if err != nil {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }
    defer stream.Close()

    // Stream AI tokens
    reader, _ := claude.CompleteStream(r.Context(), req)
    defer reader.Close()

    for {
        delta, err := reader.Next()
        if err == io.EOF {
            stream.Send(sse.Event{Event: "done", Data: "[DONE]"})
            break
        }
        stream.Send(sse.Event{Data: delta.Text})
    }
}
```

## SSE Event Format

### Event Structure

```go
type Event struct {
    Event   string // Event type (optional)
    Data    string // Event data (required)
    ID      string // Event ID for client resumption (optional)
    Retry   int    // Reconnect interval in milliseconds (optional)
    Comment string // Comment line for keep-alive (optional)
}
```

### Wire Format

```
id: 42
event: token
data: {"text": "Hello"}

data: {"text": " world"}

event: done
data: [DONE]

```

Each event is separated by a blank line. The `data` field can contain JSON.

## Stream Operations

### Send Event

```go
// Simple text event
stream.Send(sse.Event{Data: "Hello"})

// Typed event with JSON data
stream.Send(sse.Event{
    Event: "token",
    Data:  `{"text": "world", "index": 5}`,
    ID:    "42",
})

// Done signal
stream.Send(sse.Event{Event: "done", Data: "[DONE]"})
```

### Send Error

```go
stream.SendError("AI provider unavailable")

// Sends:
// event: error
// data: {"error": "AI provider unavailable"}
```

### Keep-Alive

```go
stream, _ := sse.NewStream(r.Context(), w,
    sse.WithKeepAlive(15*time.Second), // Send ping every 15s
)
```

Keep-alive prevents proxy timeouts for long-running streams:
```
: keep-alive

: keep-alive

```

## Full Streaming Chat Handler

```go
func handleChatStream(w http.ResponseWriter, r *http.Request) {
    var req struct {
        SessionID string `json:"session_id"`
        Message   string `json:"message"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Create SSE stream
    stream, err := sse.NewStream(r.Context(), w,
        sse.WithKeepAlive(20*time.Second),
    )
    if err != nil {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }
    defer stream.Close()

    // Get session
    sess, err := sessionMgr.GetOrCreate(r.Context(), req.SessionID)
    if err != nil {
        stream.SendError("session error")
        return
    }

    // Add user message
    sessionMgr.AddMessage(r.Context(), sess.ID, provider.Message{
        Role:    "user",
        Content: req.Message,
    })

    // Start LLM streaming
    reader, err := claude.CompleteStream(r.Context(), &provider.CompletionRequest{
        Model:     "claude-opus-4-6",
        System:    "You are a helpful assistant.",
        Messages:  sess.Messages,
        MaxTokens: 2048,
        Stream:    true,
    })
    if err != nil {
        stream.SendError(fmt.Sprintf("AI error: %v", err))
        return
    }
    defer reader.Close()

    // Accumulate full response for session
    var fullResponse strings.Builder

    // Stream tokens to client
    for {
        delta, err := reader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            stream.SendError("stream interrupted")
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

    // Save assistant response
    sessionMgr.AddMessage(r.Context(), sess.ID, provider.Message{
        Role:    "assistant",
        Content: fullResponse.String(),
    })

    // Send completion event
    stream.Send(sse.Event{
        Event: "done",
        Data:  fmt.Sprintf(`{"session_id":%q}`, sess.ID),
    })
}
```

## Client-Side JavaScript

```javascript
const eventSource = new EventSource('/api/chat/stream', {
    // For POST requests, use fetch + ReadableStream instead
});

eventSource.addEventListener('token', (event) => {
    const data = JSON.parse(event.data);
    document.getElementById('output').textContent += data.text;
});

eventSource.addEventListener('done', (event) => {
    eventSource.close();
    console.log('Stream complete');
});

eventSource.addEventListener('error', (event) => {
    console.error('Stream error:', event.data);
    eventSource.close();
});
```

### POST with fetch + ReadableStream

```javascript
const response = await fetch('/api/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message: userMessage, session_id: sessionId })
});

const reader = response.body.getReader();
const decoder = new TextDecoder();
let buffer = '';

while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');
    buffer = lines.pop(); // Keep incomplete line

    for (const line of lines) {
        if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data === '[DONE]') break;
            const parsed = JSON.parse(data);
            output.textContent += parsed.text;
        }
    }
}
```

## React Hook Example

```javascript
function useStreamingChat() {
    const [response, setResponse] = useState('');
    const [isStreaming, setIsStreaming] = useState(false);

    const sendMessage = async (message, sessionId) => {
        setIsStreaming(true);
        setResponse('');

        const res = await fetch('/api/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ message, session_id: sessionId })
        });

        const reader = res.body.getReader();
        const decoder = new TextDecoder();

        while (true) {
            const { done, value } = await reader.read();
            if (done) break;

            const text = decoder.decode(value);
            for (const line of text.split('\n')) {
                if (line.startsWith('data: ')) {
                    const data = line.slice(6);
                    if (data.startsWith('{')) {
                        const { text } = JSON.parse(data);
                        if (text) setResponse(prev => prev + text);
                    }
                }
            }
        }
        setIsStreaming(false);
    };

    return { response, isStreaming, sendMessage };
}
```

## Proxy and Load Balancer Configuration

SSE requires specific proxy configuration to prevent buffering:

### Nginx

```nginx
location /api/chat {
    proxy_pass http://backend;
    proxy_http_version 1.1;

    # Disable buffering for SSE
    proxy_buffering off;
    proxy_cache off;

    # Keep connection alive
    proxy_set_header Connection '';
    chunked_transfer_encoding on;

    # Increase timeout for long streams
    proxy_read_timeout 300s;
}
```

### Caddy

```
/api/chat {
    reverse_proxy backend:8080 {
        flush_interval -1
    }
}
```

## Stream Options

```go
stream, err := sse.NewStream(r.Context(), w,
    sse.WithKeepAlive(15*time.Second),       // Keep-alive interval
    sse.WithRetryInterval(3000),             // Client retry on disconnect (ms)
    sse.WithBufferSize(4096),                // Write buffer size
)
```

## Error Recovery

```go
// Client-side reconnection with Last-Event-ID
eventSource.addEventListener('open', () => {
    // Browser sends Last-Event-ID header on reconnect
    console.log('Connected');
});

// Server-side resumption
lastEventID := r.Header.Get("Last-Event-ID")
if lastEventID != "" {
    // Resume from this event ID
    events, _ := eventStore.GetAfter(lastEventID)
    for _, e := range events {
        stream.Send(e)
    }
}
```

## Related Documentation

- [Provider Interface](provider.md) — LLM streaming API
- [Session Management](session.md) — Conversation sessions
- [Streaming Processing](streaming.md) — Stream processing patterns
