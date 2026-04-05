# Card 0811

Milestone: contract cleanup
Priority: P2
State: active
Primary Module: contract
Owned Files:
- `contract/error_wrap.go`
Depends On: —

Goal:
- Unexport `NewWrappedError` and `ErrorContext` — both are lower-level building
  blocks with no external callers; exposing them alongside `WrapError` creates
  two ways to construct the same type with different semantics.

Problem:
The package exports two constructors for `WrappedErrorWithContext`:

| Function | Behaviour |
|---|---|
| `WrapError(err, op, mod, params)` | public API; preserves inner context if err is already `*WrappedErrorWithContext` |
| `NewWrappedError(err, op, mod, params)` | always creates a fresh wrapper; ignores any existing inner context |

`NewWrappedError` is only called from one place in production code — the last
line of `WrapError` itself. External callers who use `NewWrappedError` directly
bypass the inner-context precedence rule (card 0749, regression test
`TestWrapErrorPreservesInnerContextPrecedence`).

`ErrorContext` is the struct embedded in `WrappedErrorWithContext.Context`. It is
only referenced inside `error_wrap.go`. External callers never construct or type-
assert an `ErrorContext` — they only read its fields via `wrapped.Context.Operation`
etc., which works with an unexported type via an exported field. Exporting it
invites callers to construct `ErrorContext` literals directly, which again bypasses
`WrapError`.

External callers of both: **none** (verified by grep).

Scope:
- Rename `NewWrappedError` → `newWrappedError`.
- Rename `ErrorContext` → `errorContext`.
- Update all internal references in `contract/error_wrap.go`.
- Update `contract/error_handling_test.go` if it references either symbol
  (currently it does not).

Non-goals:
- No logic changes.
- No change to `WrapError`, `WrapErrorf`, or `WrappedErrorWithContext` visibility.
- The exported `.Context` field on `WrappedErrorWithContext` remains; it just
  carries an unexported struct type (callers can still read `.Context.Operation`).

Files:
- `contract/error_wrap.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `grep -rn '\bNewWrappedError\b\|\bErrorContext\b' . --include='*.go'` must
  return empty after the change.

Docs Sync: —

Done Definition:
- `NewWrappedError` and `ErrorContext` are not exported.
- `WrapError` and `WrappedErrorWithContext` remain exported and unchanged.
- All tests pass.

Outcome:
