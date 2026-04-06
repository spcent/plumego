# Card 0810

Milestone: contract cleanup
Priority: P3
State: done
Primary Module: contract
Owned Files:
- `contract/error_wrap.go`
Depends On: —

Goal:
- Remove the misleading `json:"error"` tag from `WrappedErrorWithContext.Err`
  so that serialization of the struct no longer emits `"error": null`.

Problem:
`WrappedErrorWithContext` carries JSON tags on all fields:

```go
type WrappedErrorWithContext struct {
    Err     error        `json:"error"`     // ← always null in JSON
    Context ErrorContext `json:"context"`
    Message string       `json:"message"`
    When    time.Time    `json:"when"`
}
```

The `Err error` field is of interface type. The standard `encoding/json` package
marshals interface values as `null` unless a custom marshaler is implemented. Any
code that serializes a `*WrappedErrorWithContext` (e.g., a structured logger that
marshals the error chain) always receives:

```json
{"error": null, "context": {...}, "message": "...", "when": "..."}
```

The `"error": null` is actively misleading — a reader expects to see the error
content but gets nothing. The actual error message and type are in `"message"` and
via the `Unwrap` chain.

`WrappedErrorWithContext` is never HTTP-serialized directly; `GetErrorDetails`
extracts its fields manually. The JSON tag on `Err` serves no production use and
only creates confusion.

Fix:
- Change `Err error \`json:"error"\`` → `Err error \`json:"-"\``
- This prevents the useless `"error": null` field from appearing in any
  serialized output.
- No behaviour change to `Error()`, `Unwrap()`, `WrapError`, or `GetErrorDetails`.

Non-goals:
- No custom `MarshalJSON` implementation.
- No change to how `GetErrorDetails` works.
- No removal of JSON tags from other fields.

Files:
- `contract/error_wrap.go`

Tests:
- `go test -timeout 20s ./contract/...`
- Optionally add a test: `json.Marshal(&WrappedErrorWithContext{...})` must not
  contain an `"error"` key in the output.

Docs Sync: —

Done Definition:
- `Err` field has `json:"-"` tag.
- Marshaling `*WrappedErrorWithContext` does not emit an `"error"` key.
- All tests pass.

Outcome:
