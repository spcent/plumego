# recovery — Panic Recovery Middleware

Protects the application from crashing when a request handler panics.

## Import

```go
import "github.com/spcent/plumego/middleware/recovery"
```

## Overview

`recovery.Middleware` catches panics in downstream handlers, logs sanitized metadata (without leaking internal details), and returns a generic `500 Internal Server Error` to the client.

## Exports

| Symbol | Description |
|--------|-------------|
| `Middleware(config Config)` | Constructor; returns a `middleware.Middleware` |
| `Config{Logger}` | Required configuration with a `log.StructuredLogger` |
| `ErrNilLogger` | Returned when `Config.Logger` is nil |

## Usage

```go
mw, err := recovery.Middleware(recovery.Config{Logger: logger})
if err != nil {
    return err
}
app.Use(mw)
```

## Notes

- Place `recovery` early in the middleware chain (before handlers that may panic).
- Panic details are logged server-side but never returned in the HTTP response.
- `recovery` does not recover panics inside goroutines launched by handlers.
