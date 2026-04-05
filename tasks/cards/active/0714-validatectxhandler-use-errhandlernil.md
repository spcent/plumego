# Card 0714

Priority: P3

Goal:
- Fix `ValidateCtxHandler` to return the existing `ErrHandlerNil` sentinel
  instead of a new ad-hoc error value, so callers can use `errors.Is`.

Problem:

`context_core.go:426-430`:
```go
func ValidateCtxHandler(h CtxHandlerFunc) error {
    if h == nil {
        return errors.New("context handler cannot be nil")  // ← new error each call
    }
    return nil
}
```

The package already declares (context_core.go:108-109):
```go
ErrHandlerNil = errors.New("handler cannot be nil")
```

`ValidateCtxHandler` ignores this sentinel and creates a fresh error value with
a slightly different message ("context handler cannot be nil" vs "handler cannot
be nil"). This means:

1. `errors.Is(err, contract.ErrHandlerNil)` is always `false` for errors from
   `ValidateCtxHandler` — the sentinel is useless as a guard.
2. Two different messages exist for the same condition, creating confusion when
   the error appears in logs.

Fix:
```go
func ValidateCtxHandler(h CtxHandlerFunc) error {
    if h == nil {
        return ErrHandlerNil
    }
    return nil
}
```

Non-goals:
- Do not change `ErrHandlerNil`'s message.
- Do not add a separate sentinel for "context handler nil" vs "handler nil".

Files:
- `contract/context_core.go`

Tests:
- Add a test: `errors.Is(ValidateCtxHandler(nil), ErrHandlerNil)` must be `true`.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `ValidateCtxHandler(nil)` returns `ErrHandlerNil`.
- `errors.Is(ValidateCtxHandler(nil), ErrHandlerNil)` is `true`.
- All tests pass.
