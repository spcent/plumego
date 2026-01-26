# Lifecycle Contract

This document defines stable lifecycle semantics for background tasks.

## Runner Interface
```go
type Runner interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
```

## Registration
- Register via `app.Register(runner)` or `core.WithRunner`.
- Registration must happen before `Boot()`.

## Startup Order
1. Load environment configuration.
2. Mount components (routes + middleware).
3. Setup HTTP server.
4. Start components.
5. Start runners.
6. Start HTTP server.

Runners are always started before the HTTP server begins serving requests.

## Shutdown Order
1. Stop HTTP server (graceful shutdown).
2. Stop runners (reverse registration order).
3. Stop components (reverse dependency order).

## Context Semantics
- `Start` and `Stop` must respect the provided context deadline.
- Implementations should return promptly on `ctx.Done()`.
