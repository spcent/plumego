# concurrencylimit — Concurrent Request Limit Middleware

Limits the number of concurrently processed requests.

## Import

```go
import "github.com/spcent/plumego/middleware/concurrencylimit"
```

## Overview

`concurrencylimit.Middleware` rejects requests with `503 Service Unavailable` when the active request count reaches the configured maximum. It protects the service from overload during traffic spikes.

## Exports

| Symbol | Description |
|--------|-------------|
| `Middleware(config Config)` | Constructor |
| `DefaultConfig(maxConcurrent)` | Creates config with the given limit |
| `Config{MaxConcurrent}` | Limit configuration |

## Usage

```go
app.Use(concurrencylimit.Middleware(concurrencylimit.DefaultConfig(100)))
```

## Notes

- The counter is decremented when the response writer is closed or the handler panics.
- Place this middleware before expensive work (after `recovery`, before business handlers).
