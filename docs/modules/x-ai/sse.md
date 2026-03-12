# Server-Sent Events (SSE) Streaming

> **Package**: `github.com/spcent/plumego/x/ai/sse`

## Overview

The current SSE package exposes a small explicit API:

- `sse.NewStream(ctx, w)`
- `stream.Send(...)`
- `stream.SendData(...)`
- `stream.SendJSON(...)`
- `stream.SendComment(...)`
- `stream.SetKeepAliveInterval(...)`
- `stream.Close()`
- `sse.Handle(...)`

## Quick Start

```go
func handleStream(w http.ResponseWriter, r *http.Request) {
    stream, err := sse.NewStream(r.Context(), w)
    if err != nil {
        http.Error(w, "streaming not supported", http.StatusInternalServerError)
        return
    }
    defer stream.Close()

    stream.SetKeepAliveInterval(15 * time.Second)

    for _, token := range []string{"hello", " ", "world"} {
        if err := stream.Send(&sse.Event{Event: "token", Data: token}); err != nil {
            return
        }
    }

    _ = stream.Send(&sse.Event{Event: "done", Data: "[DONE]"})
}
```

## Event shape

```go
type Event struct {
    Event   string
    Data    string
    ID      string
    Retry   int
    Comment string
}
```

## Convenience methods

```go
_ = stream.SendData("plain text")
_ = stream.SendJSON("token", `{"text":"hello"}`)
_ = stream.SendComment("keep-alive")
```

## AI streaming example

```go
reader, err := model.CompleteStream(ctx, &provider.CompletionRequest{
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
        if err := stream.Send(&sse.Event{Event: "token", Data: chunk.Delta.Text}); err != nil {
            return err
        }
    }
}
```

## Handler wrapper

```go
http.HandleFunc("/events", sse.Handle(func(stream *sse.Stream) error {
    stream.SetKeepAliveInterval(10 * time.Second)
    return stream.Send(&sse.Event{Event: "ready", Data: "ok"})
}))
```
