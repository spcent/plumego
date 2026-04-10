# Card 0933

Priority: P1
State: active
Primary Module: contract
Owned Files:
- `contract/validation.go`
Depends On:

Goal:
- Make unknown `validate` tag rules fail loudly instead of silently producing a runtime `FieldError`, so configuration mistakes are caught at development time rather than surfaced as fake validation failures to end users.

Problem:
- `applyValidationRule` (`validation.go:138`) has a default case that returns a `FieldError` for unrecognized rule names:
  ```go
  default:
      return &FieldError{Field: fieldName, Code: "unknown_rule",
          Message: fmt.Sprintf("unknown validation rule: %q", name)}
  ```
- A typo in a `validate` struct tag (e.g., `validate:"reqiured"`) will silently pass through `ValidateStruct` and produce a `FieldError` with code `"unknown_rule"` as if it were a runtime validation failure.
- The end user receives a confusing error message like `"Name: unknown validation rule: \"reqiured\""` in their 400 response, which has nothing to do with their input — it is a developer mistake embedded in the struct definition.
- Because `ValidateStruct` operates on runtime values, there is no static analysis catching this. The failure only surfaces when the handler is exercised, potentially in production.
- Every other `FieldError` in the package describes a genuine data problem. Using the same type for a configuration error conflates fundamentally different kinds of failures.

Scope:
- Change the `default` case in `applyValidationRule` to return a distinct error type (not `*FieldError`) that signals a programming mistake:
  - Concretely: return a `*badTagError` (new unexported type) or use `fmt.Errorf("unknown validation rule %q on field %s", name, fieldName)`.
- Update `validateStructAtDepth` to propagate this error directly (unwrapped from the `validationErrors` collection) so that callers of `ValidateStruct` can distinguish configuration bugs from data validation failures.
- This way `FieldErrorsFrom(err)` returns nothing for a bad-tag error, and the error message is unambiguous.
- Add a test case: `ValidateStruct` on a struct with an unknown rule must return an error that is NOT extractable via `FieldErrorsFrom`.

Non-goals:
- Do not add rule registration or plugin validation systems.
- Do not change the existing rules (`required`, `email`, `min`, `max`).
- Do not panic; return a clear error so handlers can decide how to surface it.

Files:
- `contract/validation.go`
- `contract/errors_test.go` (or equivalent test file)

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None required.

Done Definition:
- An unknown `validate` tag rule produces a non-`FieldError` error from `ValidateStruct`.
- `FieldErrorsFrom` returns nil for this error.
- A test asserts the separation.
- All existing tests pass.

Outcome:
- Pending.
