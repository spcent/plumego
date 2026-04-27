# Card 0098

Priority: P2

Goal:
- Fix `WrapErrorf` so it produces a fully populated `WrappedErrorWithContext`,
  or change it so it does not pretend to be a context-aware wrapper.

Problem:

`error_wrap.go:98-108`:
```go
func WrapErrorf(err error, format string, args ...any) error {
    if err == nil {
        return nil
    }
    return &WrappedErrorWithContext{
        Err:     err,
        Message: fmt.Sprintf(format, args...),
        When:    time.Now(),
        // Context is zero-value: no Operation, Module, or Params
    }
}
```

`WrapErrorf` returns a `*WrappedErrorWithContext` with an empty `Context` field.
When this error passes through `GetErrorDetails`, `logErrorWithContext`, or
`FormatError`, the operation and module appear as empty strings or nil in logs:

```
// logErrorWithContext result for a WrapErrorf error:
fields = {"operation": nil, "module": nil}
```

This is structurally inconsistent with `WrapError`, which always sets
`Operation`, `Module`, and `Params`. A caller choosing `WrapErrorf` for its
concise formatted-message syntax gets a less informative error that behaves
differently in the logging pipeline.

Two valid fixes:

**Option A: Add operation and module parameters**
```go
func WrapErrorf(err error, operation, module, format string, args ...any) error
```
Adds context, at the cost of a more verbose signature.

**Option B: Replace with fmt.Errorf**
Remove `WrapErrorf` entirely; its only added value over `fmt.Errorf("%w: ...", err)`
is the `When` timestamp and the `WrappedErrorWithContext` type assertion. If callers
don't need the type, `fmt.Errorf` is cleaner and standard.

**Option C: Keep as-is but document the limitation explicitly**
Add a prominent doc comment: "WrapErrorf creates a message-only wrapper with no
operation context. Use WrapError when operation metadata is needed for logging."

The lack of context fields is only invisible because downstream code gracefully
handles empty strings. It surfaces as log entries missing `operation` and `module`.

Scope:
- Choose one of the three options.
- If A: update all callers of `WrapErrorf`.
- If B: grep callers, migrate each, then remove the function.
- If C: update the doc comment only.

Files:
- `contract/error_wrap.go`
- (If A or B) all callers: `grep -rn 'WrapErrorf' . --include='*.go'`

Tests:
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `WrapErrorf` behavior matches one of A, B, or C.
- The chosen behavior is documented in the function's doc comment.
- All tests pass.

Outcome:
- Completed in the 2026-04-05 contract cleanup batch.
- Verified as part of the shared contract/task-card completion pass.

Validation Run:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
- `go build ./...`
- `go test -timeout 20s ./...`
- `go test -race -timeout 60s ./...`
- `go vet ./...`
