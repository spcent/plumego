# Card 0911

Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/errors.go`
- `contract/context_core.go`
Depends On: —

Goal:
- Fix three related API surface issues in the `contract` package: a missing nil guard in
  `WriteError`, a silent aliasing footgun in `Ctx.Headers`, and an unnecessarily exported
  `BindError` type that leaks bind internals.

---

## Issue A — `WriteError` has no nil guard for `w`, unlike `WriteJSON`

**File:** `contract/errors.go` lines 164–207 vs `contract/response.go` line 18–20

`WriteJSON` guards against a nil writer at the top of the function:
```go
func WriteJSON(w http.ResponseWriter, status int, payload any) error {
    if w == nil {
        return ErrResponseWriterNil
    }
    ...
}
```

`WriteError` has no equivalent guard. If `w` is nil, `w.Header().Set(...)` on line 203
panics. `WriteResponse` is safe because it delegates to `WriteJSON`, but `WriteError`
writes headers directly.

**Fix:** Add `if w == nil { return ErrResponseWriterNil }` at the top of `WriteError`,
before `validateAPIError`.

---

## Issue B — `Ctx.Headers` is a silent alias to `r.Header`

**File:** `contract/context_core.go` line 246

```go
ctx := &Ctx{
    ...
    Headers: r.Header,  // direct alias — not a copy
    ...
}
```

`http.Header` is a `map[string][]string`. Assigning it creates an alias, not a copy.
Any write through `ctx.Headers` — `ctx.Headers.Set(...)`, `ctx.Headers.Del(...)` —
mutates the underlying request headers directly and is visible to all subsequent
middleware and code that reads `r.Header`.

This is surprising: a developer who writes
```go
ctx.Headers.Set("X-Internal-Tag", "yes")
```
expects to annotate the context, but instead modifies the original request. The field has
no doc comment warning about this behavior.

**Fix:** Replace the exported `Headers http.Header` field with a read-only accessor
method that makes the aliasing explicit:

```go
// RequestHeaders returns the request's header map.
// Writes to the returned map modify the underlying http.Request headers directly.
func (c *Ctx) RequestHeaders() http.Header {
    if c == nil || c.R == nil {
        return nil
    }
    return c.R.Header
}
```

- Remove the `Headers` field from the `Ctx` struct.
- Grep for callers: `grep -rn '\.Headers\b' . --include='*.go'`
- Migrate all read callers from `ctx.Headers.Get(...)` to `ctx.RequestHeaders().Get(...)`.
- Any write callers must explicitly use `ctx.R.Header.Set(...)` so the mutation intent
  is visible at the call site.

---

## Issue C — `BindError` is exported but has no callers outside the bind layer

**File:** `contract/context_core.go` lines 79–84

```go
// BindError represents an error that occurred while binding a request body.
type BindError struct {
    Status  int
    Message string
    Err     error
}
```

`BindError` is exported, but inspection of the codebase shows:
- All external code handles binding errors via `WriteBindError`, `FieldErrorsFrom`, or `BindErrorToAPIError` — none of which require callers to know about `BindError`.
- No external file performs `errors.As(err, &contract.BindError{})`.
- The type is constructed only inside `context_bind.go` and `bind_helpers.go`.

Exporting `BindError` signals to callers that they should interact with this type
directly, creating an undocumented second path alongside the intended API
(`WriteBindError`, `FieldErrorsFrom`).

**Fix:**
- Rename `BindError` → `bindError` (unexported).
- Verify `grep -rn 'contract\.BindError\|\*BindError\|BindError{' . --include='*.go'`
  returns only `context_bind.go`, `bind_helpers.go`, and test files.
- Update `context_bind.go` and `bind_helpers.go` to use `bindError` internally.
- Test files that reference `contract.BindError` must be updated to test through the
  public API (`FieldErrorsFrom`, `WriteBindError`) instead.

---

Scope:
- Apply all three fixes; they are independent and can land in one pass.
- No behavior changes — only nil-safety, documentation, and visibility corrections.

Non-goals:
- Do not change `BindErrorToAPIError`, `WriteBindError`, or `FieldErrorsFrom`.
- Do not change `APIError` or `ErrorResponse`.
- Do not add new response methods.

Files:
- `contract/errors.go`
- `contract/context_core.go`
- `contract/context_bind.go`
- `contract/bind_helpers.go`
- Caller files found by grep for `Headers` and `BindError`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `WriteError` returns `ErrResponseWriterNil` when `w == nil`.
- `Ctx.Headers` field is removed; `RequestHeaders()` method exists in its place.
- `bindError` is unexported; `grep -rn 'contract\.BindError' . --include='*.go'` returns empty.
- `go build ./...` and all tests pass.

Outcome:
- A) `WriteError` now returns `ErrResponseWriterNil` when `w == nil`, matching `WriteJSON`.
- B) `Ctx.Headers http.Header` field removed; `RequestHeaders() http.Header` method added. All callers in `x/webhook` migrated to `ctx.RequestHeaders().Get(...)`.
- C) `BindError` renamed to `bindError` (unexported); updated in `context_bind.go` and `bind_helpers.go`. External tests updated to test through public API.
- `grep -rn 'contract\.BindError' . --include='*.go'` returns empty.
- `go test -timeout 20s ./...` passes.
