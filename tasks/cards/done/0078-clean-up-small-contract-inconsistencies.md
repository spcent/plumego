# Card 0078

Priority: P2

Goal:
- Remove three small inconsistencies in `contract` that were missed by the
  0501-0512 cleanup wave: a dead package-level Param function, an unused error
  sentinel, and a silent validation bypass in WithTraceIDString.

Problems:

### P2-A: Package-level Param duplicates Ctx.Param

`context_core.go:206` exports:
```go
func Param(r *http.Request, key string) (string, bool)
```
This is identical in behavior to `(c *Ctx).Param(key string) (string, bool)`
(context_core.go:331). Two equivalent accessors exist with no documented reason
to prefer one.

There is also `(c *Ctx).MustParam` (context_core.go:339) but no package-level
`MustParam`, breaking the symmetry that does exist.

Fix: Deprecate the package-level `Param(r, key)` function with a doc comment
pointing to `Ctx.Param`. Do not remove in this card (callers must be audited
first).

### P2-B: ErrSSEInvalidField is dead exported code

`context_stream.go:15` exports:
```go
var ErrSSEInvalidField = errors.New("SSE event field contains invalid
characters (newlines not allowed in id/event)")
```
No function returns this error. `sanitizeSSEField` (context_stream.go:245)
silently strips newlines instead of returning `ErrSSEInvalidField`. The
sentinel is dead code and misleads callers who check `errors.Is(err,
contract.ErrSSEInvalidField)`.

Fix: Remove `ErrSSEInvalidField`. If the silent-sanitize behavior should be
replaced with an error return, that is a separate card; for now removing the
dead sentinel is sufficient.

### P2-C: WithTraceIDString skips format validation

`trace.go:74-81` stores a raw string as a `TraceID` without calling
`ParseTraceID`:
```go
func WithTraceIDString(ctx context.Context, id string) context.Context {
    ...
    updated.TraceID = TraceID(id)   // ← no length/hex validation
    ...
}
```
`ParseTraceID` validates 32-hex-char format. An invalid string silently
becomes a `TraceID`, surfacing as a malformed trace ID in logs and headers.

Fix: Call `ParseTraceID` inside `WithTraceIDString`. If parsing fails, store
the raw string with a log warning (match the WarnFunc pattern used in
WriteError), rather than panicking or returning an error (callers do not
currently handle an error return).

Scope:
- Add `// Deprecated: Use Ctx.Param instead.` to package-level `Param`.
- Remove `var ErrSSEInvalidField`.
- Update `WithTraceIDString` to call `ParseTraceID`; on failure call `WarnFunc`
  and store the raw string unchanged (preserve behaviour, surface the problem).

Non-goals:
- Do not change SSE sanitization behavior (that is a larger change).
- Do not add a return error to `WithTraceIDString` (breaking change).
- Do not remove the package-level `Param` function in this card.

Files:
- `contract/context_core.go`
- `contract/context_stream.go`
- `contract/trace.go`

Tests:
- `go build ./...`
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- Package-level `Param` has a `// Deprecated:` first-line doc comment.
- `ErrSSEInvalidField` does not exist.
- `WithTraceIDString` calls `ParseTraceID` and logs a warning for invalid IDs.
- All tests pass.
