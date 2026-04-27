# Card 0186

Milestone: contract cleanup
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/context_core.go`
- `contract/context_bind.go`
- `contract/bind_helpers.go`
Depends On: —

Goal:
- Eliminate the contradiction between `BindError.Status` being set by every
  constructor and being ignored by the canonical conversion function
  `BindErrorToAPIError`.

Problem:
Every `BindError` construction in `context_bind.go` sets the `Status` field:

```go
&BindError{Status: http.StatusRequestEntityTooLarge, Message: ..., Err: ...}
&BindError{Status: http.StatusBadRequest, Message: ..., Err: ...}
```

Tests (`context_extended_test.go:258`, `context_test.go:120`) check `bindErr.Status`
directly and rely on it being correct.

However `BindErrorToAPIError` (the canonical path that converts a `BindError` to
the outgoing API response) entirely ignores `BindError.Status` and re-derives the
HTTP status from sentinel error matching:

```go
status := http.StatusBadRequest            // default
case errors.Is(err, ErrRequestBodyTooLarge):
    status = http.StatusRequestEntityTooLarge  // re-derived, not from .Status
```

This creates two independent sources of the same truth that can silently diverge:
if a future caller creates a `BindError` with a custom `Status` other than 400 or
413, `BindErrorToAPIError` will override it with its own re-derived value without
any warning.

Fix:
Use `BindError.Status` as the authoritative source inside `BindErrorToAPIError`
when a `*BindError` is found in the error chain.

Specifically, at the end of the switch/default block, after deriving `status` from
sentinel matching, check:
```go
var bindErr *BindError
if errors.As(err, &bindErr) && bindErr != nil && bindErr.Status != 0 {
    status = bindErr.Status
}
```

This ensures the Status that was explicitly set by the constructor is honoured,
makes the two sources consistent, and requires no change to callers or the
`BindError` struct.

Scope:
- Update `BindErrorToAPIError` in `contract/bind_helpers.go` to prefer
  `BindError.Status` when present.
- Ensure the existing sentinel-based switch still correctly sets `code` and
  `message`; only `status` is affected.
- Update or add a test asserting that a `BindError` with a non-default `Status`
  produces an `APIError` with that same status.

Non-goals:
- No change to `BindError` struct fields.
- No change to constructors in `context_bind.go`.
- No change to `BindError.Error()` or `Unwrap()`.

Files:
- `contract/bind_helpers.go`
- `contract/bind_helpers_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `BindErrorToAPIError` uses `BindError.Status` when it is non-zero.
- A test verifies that a `BindError` with `Status: 422` produces an `APIError`
  with `Status: 422`.
- All existing tests pass.

Outcome:
