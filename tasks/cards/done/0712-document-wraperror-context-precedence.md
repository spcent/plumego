# Card 0712

Priority: P2

Goal:
- Make `WrapError`'s context-precedence rule explicit in code and docs, or
  change it so that callers can build a visible call stack.

Problem:

`error_wrap.go:57-96` — `WrapError` when re-wrapping a `*WrappedErrorWithContext`:

```go
op := wrapped.Context.Operation
if op == "" {
    op = operation     // new operation only used if existing is empty
}
mod := wrapped.Context.Module
if mod == "" {
    mod = module       // new module only used if existing is empty
}
```

When a `*WrappedErrorWithContext` that already has `Operation = "db.query"` is
re-wrapped with `operation = "service.create"`, the outer operation is silently
discarded. The result contains only the original `"db.query"`.

This makes call-stack reconstruction impossible. A function that catches an error
from a dependency and re-wraps it to add its own context will silently lose its
own identity in the error chain.

The behavior is the opposite of what most callers expect from "wrap": outer
context should supplement, not be discarded. The current rule also conflicts
with how `wrappedErr.Err` is handled: the inner error IS preserved (line 84
`Err: wrapped.Err`), so the error chain exists — but the operation chain does not.

Two viable fixes:

**Option A (preferred): call-stack slice**
Change `ErrorContext.Operation` to `Operations []string` and `Module` to
`Modules []string`, appending each new context rather than replacing. Callers
can reconstruct the full call path.

**Option B: outer wins**
Change precedence so new (outer) operation/module always win. Inner context is
still reachable via the `Unwrap()` chain. This aligns with the intent of a
wrapping call: "I'm adding my context to this error."

**Option C: document as-is**
If inner-wins is intentional (e.g., to preserve the root cause location), add
a prominent doc comment to `WrapError` explaining the rule and why. Do not
change behavior.

A decision on which option to apply requires agreement with callers.

Scope:
- Choose one of A, B, or C.
- If A or B: update `WrapError`, `NewWrappedError`, `WrapErrorf`,
  `GetErrorDetails`, `FormatError`, and all serialization/logging paths.
- If C: add the doc comment only.
- Grep callers of `WrapError` to confirm no caller relies on inner-wins silently:
  `grep -rn 'WrapError' . --include='*.go'`

Non-goals:
- Do not change `SafeExecute` or `SafeExecuteWithResult` signatures.
- Do not change `WrapErrorf` in this card if Option A is chosen (tracked separately).

Files:
- `contract/error_wrap.go`
- (If A or B) `contract/error_utils.go`, `contract/errors.go`, callers

Tests:
- Add a test: re-wrapping an error that already has context should produce
  the expected operation value per the chosen rule.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `WrapError` behavior matches one of the three options above.
- The chosen rule is documented in the function's doc comment.
- Tests explicitly verify the precedence behavior.
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
