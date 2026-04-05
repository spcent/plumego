# Card 0768

Milestone: contract cleanup
Priority: P1
State: active
Primary Module: contract
Owned Files:
- `contract/context_response.go`
Depends On: —

Goal:
- Add nil-receiver guards to `Ctx` methods that are missing them, consistent with
  the nil guard present on every other `Ctx` response method.

Problem:
Four `Ctx` response methods in `context_response.go` have no nil-receiver guard,
while every other response method (`Response`, `Text`, `Bytes`, `Redirect`) starts
with `if c == nil { return ErrContextNil }`:

| Method | Returns error | Nil guard |
|---|---|---|
| `Response` | yes | ✓ |
| `Text` | yes | ✓ |
| `Bytes` | yes | ✓ |
| `Redirect` | yes | ✓ |
| `SafeRedirect` | yes | ✗ — calls `c.R` immediately |
| `File` | yes | ✗ — calls `c.W` and `c.R` |
| `Cookie` | yes | ✗ — calls `c.R.Cookie` |
| `SetCookie` | — | ✗ — calls `c.W` |

`SafeRedirect`, `File`, and `Cookie` will panic with a nil pointer dereference
if called on a nil `*Ctx`. `SetCookie` has no return value so it cannot return
`ErrContextNil`, but it should at minimum have a nil guard that is a no-op
(consistent with how nil-safe value types work in Go).

Scope:
- Add `if c == nil { return ErrContextNil }` to `SafeRedirect`, `File`, `Cookie`.
- Add `if c == nil { return }` to `SetCookie` (no error return, no-op is correct).
- Keep all existing logic unchanged.

Non-goals:
- No nil guards for non-response methods (`Set`, `Get`, `Param`, `Abort`, etc.).
  Those have their own existing patterns or are out of scope here.
- No change to method signatures.

Files:
- `contract/context_response.go`

Tests:
- Add a test case for each affected method called on a nil `*Ctx`:
  - `SafeRedirect` → must return `ErrContextNil`
  - `File` → must return `ErrContextNil`
  - `Cookie` → must return `("", ErrContextNil)`
  - `SetCookie` → must not panic
- `go test -timeout 20s ./contract/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `SafeRedirect`, `File`, `Cookie` return `ErrContextNil` on nil receiver.
- `SetCookie` returns without panicking on nil receiver.
- No existing tests regress.

Outcome:
