# Card 0703

Priority: P2

Goal:
- Add `Type ErrorType` and `Severity ErrorSeverity` as first-class fields on
  `APIError` instead of storing them as untyped values inside the `Details` map.

Problem:

`errors.go:285-294`:
```go
func (b *ErrorBuilder) Severity(severity ErrorSeverity) *ErrorBuilder {
    b.err.Details["severity"] = severity   // ← stored as any
    return b
}

func (b *ErrorBuilder) Type(errorType ErrorType) *ErrorBuilder {
    b.err.Details["type"] = errorType      // ← stored as any
    return b
}
```

`ErrorType` is a central concept: `errorTypeLookup` maps every `ErrorType` to
its canonical `Category`, `Code`, and `Status`. `ErrorSeverity` is the
observability severity level. Both are documented, typed, and enumerated. Yet
neither appears as a typed field on `APIError`; they are buried under string
keys in an untyped map.

Consequences:
- A consumer receiving an `APIError` cannot access `Type` or `Severity` without
  a type assertion on `err.Details["type"]` and knowing the key names.
- `ValidateError` does not check these fields.
- `ErrorLogger` does not log them as structured fields.
- The `ErrorBuilder.Type()` method also sets `Category`, `Code`, and `Status`
  via the lookup table but has no way to record which type was used.

Fix:
```go
type APIError struct {
    Status   int            `json:"status,omitempty"`   // note: remove json:"-"
    Code     string         `json:"code"`
    Message  string         `json:"message"`
    Category ErrorCategory  `json:"category"`
    Type     ErrorType      `json:"type,omitempty"`
    Severity ErrorSeverity  `json:"severity,omitempty"`
    TraceID  string         `json:"trace_id,omitempty"`
    Details  map[string]any `json:"details,omitempty"`
}
```

Update `ErrorBuilder.Type()` and `ErrorBuilder.Severity()` to set the struct
fields instead of the Details map. Remove the `details["type"]` and
`details["severity"]` entries. Update `ValidateError` and `ErrorLogger` to
include the new fields.

Note: `APIError.Status` has `json:"-"` today, which means it is not serialized
in error responses. This is a related inconsistency — `Status` is set but never
included in the `{"error": ...}` JSON envelope, forcing clients to rely on the
HTTP status line alone. That inconsistency is tracked separately; this card
should not change the `Status` json tag.

Non-goals:
- Do not change `APIError.Status json:"-"` (separate concern).
- Do not add new `ErrorType` or `ErrorSeverity` values.
- Do not change `errorTypeLookup`.

Files:
- `contract/errors.go`
- All callers that read `err.Details["type"]` or `err.Details["severity"]`
  (run `grep -rn '"type"\|"severity"' contract/ --include='*.go'` first)

Tests:
- Add tests verifying that `Type` and `Severity` are set as struct fields and
  appear in JSON output.
- Verify `Details` no longer contains "type" or "severity" keys.
- `go test ./contract/...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `APIError` has `Type ErrorType` and `Severity ErrorSeverity` fields.
- `ErrorBuilder.Type()` and `ErrorBuilder.Severity()` set struct fields.
- `Details` map no longer receives "type" or "severity" entries from the builder.
- All tests pass.
