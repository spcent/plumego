# Card 0745

Priority: P2
State: active
Primary Module: contract
Owned Files: contract/context_bind.go

Goal:
- Wrap `strconv` parse errors returned by `setFieldFromQuery` with
  `ErrInvalidParam` so callers can use `errors.Is(err, ErrInvalidParam)` to
  detect malformed query values.

Problem:

`context_bind.go:234-257` (excerpt):
```go
case reflect.Int, ...:
    n, err := strconv.ParseInt(val, 10, 64)
    if err != nil {
        return err   // ← raw *strconv.NumError, not ErrInvalidParam
    }
...
case reflect.Float32, reflect.Float64:
    n, err := strconv.ParseFloat(val, 64)
    if err != nil {
        return err   // ← same
    }
```

`bindQuery` wraps this error into a `BindError`:
```go
return &BindError{
    Status:  http.StatusBadRequest,
    Message: fmt.Sprintf("invalid query parameter %q: %v", name, err),
    Err:     err,   // ← strconv.NumError / strconv.ErrSyntax
}
```

`ErrInvalidParam` is defined at `context_core.go:99`:
```go
ErrInvalidParam = errors.New("invalid parameter value")
```

Because `err` in `BindError.Err` is a raw `strconv` error, not a wrapped
`ErrInvalidParam`, any caller who writes:
```go
_, err := c.BindQuery(&q)
if errors.Is(err, contract.ErrInvalidParam) { ... }  // never true
```
will never match — even though conceptually the error IS an invalid parameter
value. This is the same sentinel-bypass problem as Card 0740 (`MustParam` /
`ErrMissingParam`).

Fix — wrap parse errors in `setFieldFromQuery`:
```go
case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    n, err := strconv.ParseInt(val, 10, 64)
    if err != nil {
        return fmt.Errorf("%w: %w", ErrInvalidParam, err)
    }
    fv.SetInt(n)
case reflect.Uint, ...:
    n, err := strconv.ParseUint(val, 10, 64)
    if err != nil {
        return fmt.Errorf("%w: %w", ErrInvalidParam, err)
    }
    fv.SetUint(n)
case reflect.Float32, reflect.Float64:
    n, err := strconv.ParseFloat(val, 64)
    if err != nil {
        return fmt.Errorf("%w: %w", ErrInvalidParam, err)
    }
    fv.SetFloat(n)
case reflect.Bool:
    b, err := strconv.ParseBool(val)
    if err != nil {
        return fmt.Errorf("%w: %w", ErrInvalidParam, err)
    }
    fv.SetBool(b)
```

Using `%w: %w` (dual-wrap) preserves both `ErrInvalidParam` and the original
`strconv` error in the chain, so `errors.Is(err, ErrInvalidParam)` returns
`true` and the underlying `strconv.ErrSyntax` / `strconv.ErrRange` is still
reachable via `errors.As` for detailed diagnostics.

The `BindError.Err` field will now wrap `ErrInvalidParam` (which itself wraps
the strconv error), so `errors.Is(bindErr, ErrInvalidParam)` also returns
`true` through `BindError.Unwrap()`.

Non-goals:
- Do not change `ErrInvalidParam`'s message.
- Do not change `BindError` or `bindQuery`'s outer BindError wrapping.
- Do not add sentinel wrapping for the "non-nil pointer to struct" input
  validation checks at the top of `bindQuery` (those use a different sentinel,
  `CodeInvalidBindDst`, and are already descriptive).

Files:
- `contract/context_bind.go`

Tests:
- Add tests in `contract/context_test.go` or `contract/context_extended_test.go`:
  ```go
  func TestBindQueryWrapsErrInvalidParam(t *testing.T) {
      type Q struct {
          Age int `query:"age"`
      }
      req := httptest.NewRequest("GET", "/?age=notanumber", nil)
      c := NewCtx(httptest.NewRecorder(), req, nil)
      var q Q
      err := c.BindQuery(&q)
      if !errors.Is(err, ErrInvalidParam) {
          t.Fatalf("expected errors.Is(err, ErrInvalidParam) = true, got: %v", err)
      }
  }
  ```
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `setFieldFromQuery` wraps all `strconv` parse errors with `ErrInvalidParam`.
- `errors.Is(err, ErrInvalidParam)` returns `true` for malformed query
  parameter values.
- The original `strconv` error is still reachable via `errors.As`.
- All existing tests pass.
