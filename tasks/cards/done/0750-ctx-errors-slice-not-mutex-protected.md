# Card 0750

Priority: P2
State: active
Primary Module: contract
Owned Files: contract/context_core.go

Goal:
- Guard `Ctx.errors` with the existing `c.mu` mutex so that `Error()` and
  `CollectedErrors()` are safe for concurrent use, matching the contract
  already established by `Set()` and `Get()`.

Problem:

`context_core.go`:
```go
// Set — mutex protected ✓
func (c *Ctx) Set(key string, value any) {
    c.mu.Lock()
    ...
    c.mu.Unlock()
}

// Get — mutex protected ✓
func (c *Ctx) Get(key string) (any, bool) {
    c.mu.RLock()
    ...
    c.mu.RUnlock()
}

// Error — NOT protected ✗
func (c *Ctx) Error(err error) error {
    if err != nil {
        c.errors = append(c.errors, err)   // ← unsynchronised write
    }
    return err
}

// CollectedErrors — NOT protected ✗
func (c *Ctx) CollectedErrors() []error {
    if c == nil || len(c.errors) == 0 {   // ← unsynchronised read
        return nil
    }
    return append([]error(nil), c.errors...)
}
```

`Set` and `Get` are protected by `c.mu` and are explicitly tested for
concurrent use (`TestSetGetConcurrent`). Handlers that pass `ctx` to
goroutines they spawn can safely call `Set`/`Get` from multiple goroutines.

`Error` and `CollectedErrors` access `c.errors` without any lock. If a handler
spawns a goroutine that calls `c.Error(...)` while the handler's own goroutine
calls `c.CollectedErrors()`, there is a **data race** on the `c.errors` slice.
Running `go test -race` on such a test would fail.

Fix — protect with the existing `c.mu`:
```go
func (c *Ctx) Error(err error) error {
    if err != nil {
        c.mu.Lock()
        c.errors = append(c.errors, err)
        c.mu.Unlock()
    }
    return err
}

func (c *Ctx) CollectedErrors() []error {
    if c == nil {
        return nil
    }
    c.mu.RLock()
    defer c.mu.RUnlock()
    if len(c.errors) == 0 {
        return nil
    }
    return append([]error(nil), c.errors...)
}
```

Non-goals:
- Do not change the semantics of `Error` or `CollectedErrors`.
- Do not protect other Ctx fields (W, R, Params, etc.) — those are
  intentionally not concurrent-safe and should not be used concurrently.

Files:
- `contract/context_core.go`

Tests:
- Add a concurrent test mirroring `TestSetGetConcurrent`:
  ```go
  func TestErrorCollectedErrorsConcurrent(t *testing.T) {
      ctx := NewCtx(httptest.NewRecorder(),
          httptest.NewRequest(http.MethodGet, "/", nil), nil)
      sentinel := errors.New("concurrent error")
      done := make(chan struct{})
      const n = 100
      for i := 0; i < n; i++ {
          go func() {
              ctx.Error(sentinel)
              done <- struct{}{}
          }()
      }
      for i := 0; i < n; i++ {
          go func() {
              _ = ctx.CollectedErrors()
              done <- struct{}{}
          }()
      }
      for i := 0; i < 2*n; i++ {
          <-done
      }
      if len(ctx.CollectedErrors()) != n {
          t.Fatalf("expected %d errors, got %d", n, len(ctx.CollectedErrors()))
      }
  }
  ```
- `go test -race ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `Error` and `CollectedErrors` hold `c.mu` when reading or writing `c.errors`.
- `go test -race ./contract/...` passes with no data-race warnings.
- All existing tests pass.

Outcome:
- Completed by guarding `c.errors` with the existing `c.mu` mutex in both
  `Error` and `CollectedErrors`, plus a concurrent regression test that passes
  under the race detector.

Validation Run:
- `gofmt -w contract/context_core.go contract/context_abort_test.go`
- `go test -race -timeout 60s ./contract/...`
- `go vet ./contract/...`
