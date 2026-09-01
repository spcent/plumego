# compression — Gzip Response Compression Middleware

Conditionally compresses HTTP responses with gzip.

## Import

```go
import "github.com/spcent/plumego/middleware/compression"
```

## Overview

`compression.Middleware` buffers small responses to decide whether compression is worthwhile and streams large responses directly when compression savings are confirmed.

## Exports

| Symbol | Description |
|--------|-------------|
| `Middleware(cfg Config)` | Constructor |
| `DefaultConfig()` | Sensible defaults (gzip level default, 512 B threshold) |
| `Config{Level, MinSize, ContentTypes, ExcludePaths}` | Fine-grained control |

## Usage

```go
app.Use(compression.Middleware(compression.DefaultConfig()))
```

## Notes

- The middleware ignores already-compressed content types (`image/*`, `video/*`, `application/gzip`, etc.).
- `Content-Encoding: gzip` is set only when compression is actually applied.
- Responses smaller than `MinSize` are passed through uncompressed to avoid gzip overhead.
- The wrapped `ResponseWriter` supports `Unwrap()` for access to the underlying writer.
