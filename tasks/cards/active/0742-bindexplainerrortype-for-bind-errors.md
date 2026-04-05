# Card 0742

Priority: P2
State: active
Primary Module: contract
Owned Files: contract/bind_helpers.go

Goal:
- Fix `BindErrorToAPIError` so each bind-error variant is assigned a
  semantically correct `ErrorType`, instead of every variant inheriting the
  `ErrTypeValidation` default.

Problem:

`bind_helpers.go:44-97`:
```go
func BindErrorToAPIError(err error) APIError {
    ...
    errorType := ErrTypeValidation   // ← never updated in the switch

    switch {
    case errors.Is(err, ErrRequestBodyTooLarge):
        status = http.StatusRequestEntityTooLarge
        code = CodeRequestBodyTooLarge
        message = ErrRequestBodyTooLarge.Error()
        // errorType stays ErrTypeValidation — wrong
    case errors.Is(err, ErrEmptyRequestBody):
        code = CodeEmptyBody
        message = ErrEmptyRequestBody.Error()
        // errorType stays ErrTypeValidation — wrong
    case errors.Is(err, ErrInvalidJSON):
        code = CodeInvalidJSON
        message = ErrInvalidJSON.Error()
        // errorType stays ErrTypeValidation — arguable, but inconsistent with code
    case errors.Is(err, ErrUnexpectedExtraData):
        code = CodeUnexpectedExtraData
        message = ErrUnexpectedExtraData.Error()
        // errorType stays ErrTypeValidation — arguable
    ...
    }
    builder := NewErrorBuilder().Type(errorType)...
```

`ErrTypeValidation` signals a failed domain invariant (field rules, required
checks, format constraints). A body-too-large or empty-body error is an
**input** error but not a struct-tag validation failure. Labelling it
`ErrTypeValidation` misleads observability consumers, metrics dashboards, and
callers who gate retry logic on `ErrorType`.

The correct types for each case:

| Sentinel               | Current Type         | Correct Type            |
|------------------------|----------------------|-------------------------|
| `ErrRequestBodyTooLarge` | `ErrTypeValidation` | `ErrTypeInvalidFormat`  |
| `ErrEmptyRequestBody`  | `ErrTypeValidation`  | `ErrTypeInvalidFormat`  |
| `ErrInvalidJSON`       | `ErrTypeValidation`  | `ErrTypeInvalidFormat`  |
| `ErrUnexpectedExtraData` | `ErrTypeValidation` | `ErrTypeInvalidFormat`  |
| validation `FieldError` present | `ErrTypeValidation` | `ErrTypeValidation` ✓ |
| generic `BindError` fallback | `ErrTypeValidation` | `ErrTypeInvalidFormat`  |

`ErrTypeInvalidFormat` already exists (`errors.go:54`) and maps to
`CodeInvalidFormat` / `CategoryValidation` / 400. It correctly describes
"the input did not conform to the expected structure or size" without implying
a domain-rule failure.

Fix — update `errorType` assignment inside each case:
```go
errorType := ErrTypeInvalidFormat   // default for all structural/transport errors

switch {
case errors.Is(err, ErrRequestBodyTooLarge):
    status = http.StatusRequestEntityTooLarge
    code = CodeRequestBodyTooLarge
    message = ErrRequestBodyTooLarge.Error()
    // errorType stays ErrTypeInvalidFormat ✓
case errors.Is(err, ErrEmptyRequestBody):
    code = CodeEmptyBody
    message = ErrEmptyRequestBody.Error()
case errors.Is(err, ErrInvalidJSON):
    code = CodeInvalidJSON
    message = ErrInvalidJSON.Error()
case errors.Is(err, ErrUnexpectedExtraData):
    code = CodeUnexpectedExtraData
    message = ErrUnexpectedExtraData.Error()
default:
    ...
}

if len(fields) > 0 {
    errorType = ErrTypeValidation   // field-level validation → correct type
    code = CodeValidationError
    message = "validation failed"
}
```

Non-goals:
- Do not add new `ErrorType` constants.
- Do not change any HTTP status codes.
- Do not change `FieldError`, `BindOptions`, or `WriteBindError`.

Files:
- `contract/bind_helpers.go`

Tests:
- Extend `contract/bind_helpers_test.go` (or `contract/write_bind_error_test.go`)
  with type-checking assertions:
  ```go
  func TestBindErrorToAPIErrorType(t *testing.T) {
      cases := []struct {
          err      error
          wantType ErrorType
      }{
          {ErrRequestBodyTooLarge, ErrTypeInvalidFormat},
          {ErrEmptyRequestBody,    ErrTypeInvalidFormat},
          {ErrInvalidJSON,         ErrTypeInvalidFormat},
          {ErrUnexpectedExtraData, ErrTypeInvalidFormat},
      }
      for _, tc := range cases {
          got := BindErrorToAPIError(tc.err)
          if got.Type != tc.wantType {
              t.Errorf("BindErrorToAPIError(%v).Type = %v, want %v",
                  tc.err, got.Type, tc.wantType)
          }
      }
  }
  ```
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `BindErrorToAPIError` assigns `ErrTypeInvalidFormat` to all structural/
  transport bind errors.
- `BindErrorToAPIError` assigns `ErrTypeValidation` only when `FieldError`
  items are present.
- New tests pass; all existing tests pass.
