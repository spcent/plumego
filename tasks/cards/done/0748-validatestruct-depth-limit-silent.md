# Card 0748

Priority: P2
State: active
Primary Module: contract
Owned Files: contract/validation.go

Goal:
- Replace the silent `return nil` when the recursion depth exceeds 10 with a
  `FieldError` that surfaces the truncation so callers are not misled into
  thinking deeply nested structs were fully validated.

Problem:

`validation.go:40-42`:
```go
func validateStructAtDepth(dst any, prefix string, depth int) error {
    ...
    if depth > 10 {
        return nil   // ← silent success — no FieldError, no log, nothing
    }
```

When struct validation reaches a nesting depth greater than 10, the function
returns `nil` as if there were no validation errors. Any `validate:"required"`
or other struct tags on fields deeper than 10 levels are never checked.

This creates a silent **false-negative**: a caller who has a struct with tags
at depth 11 will receive a clean validation result even though those fields
were never inspected. There is no FieldError, no log message, and no indication
in the response that validation was truncated.

Example:
```go
type L11 struct { Value string `validate:"required"` }
type L10 struct { Child L11 }
... (chain of 10 wrappers) ...
type Root struct { Child L1 }

errs := validateStruct(&Root{})
// errs == nil — L11.Value was not checked
```

Fix — return a FieldError when the depth limit is hit:
```go
if depth > 10 {
    fieldName := prefix
    if fieldName == "" {
        fieldName = "(root)"
    }
    return validationErrors{errors: []FieldError{{
        Field:   fieldName,
        Code:    "max_depth_exceeded",
        Message: "struct nesting exceeds maximum validation depth (10); deeper fields were not validated",
    }}}
}
```

This turns the silent truncation into an observable validation failure. The
`400` response will include a `FieldError` with `code: "max_depth_exceeded"`,
making it immediately visible during development without requiring a debugger
or log search. Any struct that triggers this is almost certainly a design
problem (10 levels of nesting is extreme), so surfacing it as an error is
more useful than silently passing.

Non-goals:
- Do not increase or make the depth limit configurable in this card.
- Do not change `validateNestedStructField`; the depth check in
  `validateStructAtDepth` is sufficient.
- Do not change any other aspect of the validation logic.

Files:
- `contract/validation.go`

Tests:
- Add a test that builds an 11-level nested struct and verifies a FieldError
  with `Code: "max_depth_exceeded"` is returned:
  ```go
  func TestValidateStructDepthLimitReturnsFieldError(t *testing.T) {
      type L struct{ V string `validate:"required"` }
      // wrap 11 times
      type W10 struct{ C L }
      type W9  struct{ C W10 }
      type W8  struct{ C W9 }
      type W7  struct{ C W8 }
      type W6  struct{ C W7 }
      type W5  struct{ C W6 }
      type W4  struct{ C W5 }
      type W3  struct{ C W4 }
      type W2  struct{ C W3 }
      type W1  struct{ C W2 }
      type Root struct{ C W1 }

      err := validateStruct(&Root{})
      if err == nil {
          t.Fatal("expected depth-limit error, got nil")
      }
      fields := FieldErrorsFrom(err)
      var found bool
      for _, fe := range fields {
          if fe.Code == "max_depth_exceeded" {
              found = true
          }
      }
      if !found {
          t.Fatalf("expected FieldError with code 'max_depth_exceeded', got: %v", fields)
      }
  }
  ```
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `validateStructAtDepth` returns a `FieldError` with `Code: "max_depth_exceeded"`
  when the depth limit is hit, instead of returning nil.
- The new test passes.
- All existing validation tests pass unchanged.

Outcome:
- Completed by converting the silent depth-limit success path into a
  `validationErrors` return with `Code: "max_depth_exceeded"` and adding a
  focused regression test for an 11-level nested struct.

Validation Run:
- `gofmt -w contract/validation.go contract/active_cards_regression_test.go`
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
