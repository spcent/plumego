# Card 0934

Priority: P1
State: active
Primary Module: contract
Owned Files:
- `contract/validation.go`
- `contract/bind_helpers.go`
Depends On: 0933

Goal:
- Export `validationErrors` as `ValidationErrors` so callers can use standard Go error inspection (`errors.As`) against the type, instead of being forced to use the bespoke `FieldErrorsFrom` accessor.

Problem:
- `validationErrors` (`validation.go:13`) is an unexported type that implements both `error` and the internal `fieldErrorProvider` interface.
- Because it is unexported, callers cannot write:
  ```go
  var verr contract.ValidationErrors
  if errors.As(err, &verr) { ... }
  ```
- The workaround, `FieldErrorsFrom(err)`, works but bypasses standard Go error handling idioms. It also cannot distinguish between "zero errors" and "not a validation error" — both return `nil`.
- Every other significant error type in `contract` is exported (`APIError`, `WrappedErrorWithContext`, `FieldError`). `validationErrors` is the only collection type that hides behind a private name, creating an inconsistency in what callers can inspect.
- Downstream packages that want to act on validation errors (e.g., to format a response in a custom way) must use `FieldErrorsFrom`, which is a naming and pattern mismatch with idiomatic Go error handling.

Scope:
- Rename `validationErrors` → `ValidationErrors` (export it).
- Export the `Errors() []FieldError` method (it is already public in behavior; this just makes it accessible through the exported type).
- Update `FieldErrorsFrom` to use `errors.As(err, &ValidationErrors{})` internally, keeping it as a convenience function.
- Update `validateStructAtDepth` and `validateNestedStructField` to return and construct `ValidationErrors` instead of `validationErrors`.

Non-goals:
- Do not remove `FieldErrorsFrom`; it remains as a convenience wrapper.
- Do not change `FieldError` struct or its JSON tags.
- Do not change `ValidateStruct` signature.

Files:
- `contract/validation.go`
- `contract/bind_helpers.go`
- Any test that constructs `validationErrors` directly (likely zero — it's unexported).

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None required.

Done Definition:
- `ValidationErrors` is an exported type in the `contract` package.
- `errors.As(err, &contract.ValidationErrors{})` works for errors returned by `ValidateStruct`.
- `FieldErrorsFrom` continues to work correctly.
- All tests pass.

Outcome:
- Pending.
