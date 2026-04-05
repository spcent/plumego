# Card 0733

Priority: P2

Goal:
- Fix `logBindError` so that struct fields in the logged payload are redacted
  by the observability policy, not just top-level map keys.

Problem:

`context_bind.go:317-319`:
```go
if v := FieldErrorsFrom(err); len(v) > 0 && payload != nil {
    fields["payload"] = DefaultObservabilityPolicy.RedactFields(
        map[string]any{"payload": payload})["payload"]
```

`RedactFields` receives `map[string]any{"payload": <struct>}`. It checks
whether the key `"payload"` is sensitive (it is not), so it calls
`redactValue(<struct>)`. `redactValue` only handles `map[string]any` and
`[]any`; for any other type it returns the value unchanged:

```go
func (p ObservabilityPolicy) redactValue(v any) any {
    switch value := v.(type) {
    case map[string]any:   return p.RedactFields(value)  // ← redacted
    case []any:            // ...
    default:               return value                  // ← struct passed as-is
    }
}
```

The result: when a validation error occurs, the entire request payload struct
is logged as-is, including any fields the struct contains — even fields named
`Password`, `Token`, `Secret`, or `ApiKey`.

```go
type LoginRequest struct {
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required"` // ← logged in plaintext
}
```

A validation failure on `LoginRequest` causes `logBindError` to log the
password in the warn log.

Fix: Before logging, marshal the struct to `map[string]any` so that
`RedactFields` can walk its keys:

```go
if v := FieldErrorsFrom(err); len(v) > 0 && payload != nil {
    if m, err := structToMap(payload); err == nil {
        fields["payload"] = DefaultObservabilityPolicy.RedactFields(m)
    }
    fields["validation_fields"] = v
}
```

Where `structToMap` can be implemented via `json.Marshal` + `json.Unmarshal`
into a `map[string]any`, giving field names as they appear in JSON tags.

Non-goals:
- Do not change `RedactFields` behavior for non-struct types.
- Do not log payloads for non-validation bind errors (that gate already exists).

Files:
- `contract/context_bind.go`

Tests:
- Add a test: validation failure on a struct with a `password` field must NOT
  include the password value in the logged payload fields.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- Struct fields matching sensitive key patterns are masked in `logBindError` output.
- Validation field names in `validation_fields` are still logged (they are field
  names, not values, so they are safe to log).
- All tests pass.
