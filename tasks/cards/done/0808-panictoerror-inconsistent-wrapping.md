# Card 0808

Milestone: contract cleanup
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/error_wrap.go`
- `contract/error_handling_test.go`
Depends On: —

Goal:
- Make `PanicToError` return a `*WrappedErrorWithContext` for all panic value types,
  not only for `error`-typed panics.

Problem:
`PanicToError` in `error_wrap.go:224` returns different types depending on the
panic value's type:

```go
switch v := r.(type) {
case error:
    return WrapError(v, "panic_recovery", "recovery", nil)  // *WrappedErrorWithContext
case string:
    return fmt.Errorf("panic: %s", v)                       // *fmt.wrapError
default:
    return fmt.Errorf("panic: %v", v)                       // *fmt.wrapError
}
```

Callers who inspect the result with `GetErrorDetails`, `FormatError`, or
`errors.Unwrap` get different structured output depending on whether the panic
was an `error` value or a string:

- `GetErrorDetails(PanicToError(myErr))` → includes `"operation"`, `"module"`,
  structured context
- `GetErrorDetails(PanicToError("oops"))` → only `"message"` and `"type"` fields,
  no context

The middleware/recovery stack (or any caller) cannot assume uniform treatment
of recovered panics. The fix is to convert `string` and `default` panics to
a plain `error` first, then call `WrapError` so all three branches produce a
`*WrappedErrorWithContext` with identical structure.

Scope:
- Change the `string` branch:
  ```go
  case string:
      return WrapError(errors.New("panic: "+v), "panic_recovery", "recovery", nil)
  ```
- Change the `default` branch:
  ```go
  default:
      return WrapError(fmt.Errorf("panic: %v", v), "panic_recovery", "recovery", nil)
  ```
- Update `TestPanicToError` to assert that string and `any`-type panics also
  produce a `*WrappedErrorWithContext` with the `"panic_recovery"` operation.

Non-goals:
- No change to the `error` branch.
- No change to callers in `middleware/recovery` (the type change is compatible
  since all branches still satisfy the `error` interface).

Files:
- `contract/error_wrap.go`
- `contract/error_handling_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- All three branches of `PanicToError` return `*WrappedErrorWithContext`.
- `GetErrorDetails(PanicToError("string panic"))["operation"]` equals
  `"panic_recovery"`.
- All tests pass.

Outcome:
