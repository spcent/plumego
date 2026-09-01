# timeout — Request Timeout Middleware

Enforces a maximum duration for request processing.

## Import

```go
import "github.com/spcent/plumego/middleware/timeout"
```

## Overview

`timeout.Middleware` cancels the request context when the deadline is exceeded and returns `504 Gateway Timeout`. Handlers should observe `ctx.Done()` to stop work promptly.

## Exports

| Symbol | Description |
|--------|-------------|
| `Middleware(cfg Config)` | Constructor; returns a `middleware.Middleware` |
| `Config{Timeout, MaxBufferBytes, MaxReplayBytes}` | Timeout and buffer limits |
| `DefaultConfig()` | Sensible defaults (30 s timeout) |

## Usage

```go
cfg := timeout.DefaultConfig()
cfg.Timeout = 10 * time.Second
app.Use(timeout.Middleware(cfg))
```

## Notes

- Responses larger than `MaxReplayBytes` cannot be replayed and become a structured 504 instead of streamed content.
- Timeout does not forcibly kill goroutines; cooperative context cancellation is required.
