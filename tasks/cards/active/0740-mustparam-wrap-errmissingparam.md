# Card 0740

Priority: P2
State: active
Primary Module: contract
Owned Files: contract/context_core.go

Goal:
- Make `MustParam` return an error that wraps `ErrMissingParam` so callers
  can use `errors.Is(err, ErrMissingParam)` to detect missing-parameter
  conditions.

Problem:

`context_core.go:352-358`:
```go
func (c *Ctx) MustParam(key string) (string, error) {
    val, ok := c.Param(key)
    if !ok || val == "" {
        return "", errors.New("missing param: " + key)   // ← inline, unwrappable
    }
    return val, nil
}
```

The package already defines `ErrMissingParam = errors.New("missing parameter")`
(`context_core.go:97`). `MustParam` creates a fresh inline error that shares no
identity with that sentinel. Any caller who writes:

```go
_, err := c.MustParam("id")
if errors.Is(err, contract.ErrMissingParam) { ... }  // never true
```

will never match. This violates the package's own sentinel-error contract and
makes error-type testing in middleware or tests unnecessarily fragile.

Fix:
```go
func (c *Ctx) MustParam(key string) (string, error) {
    val, ok := c.Param(key)
    if !ok || val == "" {
        return "", fmt.Errorf("%w: %s", ErrMissingParam, key)
    }
    return val, nil
}
```

`fmt.Errorf` with `%w` preserves the key name in the error message while
keeping `errors.Is(err, ErrMissingParam)` working.

Non-goals:
- Do not change `ErrMissingParam`'s message.
- Do not change `Param`.
- Do not change `ErrInvalidParam` or unrelated param helpers.

Files:
- `contract/context_core.go`

Tests:
- Add a test in `contract/context_test.go`:
  ```go
  func TestMustParamWrapsErrMissingParam(t *testing.T) {
      c := &Ctx{Params: map[string]string{}}
      _, err := c.MustParam("id")
      if !errors.Is(err, ErrMissingParam) {
          t.Fatalf("expected errors.Is(err, ErrMissingParam) to be true, got: %v", err)
      }
  }
  ```
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `MustParam` returns an error that wraps `ErrMissingParam`.
- `errors.Is(err, ErrMissingParam)` returns `true` for any error returned by
  `MustParam` when a parameter is absent.
- The error message still includes the parameter name.
- All existing tests pass.
