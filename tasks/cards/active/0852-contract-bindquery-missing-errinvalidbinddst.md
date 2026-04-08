# Card 0852

Priority: P1
State: active
Primary Module: contract
Owned Files:
- `contract/context_bind.go`
Depends On:

Goal:
- Make both invalid-destination paths in `bindQuery` wrap `ErrInvalidBindDst`, matching the behavior of `decodeJSONBody`.

Problem:
- `bindQuery` (`context_bind.go:86-93`) has two guard clauses for invalid destinations:

  **Case 1** — non-pointer or nil pointer (line 87-89): wraps `ErrInvalidBindDst` correctly:
  ```go
  if rv.Kind() != reflect.Ptr || rv.IsNil() {
      return &bindError{..., Err: ErrInvalidBindDst}
  }
  ```

  **Case 2** — pointer to non-struct (line 91-93): missing `Err` entirely:
  ```go
  if rv.Kind() != reflect.Struct {
      return &bindError{Status: http.StatusBadRequest, Message: "bind destination must be a pointer to a struct"}
  }
  ```

- Both cases represent the same category of caller error: the destination type is wrong. Yet `errors.Is(err, ErrInvalidBindDst)` returns `true` for Case 1 and `false` for Case 2.
- `BindErrorToAPIError` (`bind_helpers.go:66`) has a `case errors.Is(err, ErrInvalidBindDst)` branch that correctly maps to `CodeInvalidBindDst`. Case 2 falls through to the generic fallback, returning `CodeRequestBindError` instead.
- `decodeJSONBody` (`context_bind.go:54`) consistently wraps `ErrInvalidBindDst` for its nil-destination check, making the inconsistency between `BindJSON` and `BindQuery` visible to callers who inspect errors.

Scope:
- In `bindQuery` (line 92), add `Err: ErrInvalidBindDst` to the Case 2 `bindError`:
  ```go
  return &bindError{
      Status:  http.StatusBadRequest,
      Message: "bind destination must be a pointer to a struct",
      Err:     ErrInvalidBindDst,
  }
  ```
- Add a test: `BindQuery` with a `*int` (pointer to non-struct) must satisfy `errors.Is(err, ErrInvalidBindDst)`.

Non-goals:
- Do not change `BindJSON`, `decodeJSONBody`, or `BindErrorToAPIError`.
- Do not change the error message text.

Files:
- `contract/context_bind.go`
- `contract/write_bind_error_test.go` (add test case)

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None required.

Done Definition:
- Both invalid-destination cases in `bindQuery` wrap `ErrInvalidBindDst`.
- `errors.Is(err, ErrInvalidBindDst)` is true for a pointer-to-non-struct destination.
- A test covers this path.
- All existing tests pass.

Outcome:
- Pending.
