# Card 0928

Priority: P2
State: active
Primary Module: contract
Owned Files:
- `contract/error_wrap.go`
Depends On:

Goal:
- Remove or redefine `WrapErrorf` so that it is not a confusing near-twin of `WrapError` with silently different behavior.

Problem:
- `WrapError(err, operation, module, params)` (`error_wrap.go:62`) wraps an error with full structured context: operation, module, and params map. These fields are used by `GetErrorDetails`, `FormatError`, and downstream logging to surface actionable diagnostics.
- `WrapErrorf(err, format, args...)` (`error_wrap.go:107`) also returns `*WrappedErrorWithContext`, but leaves `Context.Operation`, `Context.Module`, and `Context.Params` all empty. Only the `Message` field is set.
- The two functions have the same return type and similar names, but very different contracts. A caller who uses `WrapErrorf` thinking they get the same structured context as `WrapError` is silently losing all diagnostic fields.
- The `WrapErrorf` signature pattern (`format, args...`) matches `fmt.Errorf` but creates a heavier `*WrappedErrorWithContext` value instead of a simple `%w` wrapped error. Callers who only need message wrapping should use `fmt.Errorf("%w: …", err)` directly — there is no reason to prefer `WrapErrorf`.
- The existing comment on `WrapErrorf` (`error_wrap.go:105`) acknowledges the difference but does not prevent misuse.

Scope:
- Delete `WrapErrorf`.
- Search the repository for all call sites of `WrapErrorf` and replace each with either:
  - `WrapError(err, operation, module, params)` when the operation/module are known, or
  - `fmt.Errorf("%w: …", err)` when only a message annotation is needed.
- Update `error_wrap.go` to remove the deleted function and its comment.

Non-goals:
- Do not change `WrapError`.
- Do not change `WrappedErrorWithContext` struct or any other error type.
- Do not introduce a new helper to replace `WrapErrorf`.

Files:
- `contract/error_wrap.go`
- Any file in the repository that calls `WrapErrorf`.

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
- `go build ./...`

Docs Sync:
- None required.

Done Definition:
- `WrapErrorf` is deleted from `contract/error_wrap.go`.
- No call site of `WrapErrorf` remains in the repository.
- All tests pass.

Outcome:
- Pending.
