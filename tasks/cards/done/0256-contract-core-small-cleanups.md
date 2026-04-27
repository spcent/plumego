# Card 0256

Priority: P3
State: done
Primary Module: contract, core
Owned Files:
- `contract/context_stream.go`
- `contract/context_core.go`
- `core/lifecycle.go`
Depends On: —

Goal:
- Fix four small but concrete issues found across the stream, abort, and lifecycle surfaces;
  bundle them in one pass since each change is a handful of lines.

---

## Issue A — `io.EOF` comparison in `streamFromGen` and `streamFromGenWithRetry`

**File:** `contract/context_stream.go` lines 422, 460

```go
// current
if err == io.EOF { ... }
```

`io.EOF` is typically never wrapped, so this works today — but `errors.Is` is the
idiomatic Go way to check sentinel errors and is robust against future wrapping.
Both loop functions use the bare comparison; switching to `errors.Is` makes the
intent explicit and future-proof.

**Fix:** Replace with `errors.Is(err, io.EOF)` in both functions.

---

## Issue B — `initSSEStream` skips the context cancellation check that `initStream` performs

**File:** `contract/context_stream.go` lines 366–382

`initStream` checks `ctx.Err()` before writing response headers, returning early if the
request is already cancelled. `initSSEStream` writes three headers and a status code
unconditionally, relying on the call site to check the context beforehand:

```go
// initStream — guards with ctx.Err()
func (c *Ctx) initStream(contentType string) (context.Context, error) {
    ctx := c.streamContext()
    if err := ctx.Err(); err != nil { return nil, err }
    c.W.Header().Set(...)
    c.W.WriteHeader(http.StatusOK)
    return ctx, nil
}

// initSSEStream — no guard
func (c *Ctx) initSSEStream() (*SSEWriter, error) {
    c.W.Header().Set("Content-Type", "text/event-stream")
    c.W.Header().Set("Cache-Control", "no-cache")
    c.W.Header().Set("Connection", "keep-alive")
    c.W.WriteHeader(http.StatusOK)
    return NewSSEWriter(c.W)
}
```

This inconsistency requires every SSE call site to duplicate the `ctx.Err()` check
(currently done in `streamSSESlice`, `streamSSESliceChunked`, and the two SSE channel/
generator branches in `Stream()`). If a new SSE path is added without the guard, it will
write response headers to a cancelled request.

**Fix:** Add the context check inside `initSSEStream`; have it return `(context.Context, *SSEWriter, error)` to match the `initStream` return shape. Update all call sites to remove their duplicate guards and use the returned context.

---

## Issue C — `AbortWithStatus` has a window between `WriteHeader` and the atomic abort

**File:** `contract/context_core.go` lines 283–286

```go
func (c *Ctx) AbortWithStatus(code int) {
    c.W.WriteHeader(code)  // ← header written before abort flag is set
    c.Abort()
}
```

There is a narrow race window: between `WriteHeader` and the CAS in `Abort()`, another
goroutine calling `IsAborted()` would see `aborted == false` even though a response
status has already been committed. More concretely, if `AbortWithStatus` is called while
`IsAborted()` is checked concurrently, the caller may proceed with further response
writes between the header write and the abort flag being set.

The fix aligns `AbortWithStatus` with `Abort`'s own CAS pattern so both the header write
and the flag set happen atomically from the caller's perspective:

```go
func (c *Ctx) AbortWithStatus(code int) {
    if c.aborted.CompareAndSwap(false, true) {
        c.W.WriteHeader(code)
        if c.cancel != nil {
            c.cancel()
        }
    }
}
```

This also prevents a double `WriteHeader` if `AbortWithStatus` is called twice.

---

## Issue D — `Prepare()` adds a redundant if/return wrapper

**File:** `core/lifecycle.go` lines 15–20

```go
func (a *App) Prepare() error {
    if err := a.ensureServerPrepared(); err != nil {
        return err
    }
    return nil
}
```

This is identical to `return a.ensureServerPrepared()`. The wrapper adds no
documentation value, no pre/post logic, and no error transformation.

**Fix:** Collapse to a one-liner:

```go
func (a *App) Prepare() error {
    return a.ensureServerPrepared()
}
```

---

Scope:
- Apply all four fixes in a single pass.
- Each change is isolated to a few lines; no cross-file dependencies between the fixes.

Non-goals:
- Do not change streaming behavior, abort semantics, or lifecycle contract.
- Do not refactor `initStream` / `initSSEStream` beyond what Issue B requires.

Files:
- `contract/context_stream.go`
- `contract/context_core.go`
- `core/lifecycle.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/... ./core/...`
- `go test -race -timeout 60s ./contract/... ./core/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `errors.Is(err, io.EOF)` used in both generator loop functions.
- `initSSEStream` performs the context check internally; call-site guards removed.
- `AbortWithStatus` uses a single CAS block for both header write and cancel.
- `Prepare()` is a one-liner delegating to `ensureServerPrepared`.
- All tests pass, including race detector.

Outcome:
- A) `errors.Is(err, io.EOF)` used in both `streamFromGen` and `streamFromGenWithRetry`.
- B) `initSSEStream` now returns `(context.Context, *SSEWriter, error)` with internal context cancellation guard; all 4 SSE call sites updated to remove duplicate guards.
- C) `AbortWithStatus` uses a single CAS block: `CompareAndSwap(false, true)` guards both `WriteHeader` and `cancel()`; double-abort is now a no-op.
- D) `Prepare()` in `core/lifecycle.go` collapsed to `return a.ensureServerPrepared()`.
- `go test -timeout 20s ./...` passes.
