# Card 0504

Priority: P0

Goal:
- Eliminate the duplicate validation-error struct (`validationIssue` in
  `validation.go` vs `FieldError` in `bind_helpers.go`) and use a single
  exported type throughout the bind and validation pipeline.

Problem:
- Two nearly identical structs exist with different visibility and JSON tags:

  `validation.go` (unexported):
  ```go
  type validationIssue struct {
      Field   string
      Code    string
      Message string
  }
  ```

  `bind_helpers.go` (exported):
  ```go
  type FieldError struct {
      Field   string `json:"field"`
      Code    string `json:"code,omitempty"`
      Message string `json:"message"`
  }
  ```

- `extractFieldErrors()` bridges them via reflection, adding fragility and
  indirection.
- `FieldErrorsFrom(err)` returns `[]FieldError` while the internal validator
  produces `[]validationIssue`, requiring an intermediate conversion step.

Scope:
- Make `FieldError` the single canonical type.
- Rewrite `validationErrors` / `validationIssue` internals to use `FieldError`
  directly.
- Remove `extractFieldErrors()` reflection bridge.
- Ensure `FieldErrorsFrom(err error) []FieldError` works without conversion.

Non-goals:
- Do not change the JSON wire format of `FieldError`.
- Do not change how `BindError` surfaces these to callers.

Files:
- `contract/validation.go`
- `contract/bind_helpers.go`
- `contract/context_bind.go`

Tests:
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `validationIssue` struct is removed.
- `extractFieldErrors()` reflection function is removed.
- All internal validator code produces `FieldError` directly.
- `FieldErrorsFrom` requires no conversion step.
- All tests pass.
