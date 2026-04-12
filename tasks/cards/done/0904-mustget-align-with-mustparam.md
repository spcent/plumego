# Card 0904

Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/context_core.go`
Depends On: —

Goal:
- Align `MustGet()` with `MustParam()`: return `(any, error)` instead of panicking.

Problem:
Both `MustParam` and `MustGet` are "required value" accessors on `*Ctx`, but they handle
missing values differently:

```go
// MustParam — returns error on missing
func (c *Ctx) MustParam(key string) (string, error) {
    val, ok := c.Param(key)
    if !ok || val == "" {
        return "", fmt.Errorf("%w: %s", ErrMissingParam, key)
    }
    return val, nil
}

// MustGet — panics on missing
func (c *Ctx) MustGet(key string) any {
    val, ok := c.Get(key)
    if !ok {
        panic("contract.Ctx: missing key " + key)
    }
    return val
}
```

This inconsistency forces callers to reason about two different failure modes for
structurally identical operations. A panic in an HTTP handler is particularly harmful:
without recovery middleware every unhandled panic kills the goroutine or worse, crashes
the server. Returning an error lets the caller decide how to respond (log and abort,
write a 500, etc.) without requiring a global panic-recovery net.

Fix:
- Change `MustGet(key string) any` → `MustGet(key string) (any, error)`.
- Return `fmt.Errorf("%w: %s", ErrMissingKey, key)` (add `ErrMissingKey` sentinel alongside `ErrMissingParam`).
- Update the doc comment to remove the panic warning.
- Grep for callers: `grep -rn '\.MustGet(' . --include='*.go'`
- Update all callers to handle the returned error.

Non-goals:
- Do not change `Get()` or `Param()` soft-accessor signatures.
- Do not change `MustParam()`.

Files:
- `contract/context_core.go`
- Any caller files found by grep

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `MustGet` returns `(any, error)` and never panics.
- `ErrMissingKey` sentinel is defined alongside `ErrMissingParam`.
- All callers updated; `go build ./...` passes.
- All tests pass.

Outcome:
- `MustGet(key string) any` changed to `MustGet(key string) (any, error)`.
- Returns `fmt.Errorf("%w: %s", ErrMissingKey, key)` on missing key; never panics.
- `ErrMissingKey` sentinel added alongside `ErrMissingParam` in `context_core.go`.
- Doc comment updated to remove panic warning.
- `go test -timeout 20s ./...` passes.
