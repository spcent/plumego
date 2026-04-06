# Card 0812

Milestone: contract cleanup
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/context_response.go`
Depends On: 0803 (nil guard for SetCookie)

Goal:
- Fix `Ctx.SetCookie` so it does not mutate the caller's `*http.Cookie` struct.

Problem:
`SetCookie` modifies the pointer passed by the caller:

```go
func (c *Ctx) SetCookie(cookie *http.Cookie) {
    if cookie.Path == "" {
        cookie.Path = "/"         // ← mutates caller's struct
    }
    cookie.Secure = true          // ← always mutates, even if caller set false
    http.SetCookie(c.W, cookie)
}
```

Two issues:
1. **`cookie.Secure = true` is an unconditional mutation.** If a caller passes a
   cookie with `Secure: false` (e.g., an internal-only cookie or a development
   environment), `SetCookie` silently overrides the choice without any indication.
   After the call, the caller's original struct has `Secure == true`.

2. **Both `Path` and `Secure` are set on the caller's pointer.** Any code that
   re-uses the cookie struct after calling `SetCookie` (or inspects it) will see
   the method's overrides, not the original values. This is an unexpected side
   effect on caller-owned data.

Fix: make a shallow copy of `*cookie` at the start of the method, apply
modifications to the copy, and pass the copy to `http.SetCookie`:

```go
func (c *Ctx) SetCookie(cookie *http.Cookie) {
    if c == nil {        // nil guard (card 0803)
        return
    }
    local := *cookie    // copy; do not mutate caller's struct
    if local.Path == "" {
        local.Path = "/"
    }
    local.Secure = true
    http.SetCookie(c.W, &local)
}
```

Scope:
- Apply the shallow-copy fix in `context_response.go`.
- Add a test: after calling `SetCookie`, the original `*http.Cookie` struct passed
  by the caller must have its `Secure` field unchanged from what was set before
  the call.

Non-goals:
- No change to the forced-`Secure` policy (that is a deliberate security choice).
- No change to the forced default `Path = "/"`.

Files:
- `contract/context_response.go`
- `contract/context_extended_test.go`

Tests:
- `go test -timeout 20s ./contract/...`

Docs Sync: —

Done Definition:
- `SetCookie` does not mutate the caller's `*http.Cookie`.
- A test asserts that the original cookie's fields are unmodified after the call.
- All existing tests pass.

Outcome:
