# Card 0254

Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/errors.go`
- `contract/bind_helpers.go`
Depends On: —

Goal:
- Add `ErrorBuilder.TypeOnly()` to set only the `Type` field without overwriting `Status`, `Category`, or `Code`; fix `BindErrorToAPIError` to use it so the intent is explicit.

Problem:
`ErrorBuilder.Type()` sets four fields at once — `Type`, `Status`, `Category`, and `Code` —
using the canonical metadata for that `ErrorType`. Its doc explicitly warns:

> "Any Category, Code, or Status set before calling Type will be overwritten."

`BindErrorToAPIError` calls `Type()` first, then immediately overrides all three derived
fields with its own values:

```go
builder := NewErrorBuilder().
    Type(errorType).    // sets Type, Status=400, Category=CategoryValidation, Code=CodeInvalidFormat
    Status(status).     // overwrites Status (sometimes identical, sometimes different)
    Category(category). // overwrites Category (always CategoryValidation — same)
    Code(code).         // overwrites Code (DIFFERENT: CodeRequestBindError or CodeValidationError)
    Message(message)
```

The only lasting effect of `Type()` in this chain is setting `err.Type`. Readers of this
call must understand the full semantics of `Type()`, realize all three side effects are
discarded, and conclude that `Type()` is here purely for the Type field — a hidden surprise.

This pattern will silently produce wrong codes if a future caller copies this pattern
without the override trio (e.g. drops the `Code(code)` call), because `Type()`
side-effects will silently win.

Fix:
Add a `TypeOnly` method to `ErrorBuilder`:

```go
// TypeOnly sets the Type field without changing Status, Category, or Code.
// Use when the caller has already set Status, Category, and Code explicitly
// and only needs to tag the error's type for observability.
func (b *ErrorBuilder) TypeOnly(errorType ErrorType) *ErrorBuilder {
    if errorType != "" {
        b.err.Type = errorType
    }
    return b
}
```

Update `BindErrorToAPIError` to use it:

```go
builder := NewErrorBuilder().
    Status(status).
    Category(category).
    Code(code).
    Message(message).
    TypeOnly(errorType)
```

Non-goals:
- Do not change the semantics of `Type()` — callers who rely on its defaults must not be affected.
- Do not rename `Type()`.

Files:
- `contract/errors.go`
- `contract/bind_helpers.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `ErrorBuilder.TypeOnly()` exists and sets only `err.Type`.
- `BindErrorToAPIError` uses `TypeOnly` instead of `Type` in its builder chain.
- No existing `Type()` callers are changed.
- All tests pass.

Outcome:
- `ErrorBuilder.TypeOnly(errorType ErrorType) *ErrorBuilder` added to `contract/errors.go`.
- `BindErrorToAPIError` in `contract/bind_helpers.go` updated to use `TypeOnly` instead of `Type`.
- `Type()` semantics unchanged; all existing callers unaffected.
- `go test -timeout 20s ./...` passes.
