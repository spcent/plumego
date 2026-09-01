# bodylimit — Request Body Size Limit Middleware

Enforces a maximum request body size.

## Import

```go
import "github.com/spcent/plumego/middleware/bodylimit"
```

## Overview

`bodylimit.Middleware` wraps request bodies with a size limit. Requests exceeding `MaxBytes` receive `413 Payload Too Large` before the handler runs, protecting downstream parsers from memory exhaustion.

## Exports

| Symbol | Description |
|--------|-------------|
| `Middleware(config Config)` | Constructor |
| `DefaultConfig(maxBytes, logger)` | Sensible defaults with a logger |
| `Config{MaxBytes, Logger}` | Configuration |

## Usage

```go
app.Use(bodylimit.Middleware(bodylimit.DefaultConfig(1<<20, logger))) // 1 MiB
```

## Notes

- Limits apply to the body reader, not request headers.
- The wrapped `ResponseWriter` supports `http.Flusher` and `http.Hijacker` passthrough.
- The limit is checked lazily during `Read`; a handler that never reads the body will not trigger the check.
